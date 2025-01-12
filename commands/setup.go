package commands

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/internal/form"
	"github.com/RobBrazier/readflow/target"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/cli/browser"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SetupCommand struct {
	ctx context.Context
	cfg *config.Config
	cmd *cli.Command
}

func (c SetupCommand) getTarget(name string) target.SyncTarget {
	for targetName, target := range target.GetTargets() {
		if targetName == name {
			return target(c.ctx)
		}
	}
	return nil
}

func (c SetupCommand) Run() error {
	anilistTokenExists := c.cfg.Tokens.Anilist != ""
	hardcoverTokenExists := c.cfg.Tokens.Hardcover != ""

	fetchAnilistToken := !anilistTokenExists
	fetchHardcoverToken := !hardcoverTokenExists
	syncDays := strconv.Itoa(c.cfg.SyncDays)

	fetchToken := map[string]*bool{
		"anilist":   &fetchAnilistToken,
		"hardcover": &fetchHardcoverToken,
	}

	var anilistToken string
	var hardcoverToken string

	values := map[string]*string{
		"anilist":   &anilistToken,
		"hardcover": &hardcoverToken,
	}

	form := c.buildForm(fetchToken, values, &syncDays)

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

	if fetchAnilistToken && anilistToken != "" {
		c.cfg.Tokens.Anilist = anilistToken
	}
	if fetchHardcoverToken && hardcoverToken != "" {
		c.cfg.Tokens.Hardcover = strings.TrimSpace(strings.Replace(hardcoverToken, "Bearer", "", 1))
	}

	err = config.SaveConfig(c.cfg)
	if err != nil {
		return err
	}
	log.Info("Successfully saved config!")
	return nil
}

func (c SetupCommand) buildForm(fetchToken map[string]*bool, values map[string]*string, syncDays *string) *huh.Form {
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
	groups = append(groups, c.createReauthGroups(fetchToken, values)...)

	form := huh.NewForm(groups...)
	return form
}

func (c SetupCommand) createReauthGroups(fetchToken map[string]*bool, values map[string]*string) []*huh.Group {
	var (
		confirmGroups []*huh.Group
		authGroups    []*huh.Group
		groups        []*huh.Group
	)
	targets := target.GetTargets()
	for name, targetFn := range targets {
		title := cases.Title(language.English).String(name)
		target := targetFn(c.ctx)

		shouldFetch := fetchToken[name]
		value := values[name]

		confirmGroup := huh.NewGroup(
			form.Confirm(fmt.Sprintf("Token already exists in config for %s, Re-authenticate", title), shouldFetch),
		).WithHideFunc(func() bool {
			return !slices.Contains(c.cfg.Targets, name) || *shouldFetch
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
				Value(value),
		).WithHideFunc(func() bool {
			return !slices.Contains(c.cfg.Targets, name) || !*shouldFetch
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
		ctx: ctx,
		cmd: cmd,
		cfg: cfg,
	}
	return command.Run()
}
