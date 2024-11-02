package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/target"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   internal.NAME,
	Short: "Track your Kobo reads on Anilist.co and Hardcover.app using Calibre-Web and Calibre databases",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := slog.LevelInfo
		if verbose {
			level = slog.LevelDebug
		}
		slog.SetLogLoggerLevel(level)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defaultConfigFile, err := xdg.SearchConfigFile(fmt.Sprintf("%s/config.yaml", RootCmd.Name()))
	cobra.CheckErr(err)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultConfigFile, "config file")

	availableTargets := []string{}
	for _, target := range target.GetTargets() {
		name := target.GetName()
		availableTargets = append(availableTargets, name)
	}
	RootCmd.PersistentFlags().StringSliceP("targets", "t", availableTargets, "Active targets to sync reading status with")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	viper.BindPFlag("targets", RootCmd.PersistentFlags().Lookup("targets"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configPath, err := xdg.SearchConfigFile(RootCmd.Name())
		cobra.CheckErr(err)

		// Search config in xdg config directory with name "readflow/config.yaml".
		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
