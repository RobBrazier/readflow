package cmd

import (
	"log/slog"
	"maps"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:       "get",
	Short:     "Get configuration values",
	ValidArgs: slices.Collect(maps.Keys(validConfigKeys)),
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		slog.Info(key, "value", viper.Get(key))
	},
}

func init() {
	configCmd.AddCommand(getCmd)
}
