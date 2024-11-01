package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/sync"
	"github.com/RobBrazier/readflow/target"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var availableSources []string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync latest reading states to configured targets",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("sync called", "source", viper.GetString("source"))
		targetNames := []string{}
		activeTargets := target.GetActiveTargets()
		enabledSource := getEnabledSource()
		for _, target := range activeTargets {
			targetNames = append(targetNames, target.GetName())
		}
		slog.Info("target", "active", targetNames)
		action := sync.NewSyncAction(enabledSource, activeTargets)
		results, err := action.Sync()
		cobra.CheckErr(err)
		slog.Info("sync completed", "results", results)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if availableSources == nil {
			availableSources = slices.Collect(maps.Keys(source.GetSources()))
		}
		if slices.Contains(availableSources, viper.GetString("source")) {
			return nil
		}
		return errors.New(fmt.Sprintf("Invalid source. Available sources: %v", availableSources))
	},
}

func getEnabledSource() source.Source {
	key := viper.GetString("source")
	sources := source.GetSources()
	if val, ok := sources[key]; ok {
		return val
	}
	return nil
}

func init() {

	RootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("source", "s", "database", "Active source to retrieve reading data from")
	viper.BindPFlag("source", syncCmd.Flags().Lookup("source"))
}
