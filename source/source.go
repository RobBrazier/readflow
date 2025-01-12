package source

import (
	"context"

	"github.com/RobBrazier/readflow/config"
)

type Source interface {
	Name() string
	Init() error
	GetRecentReads() ([]BookContext, error)
}

type Book struct {
	BookID           int      `db:"book_id"`
	BookName         string   `db:"book_name"`
	SeriesID         *int     `db:"series_id"`
	BookSeriesIndex  *int     `db:"book_series_index"`
	ReadStatus       int      `db:"read_status"`
	ISBN             *string  `db:"isbn"`
	AnilistID        *string  `db:"anilist_id"`
	HardcoverID      *string  `db:"hardcover_id"`
	HardcoverEdition *string  `db:"hardcover_edition"`
	ProgressPercent  *float64 `db:"progress_percent"`
	ChapterCount     *int     `db:"chapter_count"`
}

type BookContext struct {
	Current  Book
	Previous []Book
}

type SourceFunc func(ctx context.Context) Source

func SourceProvider(fn SourceFunc) SourceFunc {
	return func(ctx context.Context) Source {
		return fn(ctx)
	}
}

func GetSources() map[string]SourceFunc {
	return map[string]SourceFunc{
		"database": SourceProvider(NewDatabaseSource),
	}
}

func GetActiveSource(enabled string, ctx context.Context) *Source {
	sourceFn, ok := GetSources()[enabled]
	if ok {
		source := sourceFn(ctx)
		return &source
	}
	return nil
}

func GetActiveSources(ctx context.Context) []string {
	cfg := config.GetFromContext(ctx)
	active := []string{}
	selectedSources := cfg.Source
	active = append(active, selectedSources)
	return active
}
