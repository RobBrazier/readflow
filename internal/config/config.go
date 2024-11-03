package config

type Config struct {
	Columns   ColumnConfig   `yaml:"columns"`
	Databases DatabaseConfig `yaml:"databases"`
	Source    string         `yaml:"source"`
	Targets   []string       `yaml:"targets"`
	Tokens    TokenConfig    `yaml:"tokens"`
}

// ColumnConfig represents the columns configuration
type ColumnConfig struct {
	Chapter string `yaml:"chapter"`
}

// DatabaseConfig represents the database paths configuration
type DatabaseConfig struct {
	Calibre    string `yaml:"calibre"`
	CalibreWeb string `yaml:"calibreweb"`
}

// TokenConfig represents the API tokens configuration
type TokenConfig struct {
	Anilist   string `yaml:"anilist"`
	Hardcover string `yaml:"hardcover"`
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
