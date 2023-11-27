package cmd

import (
	"errors"
	"os"

	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slack-orb-cli",
	Short: "The slack-orb-cli interface for the CircleCI Slack orb",
	Long:  `The slack-orb-cli by CircleCI is a command-line tool for sending slack notifications as a part of a CI/CD workflow.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debugFlag, err := cmd.Flags().GetBool("debug")
		if err != nil {
			// Handle the error
			log.Fatalf("Error accessing debug flag: %v", err)
		}
		if debugFlag {
			os.Setenv("SLACK_PARAM_DEBUG", "true")
		}
		debugValue := config.GetDebug()
		if debugValue {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug logging enabled")
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

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

}

func initConfig() {
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("Error loading environment configuration: \n%v\n", err)
	}
	if err := config.SlackConfig.Validate(); err != nil {
		handleConfigurationError(err)
	}
}
func handleConfigurationError(err error) {
	var envVarError *config.EnvVarError
	if errors.As(err, &envVarError) {
		switch envVarError.VarName {
		case "SLACK_ACCESS_TOKEN":
			log.Fatalf(
				"In order to use the Slack Orb an OAuth token must be present via the SLACK_ACCESS_TOKEN environment variable." +
					"\nFollow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup.",
			)
		case "SLACK_PARAM_CHANNEL":
			//nolint:lll // user message
			log.Fatalf(
				`No channel was provided. Please provide one or more channels using the "SLACK_PARAM_CHANNEL" environment variable or the "channel" parameter.`,
			)
		default:
			log.Fatalf("Configuration validation failed: Environment variable not set: %s", envVarError.VarName)
		}
	}

	log.Fatal(err)
}
