package queries

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:embed *.sql
var queries embed.FS

type chaptersRow struct {
	Id int64
}

func getQuery(file string, args ...any) (string, error) {
	contents, err := queries.ReadFile(fmt.Sprintf("%s.sql", file))
	if err != nil {
		return "", err
	}
	query := string(contents)
	return fmt.Sprintf(query, args...), nil

}

func GetChaptersColumn(db *sqlx.DB) (string, error) {
	var row chaptersRow
	query, err := getQuery("chapters_column")
	if err != nil {
		return "", err
	}
	err = db.Get(&row, query, "chapters")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("custom_column_%d", row.Id), nil
}

func GetRecentReads(db *sqlx.DB, syncDays int, chaptersColumn string) ([]DatabaseBook, error) {
	var books []DatabaseBook
	var args []any
	queryFile := "recent_reads"
	if chaptersColumn != "" {
		args = append(args, chaptersColumn)
		queryFile += "_chapters"
	}
	query, err := getQuery(queryFile, args...)
	if err != nil {
		return nil, err
	}
	daysToQuery := fmt.Sprintf("-%d day", syncDays)
	err = db.Select(&books, query, daysToQuery)
	if err != nil {
		return nil, err
	}
	return books, nil
}

func GetPreviousBooksInSeries(db *sqlx.DB, chaptersColumn string, seriesId, seriesIndex int) ([]DatabaseBook, error) {
	var books []DatabaseBook
	query, err := getQuery("previous_books_in_series_chapters", chaptersColumn)
	if err != nil {
		return nil, err
	}
	err = db.Select(&books, query, seriesId, seriesIndex)
	if err != nil {
		return nil, err
	}
	return books, nil
}
