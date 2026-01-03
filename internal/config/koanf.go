package config

import (
	"os"
	"slices"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
)

const DELIMITER = "."

var k = koanf.New(DELIMITER)

func Load(filePath string) {
	_ = k.Load(confmap.Provider(map[string]any{
		"sync.days":      7,
		"source.enabled": []string{"database"},
		"target.enabled": []string{"hardcover"},
	}, DELIMITER), nil)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		log.Info().Str("file", filePath).Msg("reading config from file")
		if err := k.Load(file.Provider(filePath), toml.Parser()); err != nil {
			log.Fatal().Err(err).Msgf("error loading config file: %s", filePath)
		}
	}

	// legacy environment variables
	_ = k.Load(env.Provider(DELIMITER, env.Opt{
		TransformFunc: func(k, v string) (string, any) {
			allowed := map[string]string{
				"SOURCE":              "source.enabled",
				"TARGETS":             "target.enabled",
				"SYNC_DAYS":           "sync.days",
				"COLUMN_CHAPTER":      "source.chapter_column",
				"DATABASE_CALIBRE":    "source.database.calibre",
				"DATABASE_CALIBREWEB": "source.database.calibre_web",
				"TOKEN_ANILIST":       "target.anilist.token",
				"TOKEN_HARDCOVER":     "target.hardcover.token",
			}
			if key, ok := allowed[k]; ok {
				if strings.Contains(v, " ") {
					return key, strings.Split(v, " ")
				}
				if strings.Contains(v, ",") {
					return key, strings.Split(v, ",")
				}
				return key, v
			}
			return "", ""
		},
	}), nil)

	_ = k.Load(env.Provider(DELIMITER, env.Opt{
		Prefix: "READFLOW_",
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "READFLOW_")), "_", ".")
			if strings.Contains(v, ",") {
				return k, strings.Split(v, ",")
			}
			// always convert slice environment variables to slices
			if slices.Contains([]string{"source.enabled", "target.enabled"}, k) {
				return k, []string{v}
			}
			return k, v
		},
	}), nil)
}

func Get() *koanf.Koanf {
	return k
}

func Set(key string, val any) error {
	return k.Set(key, val)
}
