package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-yaml"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

func LoadConfig(path string) (Config, error) {
	var config Config
	log.Debug().Str("file", path).Msg("Loading config from")
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	if err := env.Parse(&config); err != nil {
		return config, err
	}
	log.Debug().Msg("Successfully loaded config")
	return config, nil
}
