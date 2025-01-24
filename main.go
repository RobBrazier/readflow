package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/RobBrazier/readflow/commands"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func main() {
	log.SetTimeFormat(time.TimeOnly)
	log.SetLevel(log.InfoLevel)

	app := &cli.Command{
		Name:    "readflow",
		Usage:   "Track your Kobo reads on Anilist.co and Hardcover.app using Calibre-Web and Calibre databases",
		Suggest: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
				Action: func(_ context.Context, _ *cli.Command, verbose bool) error {
					if verbose {
						log.SetLevel(log.DebugLevel)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "sync",
				Usage:  "Sync latest reading states to configured targets",
				Before: loadConfig,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "dryrun",
						Aliases: []string{"d"},
						Usage:   "Enable dryrun mode (check what would change but don't sync anything)",
					},
				},
				Action: commands.NewSyncCommand,
			},
			{
				Name:   "setup",
				Usage:  "Setup configuration options and login to services",
				Before: loadConfig,
				Action: commands.NewSetupCommand,
			},
			{
				Name:   "schedule",
				Usage:  "Run sync on a specified schedule",
				Before: loadConfig,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "schedule",
						Usage:       "cron-style schedule, e.g. @hourly or 0 1 0 0 0",
						DefaultText: "@hourly",
					},
				},
				Action: commands.NewScheduleCommand,
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(ctx context.Context, c *cli.Command) (context.Context, error) {
	override := c.String("config")
	configFile := config.GetConfigPath(override)
	cfg, err := config.LoadConfig(configFile)

	if os.Getenv("READFLOW_DOCKER") == "1" {
		err = config.LoadConfigFromEnv(cfg)
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warnf("Config file doesn't seem to exist! Please run `%s setup -c \"%s\"` to populate the configuration", c.Root().Name, configFile)
		} else {
			log.Error("Unable to read config")
		}
		return ctx, err
	}
	ctx = config.AddToContext(ctx, cfg)
	return ctx, err
}
