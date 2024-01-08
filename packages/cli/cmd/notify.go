package cmd

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/circleci/ex/config/secret"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/config"
	"github.com/CircleCI-Public/slack-orb-go/packages/cli/slack"
	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

// notifyCmd represents the notify command
var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Send a slack notification",
	Long:  `Send a custom notification to slack`,
	Run:   executeNotify,
}

func init() {
	rootCmd.AddCommand(notifyCmd)
}

func executeNotify(_ *cobra.Command, _ []string) {
	cfg := config.SlackConfig
	channels := strings.Split(cfg.Channels, ",")

	invertMatch, _ := strconv.ParseBool(cfg.InvertMatch) // will default to false on a parse error
	ignoreErrors, _ := strconv.ParseBool(cfg.IgnoreErrors)

	slackNotification := slack.Notification{
		Status:         cfg.JobStatus,
		Branch:         cfg.JobBranch,
		Tag:            cfg.JobTag,
		Event:          cfg.EventToSendMessage,
		BranchPattern:  cfg.BranchPattern,
		TagPattern:     cfg.TagPattern,
		InvertMatch:    invertMatch,
		TemplateVar:    cfg.TemplateVar,
		TemplatePath:   cfg.TemplatePath,
		TemplateInline: cfg.TemplateInline,
		TemplateName:   cfg.TemplateName,
	}

	modifiedJSON, err := slackNotification.BuildMessageBody()
	if err != nil {
		if errors.Is(err, slack.ErrStatusMismatch) {
			log.Infof("Exiting without posting to Slack: The job status %q does not match the status set to send alerts %q.\n",
				slackNotification.Status, slackNotification.Event)
			os.Exit(0)
		} else if errors.Is(err, slack.ErrPostConditionNotMet) {
			log.Infof("Exiting without posting to Slack: The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.\n")
			os.Exit(0)
		}

		log.Fatalf("Failed to build message body: %v", err)
	}

	client := slack.NewClient(slack.ClientOptions{
		SlackToken: secret.String(cfg.AccessToken),
		BaseURL:    cfg.SlackAPIBaseUrl, // this is okay to set, it's ignored if the value is ""
	})

	for _, channel := range channels {
		log.Debugf("Posting the following JSON to Slack:\n")
		colorizedJSONWithChannel, err := utils.ColorizeJSON(modifiedJSON)
		if err != nil {
			log.Fatalf("Error coloring JSON: %v", err)
		}
		log.Debug(colorizedJSONWithChannel)
		err = client.PostMessage(context.Background(), modifiedJSON, channel)
		if err != nil {
			if !ignoreErrors {
				log.Fatalf("Error: \n%v\n", err)
			}

			log.Errorf("Error: \n%v\n", err)
		} else {
			log.Infof("Successfully posted message to channel: %s", channel)
		}
	}

}
