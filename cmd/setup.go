package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/form"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/cli/browser"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup configuration options and login to services",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := config.GetConfig()

		anilistTokenExists := c.Tokens.Anilist != ""
		hardcoverTokenExists := c.Tokens.Hardcover != ""

		fetchAnilistToken := !anilistTokenExists
		fetchHardcoverToken := !hardcoverTokenExists

		var anilistToken string
		var hardcoverToken string

		// Initial config form
		form := huh.NewForm(
			// Source/Target
			huh.NewGroup(
				form.SourceSelect(&c.Source),
				form.TargetSelect(&c.Targets),
			),
			// Databases
			huh.NewGroup(
				huh.NewInput().
					Title("Calibre datatabase location").
					Description("e.g. /path/to/metadata.db").
					Validate(form.ValidationRequired[string]()).
					Value(&c.Databases.Calibre),
				huh.NewInput().
					Title("Calibre-Web datatabase location").
					Description("e.g. /path/to/app.db").
					Validate(form.ValidationRequired[string]()).
					Value(&c.Databases.CalibreWeb),
			).WithHideFunc(func() bool {
				return c.Source != "database"
			}),
			// Chapters Column
			huh.NewGroup(
				huh.NewInput().
					Title("Calibre database 'chapters' custom column").
					Description("e.g. custom_column_15 (if not specified, it'll be searched for in the calibre db)\nNOTE: only used for anilist").
					Value(&c.Columns.Chapter),
			).WithHideFunc(func() bool {
				return !slices.Contains(c.Targets, "anilist") || c.Source != "database"
			}),
			// Prompt for Re-Authentication
			huh.NewGroup(
				form.Confirm("Token already exists in config for Anilist, Re-authenticate", &fetchAnilistToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(c.Targets, "anilist") || fetchAnilistToken
			}),
			huh.NewGroup(
				form.Confirm("Token already exists in config for Hardcover, Re-authenticate", &fetchHardcoverToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(c.Targets, "hardcover") || fetchHardcoverToken
			}),
			// Re-Authorize if requested
			huh.NewGroup(
				huh.NewInput().
					Title("Authenticating with Anilist").
					DescriptionFunc(func() string {
						url, _ := getTarget("anilist").Login()
						browser.OpenURL(url)
						return fmt.Sprintf(
							"Please open the following URL in your browser if it hasn't already opened:\n%s",
							url,
						)
					}, nil).
					EchoMode(huh.EchoMode(textinput.EchoPassword)).
					Value(&anilistToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(c.Targets, "anilist") || !fetchAnilistToken
			}),
			huh.NewGroup(
				huh.NewInput().
					Title("Authenticating with Hardcover").
					DescriptionFunc(func() string {
						url, _ := getTarget("hardcover").Login()
						browser.OpenURL(url)
						return fmt.Sprintf(
							"Please open the following URL in your browser if it hasn't already opened:\n%s",
							url,
						)
					}, nil).
					EchoMode(huh.EchoMode(textinput.EchoPassword)).
					Value(&hardcoverToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(c.Targets, "hardcover") || !fetchHardcoverToken
			}),
		)
		err := form.Run()
		if err != nil {
			return err
		}

		if fetchAnilistToken && anilistToken != "" {
			c.Tokens.Anilist = anilistToken
		}
		if fetchHardcoverToken && hardcoverToken != "" {
			c.Tokens.Hardcover = strings.TrimSpace(strings.Replace(hardcoverToken, "Bearer", "", 1))
		}

		err = config.SaveConfig(&c)
		if err != nil {
			return err
		}
		log.Info("Successfully saved config!")
		return nil
	},
}

func getTarget(name string) target.SyncTarget {
	for _, target := range *target.GetTargets() {
		if target.GetName() == name {
			return target
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
