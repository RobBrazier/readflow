package cmd

import (
	"github.com/RobBrazier/readflow/cmd"
	"github.com/spf13/cobra"
)

var validConfigKeys = map[string]interface{}{
	"targets":              []string{},
	"tokens.anilist":       "",
	"tokens.hardcover":     "",
	"databases.calibreweb": "",
	"databases.calibre":    "",
	"columns.chapter":      "",
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Get or Set configuration values",
}

func init() {
	cmd.RootCmd.AddCommand(configCmd)
}
