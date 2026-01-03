package source

import (
	"errors"
	"fmt"
	"os"

	"github.com/RobBrazier/readflow/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type databaseSource struct {
	log            zerolog.Logger
	chaptersColumn string
	enableChapters bool
	syncDays       int
}

type chaptersRow struct {
	Id int64
}

const CHAPTERS_COLUMN = "columns.chapter"

func (s *databaseSource) getReadOnlyDbString(file string) string {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		cobra.CheckErr(fmt.Sprintf("Unable to access database %s. Is the path correct?", file))
	}
	return fmt.Sprintf("file:%s?mode=ro", file)
}

func (s *databaseSource) getDb() *sqlx.DB {
	calibreDb := config.Get().String("source.database.calibre")
	if _, err := os.Stat(calibreDb); err != nil {
		log.Fatal().
			Err(err).
			Str("config key", "source.database.calibre").
			Msg("Unable to read calibre database")
	}
	calibreWebDb := config.Get().String("source.database.calibre_web")
	if _, err := os.Stat(calibreWebDb); err != nil {
		log.Fatal().
			Err(err).
			Str("config key", "source.database.calibre_web").
			Msg("Unable to read calibre-web database")
	}
	db := sqlx.MustConnect("sqlite", s.getReadOnlyDbString(calibreWebDb))
	db.MustExec("attach database ? as calibre", s.getReadOnlyDbString(calibreDb))
	return db
}

func (s *databaseSource) Init() error {
	return nil
}

func (s *databaseSource) getChaptersColumn() (string, error) {
	var row chaptersRow
	db := s.getDb()
	defer db.Close()
	// Search for a custom column with a label of 'chapters' and store the value
	err := db.Get(&row, CHAPTERS_QUERY, "chapters")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("custom_column_%d", row.Id), nil
}

func (s *databaseSource) getRecentReads(db *sqlx.DB) ([]Book, error) {
	books := []Book{}

	query := RECENT_READS_QUERY_NO_CHAPTERS
	if s.chaptersColumn != "" {
		query = fmt.Sprintf(RECENT_READS_QUERY, s.chaptersColumn)
	}

	daysToQuery := fmt.Sprintf("-%d day", s.syncDays)

	log.Debug().Str("days", daysToQuery).Msg("Running source query with")

	err := db.Select(&books, query, daysToQuery)
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (s *databaseSource) GetRecentReads() ([]BookContext, error) {
	recent := []BookContext{}
	db := s.getDb()
	defer db.Close()
	recentReads, err := s.getRecentReads(db)
	if err != nil {
		return nil, err
	}
	query := PREVIOUS_BOOKS_IN_SERIES_QUERY_NO_CHAPTERS
	if s.chaptersColumn != "" {
		query = fmt.Sprintf(PREVIOUS_BOOKS_IN_SERIES_QUERY, s.chaptersColumn)
	}
	for _, book := range recentReads {
		previous := []Book{}
		context := BookContext{
			Current: book,
		}
		log := s.log.With().Str("book", book.BookName).Logger()
		if book.SeriesID != nil {
			err := db.Select(&previous, query, book.SeriesID, book.BookSeriesIndex)
			if err != nil {
				log.Error().Msg("Unable to get previous books for")
			}
			context.Previous = previous
		} else {
			log.Info().Msg("Skipping retrieval of previous books as this book has no series")
		}
		recent = append(recent, context)
	}
	return recent, nil
}

func NewDatabaseSource(days int) Source {
	chapters := config.Get().String("source.database.chapter_column")
	enableChapters := chapters != ""
	logger := log.With().Str("source", "database").Logger()
	return &databaseSource{
		log:            logger,
		chaptersColumn: chapters,
		enableChapters: enableChapters,
		syncDays:       days,
	}
}
