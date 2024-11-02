package cmd

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/form"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/cli/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup configuration options and login to services",
	RunE: func(cmd *cobra.Command, args []string) error {
		source := viper.GetString(internal.CONFIG_SOURCE)
		targets := viper.GetStringSlice(internal.CONFIG_TARGETS)
		calibreDb := viper.GetString(internal.CONFIG_CALIBRE_DB)
		calibreWebDb := viper.GetString(internal.CONFIG_CALIBREWEB_DB)
		chaptersColumn := viper.GetString(internal.CONFIG_CHAPTERS_COLUMN)

		anilistTokenExists := viper.GetString(internal.CONFIG_TOKENS_ANILIST) != ""
		hardcoverTokenExists := viper.GetString(internal.CONFIG_TOKENS_HARDCOVER) != ""

		fetchAnilistToken := !anilistTokenExists
		fetchHardcoverToken := !hardcoverTokenExists

		var anilistToken string
		var hardcoverToken string

		// Initial config form
		form := huh.NewForm(
			// Source/Target
			huh.NewGroup(
				form.SourceSelect(&source),
				form.TargetSelect(&targets),
			),
			// Databases
			huh.NewGroup(
				huh.NewInput().
					Title("Calibre datatabase location").
					Description("e.g. /path/to/metadata.db").
					Value(&calibreDb),
				huh.NewInput().
					Title("Calibre-Web datatabase location").
					Description("e.g. /path/to/app.db").
					Value(&calibreWebDb),
			).WithHideFunc(func() bool {
				return source != "database"
			}),
			// Chapters Column
			huh.NewGroup(
				huh.NewInput().
					Title("Calibre database 'chapters' custom column").
					Description("e.g. custom_column_15 (if not specified, it'll be searched for in the calibre db)\nNOTE: only used for anilist").
					Value(&chaptersColumn),
			).WithHideFunc(func() bool {
				return !slices.Contains(targets, "anilist") || source != "database"
			}),
			// Prompt for Re-Authentication
			huh.NewGroup(
				form.Confirm("Token already exists in config for Anilist, Re-authenticate", &fetchAnilistToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(targets, "anilist") || fetchAnilistToken
			}),
			huh.NewGroup(
				form.Confirm("Token already exists in config for Hardcover, Re-authenticate", &fetchHardcoverToken),
			).WithHideFunc(func() bool {
				return !slices.Contains(targets, "hardcover") || fetchHardcoverToken
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
				return !slices.Contains(targets, "anilist") || !fetchAnilistToken
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
				return !slices.Contains(targets, "hardcover") || !fetchHardcoverToken
			}),
		)
		err := form.Run()
		if err != nil {
			return err
		}
		viper.Set(internal.CONFIG_SOURCE, source)
		viper.Set(internal.CONFIG_TARGETS, targets)
		viper.Set(internal.CONFIG_CALIBRE_DB, calibreDb)
		viper.Set(internal.CONFIG_CALIBREWEB_DB, calibreWebDb)
		viper.Set(internal.CONFIG_CHAPTERS_COLUMN, chaptersColumn)
		if fetchAnilistToken && anilistToken != "" {
			viper.Set(internal.CONFIG_TOKENS_ANILIST, anilistToken)
		}
		if fetchHardcoverToken && hardcoverToken != "" {
			viper.Set(internal.CONFIG_TOKENS_HARDCOVER, hardcoverToken)
		}
		err = viper.WriteConfig()
		if err != nil {
			return err
		}
		slog.Info("Saved settings to config")
		return nil
	},
}

func getTarget(name string) target.SyncTarget {
	for _, target := range target.GetTargets() {
		if target.GetName() == name {
			return target
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
