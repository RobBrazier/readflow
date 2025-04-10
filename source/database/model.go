package database

import (
	"database/sql"

	"github.com/RobBrazier/readflow/internal/model"
	"github.com/RobBrazier/readflow/source/database/queries"
)

func expandNullValue[T comparable](value sql.Null[T]) T {
	var defaultValue T
	if !value.Valid {
		return defaultValue
	}
	return value.V
}

func MapToBook(source queries.DatabaseBook) model.Book {
	identifiers := model.BookIdentifiers{
		Anilist:          expandNullValue(source.AnilistID),
		Hardcover:        expandNullValue(source.HardcoverID),
		HardcoverEdition: expandNullValue(source.HardcoverEditionID),
	}
	progress := model.BookProgress{
		Local: expandNullValue(source.ProgressPercent),
	}
	series := model.BookSeries{
		ID:    expandNullValue(source.SeriesID),
		Index: expandNullValue(source.BookSeriesIndex),
	}
	book := model.Book{
		ID:           source.BookID,
		Name:         source.BookName,
		Identifiers:  identifiers,
		Progress:     progress,
		ChapterCount: expandNullValue(source.ChapterCount),
		Series:       series,
	}
	return book
}
