package source

import (
	"sync"
	"sync/atomic"

	"github.com/RobBrazier/readflow/internal/config"
)

type Source interface {
	Init() error
	GetRecentReads() ([]BookContext, error)
}

var (
	sources     atomic.Pointer[map[string]Source]
	sourcesOnce sync.Once
)

type Book struct {
	BookID          int      `db:"book_id"`
	BookName        string   `db:"book_name"`
	SeriesID        *int     `db:"series_id"`
	BookSeriesIndex *int     `db:"book_series_index"`
	ReadStatus      int      `db:"read_status"`
	ISBN            *string  `db:"isbn"`
	AnilistID       *string  `db:"anilist_id"`
	HardcoverID     *string  `db:"hardcover_id"`
	ProgressPercent *float64 `db:"progress_percent"`
	ChapterCount    *int     `db:"chapter_count"`
}

type BookContext struct {
	Current  Book
	Previous []Book
}

func GetSources() map[string]Source {
	s := sources.Load()
	if s == nil {
		sourcesOnce.Do(func() {
			sources.CompareAndSwap(nil, &map[string]Source{
				"database": NewDatabaseSource(),
			})
		})
		s = sources.Load()
	}
	return *s
}

func GetActiveSources() []string {
	active := []string{}
	selectedSources := config.GetSource()
	active = append(active, selectedSources)
	return active
}
