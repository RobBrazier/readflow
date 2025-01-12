package sync

import (
	"sync"

	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target"
	"github.com/charmbracelet/log"
)

type SyncResult struct {
	Name string
}

type SyncAction interface {
	Sync() ([]SyncResult, error)
}

type syncAction struct {
	log     *log.Logger
	targets []target.SyncTarget
	source  source.Source
}

func (a *syncAction) Sync() ([]SyncResult, error) {
	// if the chapters column doesn't exist in config, fetch the name and store it
	a.source.Init()
	recentReads, err := a.source.GetRecentReads()
	if err != nil {
		return nil, err
	}
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
	for _, book := range reads {
		name := book.Current.BookName
		log := log.With("target", t.Name(), "book", name)
		if !t.ShouldProcess(book) {
			log.Debug("Skipping processing of ineligible book")
			continue
		}
		log.Info("Processing")
		err := t.UpdateReadStatus(book)
		if err != nil {
			log.Error("failed to update reading status", "error", err)
		}
	}
}

func NewSyncAction(enabledSource source.Source, targets []target.SyncTarget) SyncAction {
	return &syncAction{
		targets: targets,
		source:  enabledSource,
	}
}
