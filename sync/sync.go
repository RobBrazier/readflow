package sync

import (
	"log/slog"
	"sync"

	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target"
	"github.com/spf13/cobra"
)

type SyncResult struct {
	Name string
}

type SyncAction interface {
	Sync() ([]SyncResult, error)
}

type syncAction struct {
	targets []target.SyncTarget
	source  source.Source
}

func (a *syncAction) Sync() ([]SyncResult, error) {
	// if the chapters column doesn't exist in config, fetch the name and store it
	a.source.Init()
	recentReads, err := a.source.GetRecentReads()
	cobra.CheckErr(err)
	var wg sync.WaitGroup
	for _, t := range a.targets {
		wg.Add(1)
		go a.processTarget(t, recentReads, &wg)
	}

	wg.Wait()
	return []SyncResult{}, nil
}

func (a *syncAction) processTarget(t target.SyncTarget, reads []source.BookContext, wg *sync.WaitGroup) {
	defer wg.Done()
	user := t.GetCurrentUser()
	slog.Debug("current user for", "target", t.GetName(), "user", user)

	for _, book := range reads {
		slog.Info("Processing", "book", book.Current.BookName, "target", t.GetName())
		err := t.UpdateReadStatus(book)
		if err != nil {
			slog.Error("failed to update reading status", "error", err)
		}
	}
}

func NewSyncAction(enabledSource source.Source, targets []target.SyncTarget) SyncAction {
	return &syncAction{
		targets: targets,
		source:  enabledSource,
	}
}
