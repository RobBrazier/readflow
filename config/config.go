package config

import "strings"

type Config struct {
	Columns    ColumnConfig   `yaml:"columns"`
	Databases  DatabaseConfig `yaml:"databases"`
	Source     string         `yaml:"source" env:"SOURCE" default:"database"`
	Targets    []string       `yaml:"targets" env:"TARGETS"`
	Tokens     TokenConfig    `yaml:"tokens"`
	SyncDays   int            `yaml:"syncDays" env:"SYNC_DAYS" default:"1"`
	sourceFile string
}

// ColumnConfig represents the columns configuration
type ColumnConfig struct {
	Chapter string `yaml:"chapter" env:"COLUMN_CHAPTER"`
}

// DatabaseConfig represents the database paths configuration
type DatabaseConfig struct {
	Calibre    string `yaml:"calibre" env:"DATABASE_CALIBRE"`
	CalibreWeb string `yaml:"calibreweb" env:"DATABASE_CALIBREWEB"`
}

// TokenConfig represents the API tokens configuration
type TokenConfig struct {
	Anilist   string `yaml:"anilist" env:"TOKEN_ANILIST"`
	Hardcover string `yaml:"hardcover" env:"TOKEN_HARDCOVER"`
}

func (c Config) AreChaptersEnabled() bool {
	if strings.ToLower(c.Columns.Chapter) == "false" {
		return false
	}
	return true
}

func (c Config) GetSourceFile() string {
	return c.sourceFile
}
