package source

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

const CHAPTERS_QUERY = "SELECT id FROM calibre.custom_columns WHERE label = ?;"

type databaseSource struct {
	calibre    string
	calibreweb string
}

type chaptersRow struct {
	Id int64
}

func (s *databaseSource) getReadOnlyDbString(file string) string {
	return fmt.Sprintf("file:%s?mode=ro", file)
}

func (s *databaseSource) getDb() *sqlx.DB {
	db := sqlx.MustConnect("sqlite3", s.getReadOnlyDbString(s.calibreweb))
	db.MustExec("attach database ? as calibre", s.getReadOnlyDbString(s.calibre))
	return db
}

func (s *databaseSource) GetChaptersColumn() (string, error) {
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

func NewCalibreSource() Source {
	return &databaseSource{
		calibre:    viper.GetString("databases.calibre"),
		calibreweb: viper.GetString("databases.calibreweb"),
	}
}
