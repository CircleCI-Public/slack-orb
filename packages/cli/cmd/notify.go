package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"os"
	"strconv"
	"strings"

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

func executeNotify(cmd *cobra.Command, args []string) {
	invertMatch, _ := strconv.ParseBool(config.SlackConfig.InvertMatchStr)
	ignoreErrors, _ := strconv.ParseBool(config.SlackConfig.IgnoreErrorsStr)
	channels := strings.Split(config.SlackConfig.ChannelsStr, ",")

	slackNotification := slack.Notification{
		Status:         config.SlackConfig.JobStatus,
		Branch:         config.SlackConfig.JobBranch,
		Tag:            config.SlackConfig.JobTag,
		Event:          config.SlackConfig.EventToSendMessage,
		BranchPattern:  config.SlackConfig.BranchPattern,
		TagPattern:     config.SlackConfig.TagPattern,
		InvertMatch:    invertMatch,
		TemplateVar:    config.SlackConfig.TemplateVar,
		TemplatePath:   config.SlackConfig.TemplatePath,
		TemplateInline: config.SlackConfig.TemplateInline,
		TemplateName:   config.SlackConfig.TemplateName,
	}

	modifiedJSON, err := slackNotification.BuildMessageBody()
	if err != nil {
		if errors.Is(err, slack.ErrStatusMismatch) {
			log.Fatalf("Exiting without posting to Slack: The job status %q does not match the status set to send alerts %q.\n",
				slackNotification.Status, slackNotification.Event)
		} else if errors.Is(err, slack.ErrPostConditionNotMet) {
			log.Fatalf("Exiting without posting to Slack: The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.\n")
		}

		log.Fatalf("Failed to build message body: %v", err)
	}

	client := slack.NewClient(slack.ClientOptions{
		SlackToken: secret.String(config.SlackConfig.AccessToken),
		BaseURL:    config.SlackConfig.SlackAPIBaseUrl, // this is okay to set, it's ignored if the value is ""
	})

	for _, channel := range channels {
		log.Debugf("Posting the following JSON to Slack:\n")
		colorizedJSONWitChannel, err := utils.ColorizeJSON(modifiedJSON)
		if err != nil {
			log.Fatalf("Error coloring JSON: %v", err)
		}
		log.Debug(colorizedJSONWitChannel)
		err = client.PostMessage(context.Background(), modifiedJSON, channel)
		if err != nil {
			if !ignoreErrors {
				log.Fatalf("Error: \n%v\n", err)
			}

			log.Errorf("Error: \n%v\n", err)
		} else {
			log.Info("Successfully posted message to channel: ", channel)
		}
	}

}
