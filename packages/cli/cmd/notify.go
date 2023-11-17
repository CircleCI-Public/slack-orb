package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		Status:         SlackConfig.JobStatus,
		Branch:         SlackConfig.JobBranch,
		Tag:            SlackConfig.JobTag,
		Event:          SlackConfig.EventToSendMessage,
		BranchPattern:  SlackConfig.BranchPattern,
		TagPattern:     SlackConfig.TagPattern,
		InvertMatch:    invertMatch,
		TemplateVar:    SlackConfig.TemplateVar,
		TemplatePath:   SlackConfig.TemplatePath,
		TemplateInline: SlackConfig.TemplateInline,
		TemplateName:   SlackConfig.TemplateName,

	}

	modifiedJSON, err := slackNotification.BuildMessageBody()
	if err != nil {
		if errors.Is(err, slack.ErrStatusMismatch) {
			//nolint:lll // user message
			fmt.Printf("Exiting without posting to Slack: The job status %q does not match the status set to send alerts %q.\n",
				slackNotification.Status, slackNotification.Event)
			os.Exit(0)
		} else if errors.Is(err, slack.ErrPostConditionNotMet) {
			//nolint:lll // user message
			fmt.Printf("Exiting without posting to Slack: The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.\n")
			os.Exit(0)
		}

		log.Fatalf("Failed to build message body: %v", err)
	}

	client := slack.NewClient(slack.ClientOptions{
		SlackToken: secret.String(SlackConfig.AccessToken),
		BaseURL:    SlackConfig.SlackAPIBaseUrl, // this is okay to set, it's ignored if the value is ""
	})

	for _, channel := range channels {
		fmt.Printf("Posting the following JSON to Slack:\n")
		colorizedJSONWitChannel, err := utils.ColorizeJSON(modifiedJSON)
		if err != nil {
			log.Fatalf("Error coloring JSON: %v", err)
		}
		fmt.Println(colorizedJSONWitChannel)
		err = client.PostMessage(context.Background(), modifiedJSON, channel)
		if err != nil {
			if !ignoreErrors {
				log.Fatalf("Error: \n%v\n", err)
			}

			fmt.Printf("Error: \n%v\n", err)
		} else {
			fmt.Println("Successfully posted message to channel: ", channel)
		}
	}

}
