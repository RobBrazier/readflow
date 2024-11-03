package cmd

import (
	"errors"
	"os"
	"time"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var cfgFile string
var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   internal.NAME,
	Short: "Track your Kobo reads on Anilist.co and Hardcover.app using Calibre-Web and Calibre databases",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
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

	log.SetTimeFormat(time.TimeOnly)
	log.SetLevel(log.InfoLevel)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", config.GetConfigPath(&cfgFile), "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := config.GetConfigPath(&cfgFile)
	err := config.LoadConfig(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warnf("Config file doesn't seem to exist! Please run `%s setup -c \"%s\"` to populate the configuration", internal.NAME, cfgFile)
		} else {
			log.Error("Unable to read config", "error", err)
			os.Exit(1)
		}
	}
}
