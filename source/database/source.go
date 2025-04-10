package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/model"
	"github.com/RobBrazier/readflow/source/database/queries"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type databaseSource struct {
	internal.Source
	log            *log.Logger
	chaptersColumn string
}

func (s databaseSource) getSyncDays() int {
	return config.GetSyncDays()
}

func (s databaseSource) getReadOnlyDbString(file string) string {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		cobra.CheckErr(fmt.Sprintf("Unable to access database %s. Is the path correct?", file))
	}
	return fmt.Sprintf("file:%s?mode=ro", file)
}

func (s databaseSource) getDb() *sqlx.DB {
	db := sqlx.MustConnect("sqlite", s.getReadOnlyDbString(config.GetDatabases().CalibreWeb))
	db.MustExec("attach database ? as calibre", s.getReadOnlyDbString(config.GetDatabases().Calibre))
	return db
}

func (s *databaseSource) hasChaptersColumn() bool {
	s.chaptersColumn = config.GetColumns().Chapter
	if strings.ToLower(s.chaptersColumn) == "false" {
		return false
	}
	if s.chaptersColumn != "" {
		return true
	}
	return s.findChaptersColumn()
}

func (s databaseSource) getChaptersColumn() string {
	if s.hasChaptersColumn() {
		return s.chaptersColumn
	}
	return ""
}

func (s databaseSource) findChaptersColumn() bool {
	db := s.getDb()
	defer db.Close()
	name, err := queries.GetChaptersColumn(db)
	if err != nil {
		s.log.Debug("Unable to retrieve chapters column")
		return false
	}
	s.chaptersColumn = name
	conf := config.GetConfig()
	conf.Columns.Chapter = name

	s.log.Info("Stored chapters column", "column", name)
	config.SaveConfig(&conf)
	return true
}

func (s databaseSource) GetRecentReads() (books []model.Book, err error) {
	db := s.getDb()
	defer db.Close()

	chaptersColumn := s.getChaptersColumn()

	recentReads, err := queries.GetRecentReads(db, s.getSyncDays(), chaptersColumn)

	for _, read := range recentReads {
		book := MapToBook(read)
		// start collating series to check if I have chapter information
		if book.Series.ID != 0 && chaptersColumn != "" {
			seriesBooks, err := queries.GetPreviousBooksInSeries(db, chaptersColumn, book.Series.ID, book.Series.Index)
			if err != nil {
				s.log.Error("Unable to get previous books for", "book", book.Name, "error", err)
			}
			var mappedBooks []model.Book
			for _, seriesBook := range seriesBooks {
				mappedBooks = append(mappedBooks, MapToBook(seriesBook))
			}
			book.AddBooksToSeries(mappedBooks)
		}
		books = append(books, book)
	}

	return books, nil
}

func New(_ context.Context) internal.SyncSource {
	name := "database"
	return &databaseSource{
		log: log.WithPrefix(name),
		Source: internal.Source{
			Name: name,
		},
	}
}
