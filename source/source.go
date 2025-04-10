package source

import (
	"context"

	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/source/database"
	"github.com/charmbracelet/log"
)

type SourceFunc func(context.Context) internal.SyncSource

var sources = map[string]SourceFunc{
	"database": database.New,
}

func GetSources() map[string]SourceFunc {
	return sources
}

func GetActiveSources(ctx context.Context) []internal.SyncSource {
	var active []internal.SyncSource
	selectedSources := config.GetSource()
	log.Info("selected source", "source", selectedSources)
	for name, source := range sources {
		if name == selectedSources {
			active = append(active, source(ctx))
		}
	}
	return active
}
