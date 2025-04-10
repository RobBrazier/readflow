package cmd

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/sync"
	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var activeSource string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync latest reading states to configured targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		enabledSource := getEnabledSource(cmd.Context())
		if enabledSource == nil {
			availableSources := slices.Collect(maps.Keys(source.GetSources()))
			return errors.New(fmt.Sprintf("Invalid source. Available sources: %v", availableSources))
		}
		log.Debug("sync called", "source", enabledSource.GetName())
		reads, err := enabledSource.GetRecentReads()
		if err != nil {
			return err
		}
		for _, read := range reads {
			log.Info("recent reads", "val", read)
		}
		targetNames := []string{}
		activeTargets := target.GetActiveTargets(cmd.Context())
		for _, target := range activeTargets {
			targetNames = append(targetNames, target.GetName())
		}
		log.Debug("target", "active", targetNames)
		action := sync.NewSyncAction(enabledSource, activeTargets)
		results, err := action.Sync()
		if err != nil {
			return err
		}
		log.Info("sync completed", "results", results)
		return nil
	},
}

func getEnabledSource(ctx context.Context) internal.SyncSource {
	sources := source.GetActiveSources(ctx)
	if len(sources) > 0 {
		return sources[0]
	}
	return nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
