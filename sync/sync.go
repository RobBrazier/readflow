package sync

import (
	"encoding/json"
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
	b, err := json.Marshal(recentReads)
	cobra.CheckErr(err)
	slog.Info(string(b))
	var wg sync.WaitGroup
	for _, t := range a.targets {
		wg.Add(1)
		go a.processTarget(t, &wg)
	}

	wg.Wait()
	return []SyncResult{}, nil
}

func (a *syncAction) processTarget(t target.SyncTarget, wg *sync.WaitGroup) {
	defer wg.Done()
	user := t.GetCurrentUser()
	slog.Info("current user for", "target", t.GetName(), "user", user)
}

func NewSyncAction(enabledSource source.Source, targets []target.SyncTarget) SyncAction {
	return &syncAction{
		targets: targets,
		source:  enabledSource,
	}
}
