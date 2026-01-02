package commands

import (
	"context"
	"fmt"
	"slices"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/yaml"
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

func SyncCommand(configFile *string) *cli.Command {
	configSource := altsrc.NewStringPtrSourcer(configFile)
	return &cli.Command{
		Name: "sync",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "source",
				Aliases: []string{"s"},
				Value:   []string{"database"},
				Usage:   fmt.Sprintf("Allowed values: %s", ALLOWED_SOURCES),
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("READFLOW_SOURCE"),
					yaml.YAML("source.enabled", configSource),
					cli.EnvVar("SOURCE"), // legacy
				),
				Validator: func(s []string) error {
					return validateAllowedValues("source", s, ALLOWED_SOURCES)
				},
			},
			&cli.StringSliceFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Value:   []string{"hardcover"},
				Usage:   fmt.Sprintf("Allowed values: %s", ALLOWED_TARGETS),
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("READFLOW_TARGET"),
					yaml.YAML("target.enabled", configSource),
					cli.EnvVar("TARGETS"), // legacy
				),
				Validator: func(s []string) error {
					return validateAllowedValues("target", s, ALLOWED_TARGETS)
				},
			},
			&cli.IntFlag{
				Name:    "days",
				Aliases: []string{"sync-days", "d"},
				Value:   7,
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("READFLOW_SYNC_DAYS"),
					yaml.YAML("sync.days", configSource),
					cli.EnvVar("SYNC_DAYS"), // legacy
				),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			days := c.Int("days")
			sources := c.StringSlice("source")
			targets := c.StringSlice("target")
			cfg, err := config.LoadConfig(*configFile)
			if err != nil {
				log.Error().Err(err).Msg("Unable to load config")
			}
			availableSources := source.GetSources(cfg, days)
			availableTargets := target.GetTargets(ctx, cfg)

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
			_, err = action.Sync()
			if err != nil {
				return err
			}

			return nil
		},
	}
}
