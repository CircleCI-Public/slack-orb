package main

import (
	"context"
	"errors"
	"fmt"
	"log" //nolint:depguard // log is allowed for a top-level fatal
	"os"
	"strconv"
	"strings"

	"github.com/circleci/ex/config/secret"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/config"
	"github.com/CircleCI-Public/slack-orb-go/packages/cli/slack"
	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading environment configuration: %v", err)
	}

	if err := conf.Validate(); err != nil {
		handleConfigurationError(err)
	}

	invertMatch, _ := strconv.ParseBool(conf.InvertMatchStr)
	ignoreErrors, _ := strconv.ParseBool(conf.IgnoreErrorsStr)
	channels := strings.Split(conf.ChannelsStr, ",")

	slackNotification := slack.Notification{
		Status:         conf.JobStatus,
		Branch:         conf.JobBranch,
		Tag:            conf.JobTag,
		Event:          conf.EventToSendMessage,
		BranchPattern:  conf.BranchPattern,
		TagPattern:     conf.TagPattern,
		InvertMatch:    invertMatch,
		TemplateVar:    conf.TemplateVar,
		TemplatePath:   conf.TemplatePath,
		TemplateInline: conf.TemplateInline,
		TemplateName:   conf.TemplateName,
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
		SlackToken: secret.String(conf.AccessToken),
		BaseURL:    conf.SlackAPIBaseUrl, // this is okay to set, it's ignored if the value is ""
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
				log.Fatalf("Error: %v", err)
			}

			fmt.Printf("Error: %v", err)
		} else {
			fmt.Println("Successfully posted message to channel: ", channel)
		}
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
