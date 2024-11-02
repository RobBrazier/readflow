package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var availableSources []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   internal.NAME,
	Short: "Track your Kobo reads on Anilist.co and Hardcover.app using Calibre-Web and Calibre databases",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := slog.LevelInfo
		if verbose {
			level = slog.LevelDebug
		}
		slog.SetLogLoggerLevel(level)

		if availableSources == nil {
			availableSources = slices.Collect(maps.Keys(source.GetSources()))
		}
		if slices.Contains(availableSources, viper.GetString(internal.CONFIG_SOURCE)) {
			return nil
		}
		return errors.New(fmt.Sprintf("Invalid source. Available sources: %v", availableSources))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defaultConfigFile, err := xdg.SearchConfigFile(fmt.Sprintf("%s/config.yaml", rootCmd.Name()))
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultConfigFile, "config file")

	availableTargets := []string{}
	for _, target := range target.GetTargets() {
		name := target.GetName()
		availableTargets = append(availableTargets, name)
	}
	rootCmd.PersistentFlags().StringSliceP("targets", "t", availableTargets, "Active targets to sync reading status with")
	rootCmd.PersistentFlags().StringP("source", "s", "database", "Active source to retrieve reading data from")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	viper.BindPFlag("targets", rootCmd.PersistentFlags().Lookup("targets"))
	viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configPath, err := xdg.SearchConfigFile(rootCmd.Name())
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
