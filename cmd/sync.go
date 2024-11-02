package cmd

import (
	"log/slog"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/sync"
	"github.com/RobBrazier/readflow/target"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync latest reading states to configured targets",
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

	rootCmd.AddCommand(syncCmd)

}
