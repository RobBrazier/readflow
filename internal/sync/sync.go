package sync

import (
	"sync"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/model"
	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
)

type SyncResult struct {
	Name string
}

type SyncAction interface {
	Sync() ([]SyncResult, error)
}

type syncAction struct {
	log     *log.Logger
	targets []internal.SyncTarget
	source  internal.SyncSource
}

func (a *syncAction) Sync() ([]SyncResult, error) {
	// if the chapters column doesn't exist in config, fetch the name and store it
	recentReads, err := a.source.GetRecentReads()
	if err != nil {
		return nil, err
	}

	titles := []string{}
	for _, read := range recentReads {
		titles = append(titles, read.Name)
	}
	if len(titles) > 0 {
		log.Info("Found recent reads", "count", len(titles), "titles", titles)
	}
	var wg sync.WaitGroup
	for _, t := range a.targets {
		wg.Add(1)
		go a.processTarget(t, recentReads, &wg)
	}

	wg.Wait()
	return []SyncResult{}, nil
}

func (a *syncAction) processTarget(t internal.SyncTarget, reads []model.Book, wg *sync.WaitGroup) {
	defer wg.Done()
	processed, err := t.ProcessReads(reads)
	if err != nil {
		log.Error("error processing reads", "err", err)
		return
	}
	for _, book := range processed {
		name := book.Name
		log := log.With("target", t.GetName(), "book", name)
		log.Debug("Updating status for", "book", spew.Sdump(book))
		err := t.UpdateStatus(book)
		if err != nil {
			log.Error("failed to update reading status", "error", err)
		}
	}
}

func NewSyncAction(enabledSource internal.SyncSource, targets []internal.SyncTarget) SyncAction {
	return &syncAction{
		targets: targets,
		source:  enabledSource,
	}
}
