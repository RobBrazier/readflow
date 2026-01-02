package main

import (
	"context"
	"os"

	"github.com/RobBrazier/readflow/commands"
	"github.com/adrg/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	configFile, _ := xdg.ConfigFile("readflow/config.yaml")

	cmd := cli.Command{
		Name:        "readflow",
		Usage:       "Track your Kobo reads from Calibre-Web and Calibre databases",
		Description: "Sync reading status to Anilist.co and Hardcover.app",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"debug"},
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       configFile,
				Destination: &configFile,
				TakesFile:   true,
				Action: func(ctx context.Context, c *cli.Command, s string) error {
					log.Debug().Str("file", s).Msg("Using config file")
					return nil
				},
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			level := zerolog.InfoLevel
			if c.Bool("verbose") {
				level = zerolog.DebugLevel
			}
			zerolog.SetGlobalLevel(level)
			return ctx, nil
		},
		Commands: []*cli.Command{
			commands.SyncCommand(&configFile),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err)
	}
}
