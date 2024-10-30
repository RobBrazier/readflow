package sync

import (
	"log/slog"
	"strings"

	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target"
	"github.com/spf13/viper"
)

const CHAPTERS_COLUMN = "columns.chapter"

type SyncResult struct {
	Name string
}

type SyncAction interface {
	Sync() ([]SyncResult, error)
}

type syncAction struct {
	targets        []target.SyncTarget
	chaptersColumn string
	enableChapters bool
	source         source.Source
}

func (a *syncAction) ensureChaptersColumn() {
	// figure out the chapters column only if it's enabled
	if a.chaptersColumn == "" && a.enableChapters {
		column, err := a.source.GetChaptersColumn()
		if err != nil {
			slog.Error("Unable to find chapters column", "err", err)
			return
		}
		a.chaptersColumn = column
		viper.Set(CHAPTERS_COLUMN, column)
		slog.Info("Stored chapters column", "column", column)
		viper.WriteConfig()
	}
	slog.Debug("column", "enabled", a.enableChapters, "name", a.chaptersColumn)

}

func (a *syncAction) Sync() ([]SyncResult, error) {
	// if the chapters column doesn't exist in config, fetch the name and store it
	a.ensureChaptersColumn()
	results := []SyncResult{}
	return results, nil
}

func NewSyncAction(enabledSource source.Source, targets []target.SyncTarget) SyncAction {
	chapters := viper.GetString(CHAPTERS_COLUMN)
	enableChapters := true
	if strings.ToLower(chapters) == "false" {
		enableChapters = false
		chapters = ""
	}
	return &syncAction{
		targets:        targets,
		chaptersColumn: chapters,
		enableChapters: enableChapters,
		source:         enabledSource,
	}
}
