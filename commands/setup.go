package commands

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/internal/factory"
	"github.com/RobBrazier/readflow/internal/form"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/cli/browser"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SetupCommand struct {
	ctx     context.Context
	cfg     *config.Config
	cmd     *cli.Command
	targets factory.TargetFactory
}

type tokenLink struct {
	Name        string
	source      *string
	ShouldFetch bool
	Value       string
}

func (t *tokenLink) Save() {
	if t.Value != "" {
		// remove bearer from token
		*t.source = strings.TrimSpace(strings.Replace(t.Value, "Bearer", "", 1))
	}
}

func newTokenLink(name string, source *string) *tokenLink {
	shouldFetch := *source == ""
	return &tokenLink{
		Name:        name,
		source:      source,
		ShouldFetch: shouldFetch,
		Value:       "",
	}
}

func (c SetupCommand) Run() error {
	tokens := []*tokenLink{
		newTokenLink("anilist", &c.cfg.Tokens.Anilist),
		newTokenLink("hardcover", &c.cfg.Tokens.Hardcover),
	}

	syncDays := strconv.Itoa(c.cfg.SyncDays)

	form := c.buildForm(tokens, &syncDays)

	err := form.Run()
	if err != nil {
		return err
	}
	syncDaysInt, err := strconv.Atoi(syncDays)
	if err != nil {
		log.Warn("Unable to parse sync days from setup", "value", syncDays)
	} else {
		c.cfg.SyncDays = syncDaysInt
	}

	for _, token := range tokens {
		token.Save()
	}

	err = config.SaveConfig(c.cfg)
	if err != nil {
		return err
	}
	log.Info("Successfully saved config!")
	return nil
}

func (c SetupCommand) buildForm(tokens []*tokenLink, syncDays *string) *huh.Form {
	// Initial config form
	sourceTargetGroup := huh.NewGroup(
		form.SourceSelect(c.ctx, &c.cfg.Source),
		form.TargetSelect(c.ctx, &c.cfg.Targets),
		form.SyncDays(syncDays),
	)

	databaseGroup := huh.NewGroup(
		huh.NewInput().
			Title("Calibre datatabase location").
			Description("e.g. /path/to/metadata.db").
			Validate(form.ValidationRequired[string]()).
			Value(&c.cfg.Databases.Calibre),
		huh.NewInput().
			Title("Calibre-Web datatabase location").
			Description("e.g. /path/to/app.db").
			Validate(form.ValidationRequired[string]()).
			Value(&c.cfg.Databases.CalibreWeb),
	).WithHideFunc(func() bool {
		return c.cfg.Source != "database"
	})

	chaptersColumnGroup := huh.NewGroup(
		huh.NewInput().
			Title("Calibre database 'chapters' custom column").
			Description("e.g. custom_column_15 (if not specified, it'll be searched for in the calibre db)\nNOTE: only used for anilist").
			Value(&c.cfg.Columns.Chapter),
	).WithHideFunc(func() bool {
		return !slices.Contains(c.cfg.Targets, "anilist") || c.cfg.Source != "database"
	})
	groups := []*huh.Group{
		sourceTargetGroup,
		databaseGroup,
		chaptersColumnGroup,
	}
	groups = append(groups, c.createReauthGroups(tokens)...)

	form := huh.NewForm(groups...)
	return form
}

func (c SetupCommand) createReauthGroups(tokens []*tokenLink) []*huh.Group {
	var (
		confirmGroups []*huh.Group
		authGroups    []*huh.Group
		groups        []*huh.Group
	)
	tokenMap := make(map[string]*tokenLink)
	for _, token := range tokens {
		tokenMap[token.Name] = token
	}
	for _, name := range c.targets.GetAvailable() {
		title := cases.Title(language.English).String(name)
		target, _ := c.targets.GetTarget(name)

		token := tokenMap[name]
		shouldFetch := token.ShouldFetch

		confirmGroup := huh.NewGroup(
			form.Confirm(fmt.Sprintf("Token already exists in config for %s, Re-authenticate", title), &shouldFetch),
		).WithHideFunc(func() bool {
			return !slices.Contains(c.cfg.Targets, name) || shouldFetch
		})
		confirmGroups = append(confirmGroups, confirmGroup)

		authGroup := huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Authenticating with %s - Please paste the token below", title)).
				DescriptionFunc(func() string {
					url, _ := target.Login()
					err := browser.OpenURL(url)
					message := "Please open the following URL in your browser if it hasn't already opened: %s"
					if err != nil {
						message = "Please copy the following URL and open in your browser (couldn't open browser): %s"
					}
					return fmt.Sprintf(message, url)
				}, nil).
				EchoMode(huh.EchoMode(textinput.EchoPassword)).
				Value(&token.Value),
		).WithHideFunc(func() bool {
			return !slices.Contains(c.cfg.Targets, name) || !shouldFetch
		})
		authGroups = append(authGroups, authGroup)
	}

	groups = append(groups, confirmGroups...)
	groups = append(groups, authGroups...)

	return groups
}

func NewSetupCommand(ctx context.Context, cmd *cli.Command) error {
	cfg := config.GetFromContext(ctx)
	command := SetupCommand{
		ctx:     ctx,
		cmd:     cmd,
		cfg:     cfg,
		targets: factory.NewTargetFactory(ctx),
	}
	return command.Run()
}
