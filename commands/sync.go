package commands

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/internal/factory"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/source"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

type SyncCommand struct {
	ctx    context.Context
	cfg    *config.Config
	cmd    *cli.Command
	dryrun bool
}

func (c SyncCommand) validate(allowedTargets []string) error {
	var errs []error
	cfgSource := c.cfg.Source
	cfgTargets := c.cfg.Targets
	sources := source.GetSources()
	allowedSources := slices.Collect(maps.Keys(sources))
	if _, ok := sources[cfgSource]; !ok {
		errs = append(errs, errors.New(fmt.Sprintf("Invalid source [%s] provided in configuration. Allowed values: %v", cfgSource, allowedSources)))
	}
	for _, t := range cfgTargets {
		if !slices.Contains(allowedTargets, t) {
			errs = append(errs, errors.New(fmt.Sprintf("Invalid target [%s] provided in configuration. Allowed values: %v", t, allowedTargets)))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c SyncCommand) Run() error {
	cfgSource := c.cfg.Source
	cfgTargets := c.cfg.Targets

	log.Debug("sync called", "source", cfgSource, "targets", cfgTargets)

	targetFactory := factory.NewTargetFactory(c.ctx)
	allowedTargets := targetFactory.GetAvailable()

	// run validation to ensure that valid source / targets are provided
	err := c.validate(allowedTargets)
	if err != nil {
		return err
	}

	syncSource := source.GetActiveSource(cfgSource, c.ctx)
	syncTargets := targetFactory.GetTargets(cfgTargets)

	action := sync.NewSyncAction(*syncSource, syncTargets, c.dryrun)
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

	command := SyncCommand{
		ctx:    ctx,
		cmd:    cmd,
		cfg:    cfg,
		dryrun: dryrun,
	}
	return command.Run()
}
