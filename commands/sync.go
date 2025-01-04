package commands

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

type SyncCommand struct {
	ctx     context.Context
	cfg     *config.Config
	source  string
	targets []string
	cmd     *cli.Command
	dryrun  bool
}

func (c SyncCommand) validate() error {
	var errs []error
	sources := source.GetSources()
	allowedSources := slices.Collect(maps.Keys(sources))
	if _, ok := sources[c.source]; !ok {
		errs = append(errs, errors.New(fmt.Sprintf("Invalid source [%s] provided in configuration. Allowed values: %v", c.source, allowedSources)))
	}
	targets := target.GetTargets()
	allowedTargets := slices.Collect(maps.Keys(targets))
	for _, t := range c.targets {
		if _, ok := targets[t]; !ok {
			errs = append(errs, errors.New(fmt.Sprintf("Invalid target [%s] provided in configuration. Allowed values: %v", t, allowedTargets)))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c SyncCommand) Run() error {
	log.Debug("sync called", "source", c.source, "targets", c.targets)

	// run validation to ensure that valid source / targets are provided
	err := c.validate()
	if err != nil {
		return err
	}

	syncSource := source.GetActiveSource(c.source, c.ctx)
	syncTargets := target.GetActiveTargets(c.targets, c.ctx)

	action := sync.NewSyncAction(*syncSource, syncTargets)
	results, err := action.Sync()
	if err != nil {
		return err
	}
	log.Info("sync completed", "results", results)
	return nil
}

func NewSyncCommand(ctx context.Context, cmd *cli.Command) error {
	dryrun := cmd.Bool("dryrun")
	cfg := config.GetFromContext(ctx)

	source := cfg.Source
	targets := cfg.Targets

	command := SyncCommand{
		ctx:     ctx,
		cmd:     cmd,
		source:  source,
		targets: targets,
		cfg:     cfg,
		dryrun:  dryrun,
	}
	return command.Run()
}
