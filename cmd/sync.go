package cmd

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/internal/target"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var availableSources []string
var activeSource string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync latest reading states to configured targets",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if availableSources == nil {
			availableSources = slices.Collect(maps.Keys(source.GetSources()))
		}
		activeSources := source.GetActiveSources()
		if len(activeSources) > 0 {
			activeSource = activeSources[0]
		}
		if slices.Contains(availableSources, activeSource) {
			return nil
		}
		return errors.New(fmt.Sprintf("Invalid source. Available sources: %v", availableSources))
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("sync called", "source", activeSource)
		targetNames := []string{}
		activeTargets := target.GetActiveTargets()
		enabledSource := getEnabledSource()
		for _, target := range activeTargets {
			targetNames = append(targetNames, target.GetName())
		}
		log.Debug("target", "active", targetNames)
		action := sync.NewSyncAction(enabledSource, activeTargets)
		results, err := action.Sync()
		cobra.CheckErr(err)
		log.Info("sync completed", "results", results)
	},
}

func getEnabledSource() source.Source {
	sources := source.GetSources()
	if val, ok := sources[activeSource]; ok {
		return val
	}
	return nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
