package queries

import "database/sql"

type DatabaseBook struct {
	BookID             int               `db:"book_id"`
	BookName           string            `db:"book_name"`
	SeriesID           sql.Null[int]     `db:"series_id"`
	BookSeriesIndex    sql.Null[int]     `db:"book_series_index"`
	AnilistID          sql.Null[string]  `db:"anilist_id"`
	HardcoverID        sql.Null[string]  `db:"hardcover_id"`
	HardcoverEditionID sql.Null[string]  `db:"hardcover_edition_id"`
	ProgressPercent    sql.Null[float64] `db:"progress_percent"`
	ChapterCount       sql.Null[int]     `db:"chapter_count"`
}
