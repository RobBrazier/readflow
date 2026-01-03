package source

type Source interface {
	Init() error
	GetRecentReads() ([]BookContext, error)
}

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

func GetSources(days int) map[string]Source {
	return map[string]Source{
		"database": NewDatabaseSource(days),
	}
}
