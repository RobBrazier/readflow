package factory

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/target"
	"github.com/RobBrazier/readflow/target/anilist"
	"github.com/RobBrazier/readflow/target/hardcover"
)

type TargetFactory interface {
	GetAvailable() []string
	GetTargets(enabled []string) []target.SyncTarget
	GetTarget(target string) (target.SyncTarget, error)
}

type SyncTargetFunc func(ctx context.Context) target.SyncTarget

func TargetProvider(fn SyncTargetFunc) SyncTargetFunc {
	return func(ctx context.Context) target.SyncTarget {
		return fn(ctx)
	}
}

type targetFactory struct {
	ctx     context.Context
	targets map[string]SyncTargetFunc
}

func (f targetFactory) GetAvailable() []string {
	return slices.Collect(maps.Keys(f.targets))
}

func (f targetFactory) GetTargets(enabled []string) (targets []target.SyncTarget) {
	for name, target := range f.targets {
		if slices.Contains(enabled, name) {
			targets = append(targets, target(f.ctx))
		}
	}
	return targets
}

func (f targetFactory) GetTarget(name string) (target.SyncTarget, error) {
	if target, ok := f.targets[name]; ok {
		return target(f.ctx), nil
	} else {
		return nil, fmt.Errorf("invalid target %s", name)
	}
}

func NewTargetFactory(ctx context.Context) TargetFactory {
	targets := map[string]SyncTargetFunc{
		"anilist":   TargetProvider(anilist.NewTarget),
		"hardcover": TargetProvider(hardcover.NewTarget),
	}
	return &targetFactory{
		ctx:     ctx,
		targets: targets,
	}
}
