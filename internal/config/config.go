package config

type Config struct {
	Columns   ColumnConfig   `yaml:"columns"`
	Databases DatabaseConfig `yaml:"databases"`
	Source    string         `yaml:"source" env:"SOURCE"`
	Targets   []string       `yaml:"targets" env:"TARGETS"`
	Tokens    TokenConfig    `yaml:"tokens"`
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

var (
	config     Config
	configPath string
)

func GetConfig() Config {
	return config
}

func GetColumns() ColumnConfig {
	return config.Columns
}

func GetDatabases() DatabaseConfig {
	return config.Databases
}

func GetSource() string {
	return config.Source
}

func GetTargets() []string {
	return config.Targets
}

func GetTokens() TokenConfig {
	return config.Tokens
}
