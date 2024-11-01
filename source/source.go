package source

type Source interface {
	Init() error
	GetRecentReads() ([]Book, error)
}

var sources map[string]Source

type Book struct {
	BookID          int      `json:"book_id" db:"book_id"`
	BookName        string   `json:"book_name" db:"book_name"`
	SeriesID        *int     `json:"series_id" db:"series_id"`                 // Nullable, use pointer to handle NULLs
	BookSeriesIndex *int     `json:"book_series_index" db:"book_series_index"` // Nullable, use pointer to handle NULLs
	ReadStatus      int      `json:"read_status" db:"read_status"`
	ISBN            *string  `json:"isbn" db:"isbn"`
	AnilistID       *string  `json:"anilist_id" db:"anilist_id"`             // Nullable, use pointer to handle NULLs
	HardcoverID     *string  `json:"hardcover_id" db:"hardcover_id"`         // Nullable, use pointer to handle NULLs
	ProgressPercent *float64 `json:"progress_percent" db:"progress_percent"` // Nullable, use pointer to handle NULLs
	ChapterCount    *int     `json:"chapter_count" db:"chapter_count"`       // Nullable, use pointer to handle NULLs
}

func GetSources() map[string]Source {
	if sources == nil {
		sources = make(map[string]Source)
		sources["database"] = NewDatabaseSource()
	}
	return sources
}
