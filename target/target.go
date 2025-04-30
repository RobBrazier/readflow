package target

import (
	"context"
	"slices"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/target/anilist"
	"github.com/RobBrazier/readflow/target/hardcover"
	"github.com/charmbracelet/log"
)

type TargetFunc func(context.Context) internal.SyncTarget

var targets = map[string]TargetFunc{
	"hardcover": hardcover.New,
	"anilist":   anilist.New,
}

func GetTargets() map[string]TargetFunc {
	return targets
}

func GetActiveTargets(ctx context.Context, selectedTargets []string) []internal.SyncTarget {
	var active []internal.SyncTarget
	log.Info("selected target", "target", selectedTargets)
	for name, target := range targets {
		if slices.Contains(selectedTargets, name) {
			active = append(active, target(ctx))
		}
	}
	return active
}
