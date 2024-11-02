package source

import (
	"github.com/RobBrazier/readflow/internal"
	"github.com/spf13/viper"
)

type Source interface {
	Init() error
	GetRecentReads() ([]BookContext, error)
}

var sources map[string]Source

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
	if sources == nil {
		sources = make(map[string]Source)
		sources["database"] = NewDatabaseSource()
	}
	return sources
}

func GetActiveSources() []string {
	active := []string{}
	selectedSources := viper.GetString(internal.CONFIG_SOURCE)
	active = append(active, selectedSources)
	return active
}
