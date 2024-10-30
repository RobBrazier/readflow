package cmd

import (
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:       "set",
	Short:     "Update",
	ValidArgs: slices.Collect(maps.Keys(validConfigKeys)),
	Args: func(cmd *cobra.Command, args []string) error {
		// First, check that we have exactly 2 arguments
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		// Then validate only the first argument
		return cobra.OnlyValidArgs(cmd, args[:1])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		switch validConfigKeys[key].(type) {
		case []string:
			sliceValue := strings.Split(value, ",")
			viper.Set(key, sliceValue)
			slog.Info("Updated config for", "key", key, "value", sliceValue)
		case string:
			viper.Set(key, value)
			slog.Info("Updated config for", "key", key, "value", value)
		case bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				slog.Error("Invalid value passed - expected boolean value for", "key", key, "value", value)
				return err
			}
			viper.Set(key, b)
			slog.Info("Updated config for", "key", key, "value", b)
		}
		return viper.WriteConfig()
	},
}

func init() {
	configCmd.AddCommand(setCmd)
}
