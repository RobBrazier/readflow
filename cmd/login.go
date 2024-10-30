package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/RobBrazier/readflow/target"
	"github.com/cli/browser"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login latest reading states to configured targets",
	Run: func(cmd *cobra.Command, args []string) {
		activeTargets := target.GetActiveTargets()
		slog.Info("login called with", "targets", activeTargets)
		for _, target := range activeTargets {
			name := target.GetName()
			if target.HasToken() {
				confirmPrompt := promptui.Prompt{
					Label:     fmt.Sprintf("Token already exists in config for %s, Re-authenticate", name),
					IsConfirm: true,
				}
				_, err := confirmPrompt.Run()
				if err != nil {
					slog.Info("Skipping token update for", "target", name)
					continue
				}
			}
			slog.Info("Setting authentication token for", "target", name)
			url, err := target.Login()
			cobra.CheckErr(err)
			slog.Info(fmt.Sprintf("Please open the following URL in your browser if it hasn't already opened: %s", url))
			browser.OpenURL(url)
			slog.Info("Please login and paste the token shown on the website below")
			token := getTokenInput()
			target.SaveToken(token)
		}
		viper.WriteConfig()
	},
}

func getTokenInput() string {
	fmt.Print("> ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	token := input.Text()
	return token
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
