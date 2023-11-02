package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/CircleCI-Public/slack-orb-go/src/scripts/config"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/httputils"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/jsonutils"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/stringutils"
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

	// Exit if the job status does not match the message send event and the message send event is not set to "always"
	if !stringutils.IsEventMatchingStatus(conf.EventToSendMessage, conf.JobStatus) {
		message := fmt.Sprintf("The job status %q does not match the status set to send alerts %q.", conf.JobStatus, conf.EventToSendMessage)
		fmt.Println(message)
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Check if the branch and tag match their respective patterns and parse the invert match parameter
	branchMatches, err := stringutils.IsPatternMatchingString(conf.BranchPattern, conf.JobBranch)
	if err != nil {
		log.Fatal("Error parsing the branch pattern:", err)
	}
	tagMatches, err := stringutils.IsPatternMatchingString(conf.TagPattern, conf.JobTag)
	if err != nil {
		log.Fatal("Error parsing the tag pattern:", err)
	}
	invertMatch, _ := strconv.ParseBool(conf.InvertMatchStr)
	if !stringutils.IsPostConditionMet(branchMatches, tagMatches, invertMatch) {
		fmt.Println("The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.")
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Build the message body
	template, err := jsonutils.DetermineTemplate(conf.InlineTemplate, conf.JobStatus, conf.EnvVarContainingTemplate)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if template == "" {
		log.Fatalf("the template %q is empty. Exiting without posting to Slack...", template)
	}

	// Expand environment variables in the template
	templateWithExpandedVars, err := jsonutils.ApplyFunctionToJSON(template, jsonutils.ExpandEnvVarsInInterface)
	if err != nil {
		log.Fatal(err)
	}

	// Add a "channel" property with a nested "myChannel" property
	modifiedJSON, err := jsonutils.ApplyFunctionToJSON(templateWithExpandedVars, jsonutils.AddRootProperty("channel", "my_channel"))
	if err != nil {
		log.Fatalf("%v", err)
	}

	ignoreErrors, _ := strconv.ParseBool(conf.IgnoreErrorsStr)
	channels := strings.Split(conf.ChannelsStr, ",")
	for _, channel := range channels {
		// Add a "channel" property with the current channel
		jsonWithChannel, err := jsonutils.ApplyFunctionToJSON(modifiedJSON, jsonutils.AddRootProperty("channel", channel))
		if err != nil {
			log.Fatalf("%v", err)
		}
		fmt.Printf("Posting the following JSON to Slack:\n%s\n", jsonutils.Colorize(jsonWithChannel))

		// Post the message to Slack
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + conf.AccessToken,
		}
		response, err := httputils.SendHTTPRequest("POST", "https://slack.com/api/chat.postMessage", jsonWithChannel, headers)
		if err != nil {
			log.Fatalf("Error posting to Slack: %v", err)
		}
		fmt.Printf("Slack API response:\n%s\n", jsonutils.Colorize(response))

		// Check if the Slack API returned an error message
		errorMsg, err := jsonutils.ApplyFunctionToJSON(response, jsonutils.ExtractRootProperty("error"))
		if err != nil {
			log.Fatalf("Error extracting error message: %v", err)
		}

		// Exit if the Slack API returned an error message and the ignore errors parameter is not set to true
		if errorMsg != "" {
			fmt.Printf("Slack API returned an error message:\n%s", errorMsg)
			fmt.Println("\n\nView the Setup Guide: https://github.com/CircleCI-Public/slack-orb/wiki/Setup")
			if !ignoreErrors {
				os.Exit(1)
			}
		}
	}
}
