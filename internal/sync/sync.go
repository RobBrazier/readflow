package sync

import (
	"sync"

	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/rs/zerolog/log"
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
	var result []SyncResult
	if err != nil {
		return result, err
	}
	var wg sync.WaitGroup
	for _, t := range a.targets {
		wg.Add(1)
		go a.processTarget(t, recentReads, &wg)
	}

	wg.Wait()
	return result, nil
}

func (a *syncAction) processTarget(
	t target.SyncTarget,
	reads []source.BookContext,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for _, book := range reads {
		name := book.Current.BookName
		log := log.With().Str("target", t.GetName()).Str("book", name).Logger()
		if !t.ShouldProcess(book) {
			log.Debug().Msg("Skipping processing of ineligible book")
			continue
		}
		log.Info().Msg("Processing")
		err := t.UpdateReadStatus(book)
		if err != nil {
			log.Error().Err(err).Msg("failed to update reading status")
		}
	}
}

func NewSyncAction(sources []source.Source, targets []target.SyncTarget) SyncAction {
	enabledSource := sources[0]
	return &syncAction{
		targets: targets,
		source:  enabledSource,
	}
}
