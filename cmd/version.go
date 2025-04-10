package cmd

import (
	"fmt"

	"github.com/RobBrazier/readflow/internal"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of readflow",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("readflow version=%s, commit=%s, date=%s\n", internal.Version, internal.Commit, internal.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
