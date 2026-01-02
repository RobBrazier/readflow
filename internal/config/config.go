package config

type Config struct {
	Source SourceConfig `yaml:"source"`
	Target TargetConfig `yaml:"target"`
}

type SourceConfig struct {
	Database DatabaseConfig `yaml:"database"`
}

type DatabaseConfig struct {
	Calibre    string       `yaml:"calibre"    env:"DATABASE_CALIBRE"`
	CalibreWeb string       `yaml:"calibreweb" env:"DATABASE_CALIBREWEB"`
	Columns    ColumnConfig `yaml:"columns"`
}

type ColumnConfig struct {
	Chapter string `yaml:"chapter" env:"COLUMN_CHAPTER"`
}

type TargetConfig struct {
	Hardcover HardcoverConfig `yaml:"hardcover"`
	Anilist   AnilistConfig   `yaml:"anilist"`
}

type HardcoverConfig struct {
	Token string `yaml:"token" env:"TOKEN_HARDCOVER"`
}

type AnilistConfig struct {
	Token string `yaml:"token" env:"TOKEN_ANILIST"`
}
