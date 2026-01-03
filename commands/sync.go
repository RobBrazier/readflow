package commands

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var (
	ALLOWED_SOURCES = []string{"database"}
	ALLOWED_TARGETS = []string{"anilist", "hardcover"}
)

func validateAllowedValues(flag string, in, allowed []string) error {
	logger := log.Fatal().Strs("allowed", allowed)
	if len(in) == 0 {
		logger.Msgf("At least one %s must be provided", flag)
		return fmt.Errorf("at least one value must be provided. Allowed values: %s", allowed)
	}
	for _, val := range in {
		if slices.Contains(allowed, val) {
			continue
		}
		logger.Str("invalid", val).Msgf("Invalid %s", flag)
	}
	return nil
}

func SyncCommand() *cli.Command {
	return &cli.Command{
		Name: "sync",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "source",
				Aliases: []string{"s"},
				Value:   []string{"database"},
				Usage:   fmt.Sprintf("Allowed values: %s", ALLOWED_SOURCES),
				Action: func(ctx context.Context, c *cli.Command, s []string) error {
					return config.Set("source.enabled", s)
				},
			},
			&cli.StringSliceFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Value:   []string{"hardcover"},
				Usage:   fmt.Sprintf("Allowed values: %s", ALLOWED_TARGETS),
				Action: func(ctx context.Context, c *cli.Command, s []string) error {
					return config.Set("target.enabled", s)
				},
			},
			&cli.IntFlag{
				Name:    "days",
				Aliases: []string{"sync-days", "d"},
				Value:   7,
				Action: func(ctx context.Context, c *cli.Command, i int) error {
					return config.Set("sync.days", i)
				},
			},
			&cli.StringFlag{
				Name:    "schedule",
				Aliases: []string{"cron"},
				Action: func(ctx context.Context, c *cli.Command, s string) error {
					return config.Set("sync.schedule", s)
				},
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			days := config.Get().Int("sync.days")
			sources := config.Get().Strings("source.enabled")
			if err := validateAllowedValues("source", sources, ALLOWED_SOURCES); err != nil {
				return err
			}
			targets := config.Get().Strings("target.enabled")
			if err := validateAllowedValues("target", targets, ALLOWED_TARGETS); err != nil {
				return err
			}
			log.Info().
				Strs("sources", sources).
				Strs("targets", targets).
				Int("days", days).
				Msg("running with config")

			availableSources := source.GetSources(days)
			availableTargets := target.GetTargets(ctx)

			var enabledSources []source.Source
			var enabledTargets []target.SyncTarget

			for key, val := range availableSources {
				if slices.Contains(sources, key) {
					enabledSources = append(enabledSources, val)
				}
			}

			for key, val := range availableTargets {
				if slices.Contains(targets, key) {
					enabledTargets = append(enabledTargets, val)
				}
			}

			action := sync.NewSyncAction(enabledSources, enabledTargets)
			_, err := action.Sync()
			if err != nil {
				return err
			}
			if schedule := config.Get().String("sync.schedule"); schedule != "" {
				c := cron.New()
				log.Info().Str("schedule", schedule).Msg("Starting scheduled sync")
				id, err := c.AddFunc(schedule, func() { action.Sync() })
				entry := c.Entry(id)
				log.Info().Msgf("Next execution: %s", entry.Next)
				if err != nil {
					log.Error().Err(err).Msg("Unable to schedule cron")
					return err
				}
				c.Start()
				for len(c.Entries()) > 0 {
					time.Tick(time.Minute)
				}
			}

			return nil
		},
	}
}
