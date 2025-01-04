package source

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type databaseSource struct {
	name           string
	log            *log.Logger
	chaptersColumn string
	enableChapters bool
	config         *config.Config
	databases      config.DatabaseConfig
	syncDays       int
}

type chaptersRow struct {
	Id int64
}

const CHAPTERS_COLUMN = "columns.chapter"

func (s *databaseSource) getReadOnlyDbString(file string) string {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		log.Fatal("Unable to access database. Is the path correct?", "path", file)
	}
	return fmt.Sprintf("file:%s?mode=ro", file)
}

func (s *databaseSource) getDb() *sqlx.DB {
	db := sqlx.MustConnect("sqlite", s.getReadOnlyDbString(s.databases.CalibreWeb))
	db.MustExec("attach database ? as calibre", s.getReadOnlyDbString(s.databases.Calibre))
	return db
}

func (s databaseSource) Name() string {
	return s.name
}

func (s *databaseSource) Init() error {
	// figure out the chapters column only if it's enabled
	if s.chaptersColumn == "" && s.enableChapters {
		column, err := s.getChaptersColumn()
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to find chapters column - configure via `%s config set %s NAME` (or set to 'false' to disable reading progress tracking)", internal.NAME, CHAPTERS_COLUMN))
		}
		s.chaptersColumn = column
		s.config.Columns.Chapter = column

		s.log.Info("Stored chapters column", "column", column)
		config.SaveConfig(s.config)
	}
	s.log.Debug("column", "enabled", s.enableChapters, "name", s.chaptersColumn)
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
	var books = []Book{}

	query := RECENT_READS_QUERY_NO_CHAPTERS
	if s.chaptersColumn != "" {
		query = fmt.Sprintf(RECENT_READS_QUERY, s.chaptersColumn)
	}

	daysToQuery := fmt.Sprintf("-%d day", s.syncDays)

	log.Debug("Running source query with", "days", daysToQuery)

	err := db.Select(&books, query, daysToQuery)
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (s *databaseSource) GetRecentReads() ([]BookContext, error) {
	var recent = []BookContext{}
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
		var previous = []Book{}
		context := BookContext{
			Current: book,
		}
		if book.SeriesID != nil {
			err := db.Select(&previous, query, book.SeriesID, book.BookSeriesIndex)
			if err != nil {
				s.log.Error("Unable to get previous books for", "book", book.BookName)
			}
			context.Previous = previous
		} else {
			s.log.Info("Skipping retrieval of previous books as this book has no series", "book", book.BookName)
		}
		recent = append(recent, context)
	}
	return recent, nil
}

func NewDatabaseSource(ctx context.Context) Source {
	cfg := config.GetFromContext(ctx)
	chapters := cfg.Columns.Chapter
	name := "database"
	logger := log.WithPrefix(name)
	return &databaseSource{
		name:           name,
		log:            logger,
		chaptersColumn: chapters,
		enableChapters: cfg.AreChaptersEnabled(),
		config:         cfg,
		databases:      cfg.Databases,
		syncDays:       cfg.SyncDays,
	}
}
