package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/internal"
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if availableSources == nil {
			availableSources = slices.Collect(maps.Keys(source.GetSources()))
		}
		if slices.Contains(availableSources, viper.GetString(internal.CONFIG_SOURCE)) {
			return nil
		}
		return errors.New(fmt.Sprintf("Invalid source. Available sources: %v", availableSources))
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("sync called", "source", viper.GetString(internal.CONFIG_SOURCE))
		targetNames := []string{}
		activeTargets := target.GetActiveTargets()
		enabledSource := getEnabledSource()
		for _, target := range activeTargets {
			targetNames = append(targetNames, target.GetName())
		}
		slog.Debug("target", "active", targetNames)
		action := sync.NewSyncAction(enabledSource, activeTargets)
		results, err := action.Sync()
		cobra.CheckErr(err)
		slog.Info("sync completed", "results", results)
	},
}

func getEnabledSource() source.Source {
	key := viper.GetString(internal.CONFIG_SOURCE)
	sources := source.GetSources()
	if val, ok := sources[key]; ok {
		return val
	}
	return nil
}

func init() {
	// availableTargets := []string{}
	// for _, target := range target.GetTargets() {
	// 	name := target.GetName()
	// 	availableTargets = append(availableTargets, name)
	// }
	//
	// syncCmd.PersistentFlags().StringSliceP("targets", "t", availableTargets, "Active targets to sync reading status with")
	// viper.BindPFlag("targets", syncCmd.Flags().Lookup("targets"))
	//
	// syncCmd.PersistentFlags().StringP("source", "s", "database", "Active source to retrieve reading data from")
	// viper.BindPFlag("source", syncCmd.Flags().Lookup("source"))

	rootCmd.AddCommand(syncCmd)
}
