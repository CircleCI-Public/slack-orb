package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/circleci/ex/config/secret"

	"github.com/CircleCI-Public/slack-orb-go/src/scripts/config"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/jsonutils"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/slack"
)

func main() {
	// Load environment variables from BASH_ENV and SLACK_JOB_STATUS files
	// This has to be done before loading the configuration because the configuration
	// depends on the environment variables loaded from these files
	if err := config.LoadEnvFromFile(os.Getenv("BASH_ENV")); err != nil {
		log.Fatal(err)
	}
	if err := config.LoadEnvFromFile("/tmp/SLACK_JOB_STATUS"); err != nil {
		log.Fatal(err)
	}

	conf := config.NewConfig()

	if err := conf.ExpandEnvVariables(); err != nil {
		log.Fatalf("Error expanding environment variables: %v", err)
	}

	if err := conf.Validate(); err != nil {
		if envVarError, ok := err.(*config.EnvVarError); ok {
			switch envVarError.VarName {
			case "SLACK_ACCESS_TOKEN":
				log.Fatalf(
					"In order to use the Slack Orb an OAuth token must be present via the SLACK_ACCESS_TOKEN environment variable." +
						"\nFollow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup.",
				)
			case "SLACK_PARAM_CHANNEL":
				log.Fatalf(
					`No channel was provided. Please provide one or more channels using the "SLACK_PARAM_CHANNEL" environment variable or the "channel" parameter.`,
				)
			default:
				log.Fatalf("Configuration validation failed: Environment variable not set: %s", envVarError.VarName)
			}
		} else {
			log.Fatalf("Configuration validation failed: %v", err)
		}
	}

	invertMatch, _ := strconv.ParseBool(conf.InvertMatchStr)
	ignoreErrors, _ := strconv.ParseBool(conf.IgnoreErrorsStr)
	channels := strings.Split(conf.ChannelsStr, ",")

	slackNotification := slack.Notification{
		Status:                   conf.JobStatus,
		Branch:                   conf.JobBranch,
		Tag:                      conf.JobTag,
		Event:                    conf.EventToSendMessage,
		BranchPattern:            conf.BranchPattern,
		TagPattern:               conf.TagPattern,
		InvertMatch:              invertMatch,
		InlineTemplate:           conf.InlineTemplate,
		EnvVarContainingTemplate: conf.EnvVarContainingTemplate,
	}

	modifiedJSON, err := slackNotification.BuildMessageBody()
	if err != nil {
		if strings.HasPrefix(err.Error(), "Exiting without posting to Slack") {
			fmt.Println(err)
			os.Exit(0)
		} else {
			log.Fatalf("Failed to build message body: %v", err)
		}
	}

	client := slack.NewClient(slack.ClientOptions{SlackToken: secret.String(conf.AccessToken)})

	for _, channel := range channels {
		fmt.Printf("Posting the following JSON to Slack:\n")
		colorizedJSONWitChannel, err := jsonutils.Colorize(modifiedJSON)
		if err != nil {
			log.Fatalf("Error coloring JSON: %v", err)
		}
		fmt.Println(colorizedJSONWitChannel)
		err = client.PostMessage(context.Background(), modifiedJSON, channel)
		if err != nil {
			if !ignoreErrors {
				log.Fatalf("Error: %v", err)
			} else {
				fmt.Printf("Error: %v", err)
			}
		} else {
			fmt.Println("Successfully posted message to channel: ", channel)
		}
	}
}
