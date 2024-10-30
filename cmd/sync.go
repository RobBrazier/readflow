package cmd

import (
	"fmt"
	"log/slog"

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
		slog.Info("sync called")
		targetNames := []string{}
		activeTargets := target.GetActiveTargets()
		enabledSource := getEnabledSource()
		for _, target := range activeTargets {
			targetNames = append(targetNames, target.GetName())
		}
		slog.Info("target", "active", targetNames)
		action := sync.NewSyncAction(enabledSource, activeTargets)
		action.Sync()
	},
}

func getEnabledSource() source.Source {
	return source.NewCalibreSource()
}

func init() {
	RootCmd.AddCommand(syncCmd)

	for _, target := range target.GetTargets() {
		name := target.GetName()
		flagName := fmt.Sprintf("%s-token", name)
		syncCmd.Flags().String(flagName, "", fmt.Sprintf("token for authenticating with %s (defaults to the value of [tokens.%s] in the configuration file)", target.GetHostname(), name))
		viper.BindPFlag(fmt.Sprintf("tokens.%s", name), syncCmd.Flags().Lookup(flagName))
	}
	syncCmd.Flags().String("calibre-db", "", "Location of the calibre metadata.db file (defaults to value of [databases.calibre] in the configuration file)")
	viper.BindPFlag("databases.calibre", syncCmd.Flags().Lookup("calibre-db"))
	syncCmd.Flags().String("calibreweb-db", "", "Location of the calibre-web app.db file (defaults to value of [databases.calibreweb] in the configuration file)")
	viper.BindPFlag("databases.calibreweb", syncCmd.Flags().Lookup("calibreweb-db"))
}
