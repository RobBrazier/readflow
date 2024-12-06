package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/RobBrazier/readflow/internal"
	"github.com/adrg/xdg"
	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

func GetConfigPath(override *string) string {
	// Has the config flag been passed in? - if it's got a value, use it
	if override != nil {
		if *override != "" {
			configPath = *override
		}
	}

	if configPath == "" {
		// look in the XDG_CONFIG_HOME location
		var err error
		configPath, err = xdg.ConfigFile(filepath.Join(internal.NAME, "config.yaml"))

		if err != nil {
			// if that doesn't work for some reason, fall back to the current dir
			currentDir, err := os.Getwd()
			if err != nil {
				currentDir = "."
			}
			configPath = filepath.Join(currentDir, "readflow.yaml")
		}
	}

	return configPath

}

func LoadConfigFromEnv() error {
	err := env.Parse(&config)
	return err
}

func LoadConfig(path string) error {
	log.Debug("Loading config from", "file", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	log.Debug("Successfully loaded config")
	return nil
}

func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if _, err = os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		// file doesn't exist - lets create the folder structure required
		path := filepath.Dir(configPath)
		if path != "." {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				log.Error("Something went wrong when trying to create the folder structure for", "file", configPath, "path", path, "error", err)
			}
		}
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		log.Error("Couldn't save config to", "file", configPath)
		return err
	}
	log.Debug("Successfully saved config to", "file", configPath)
	config = *cfg
	return nil
}
