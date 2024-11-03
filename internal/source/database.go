package source

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/RobBrazier/readflow/internal"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

type databaseSource struct {
	log            *log.Logger
	calibre        string
	calibreweb     string
	chaptersColumn string
	enableChapters bool
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
	db := sqlx.MustConnect("sqlite", s.getReadOnlyDbString(s.calibreweb))
	db.MustExec("attach database ? as calibre", s.getReadOnlyDbString(s.calibre))
	return db
}

func (s *databaseSource) Init() error {
	// figure out the chapters column only if it's enabled
	if s.chaptersColumn == "" && s.enableChapters {
		column, err := s.getChaptersColumn()
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to find chapters column - configure via `%s config set %s NAME` (or set to 'false' to disable reading progress tracking)", internal.NAME, CHAPTERS_COLUMN))
		}
		s.chaptersColumn = column
		viper.Set(CHAPTERS_COLUMN, column)
		s.log.Info("Stored chapters column", "column", column)
		viper.WriteConfig()
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

	daysToQuery := "-7 day"

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

func NewDatabaseSource() Source {
	chapters := viper.GetString(CHAPTERS_COLUMN)
	enableChapters := true
	if strings.ToLower(chapters) == "false" {
		enableChapters = false
		chapters = ""
	}
	logger := log.WithPrefix("database")
	return &databaseSource{
		log:            logger,
		calibre:        viper.GetString(internal.CONFIG_CALIBRE_DB),
		calibreweb:     viper.GetString(internal.CONFIG_CALIBREWEB_DB),
		chaptersColumn: chapters,
		enableChapters: enableChapters,
	}
}
