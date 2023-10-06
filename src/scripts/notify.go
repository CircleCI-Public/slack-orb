package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/EricRibeiro/slack-orb-go/src/scripts/ioutils"
	"github.com/EricRibeiro/slack-orb-go/src/scripts/jsonutils"
	"github.com/EricRibeiro/slack-orb-go/src/scripts/stringutils"
	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"
)

func main() {
	// Load the environment variables from the configuration file
	// This has to be done before anything else to ensure that the environment variables modified by the configuration file are available
	bashEnv := os.Getenv("BASH_ENV")
	if ioutils.FileExists(bashEnv) {
		fmt.Println("Loading BASH_ENV into the environment...")
		if err := godotenv.Load(bashEnv); err != nil {
			log.Fatal("Error loading BASH_ENV file:", err)
		}
	}

	// Fetch environment variables
	accessToken := os.Getenv("SLACK_ACCESS_TOKEN")
	branchPattern := os.Getenv("SLACK_PARAM_BRANCHPATTERN")
	channels := os.Getenv("SLACK_PARAM_CHANNEL")
	envVarContainingTemplate := os.Getenv("SLACK_PARAM_TEMPLATE")
	eventToSendMessage := os.Getenv("SLACK_PARAM_EVENT")
	inlineTemplate := os.Getenv("SLACK_PARAM_CUSTOM")
	invertMatchStr := os.Getenv("SLACK_PARAM_INVERT_MATCH")
	isDebugStr := os.Getenv("SLACK_PARAM_DEBUG")
	jobBranch := os.Getenv("CIRCLE_BRANCH")
	jobStatus := os.Getenv("CCI_STATUS")
	jobTag := os.Getenv("CIRCLE_TAG")
	tagPattern := os.Getenv("SLACK_PARAM_TAGPATTERN")

	// Expand environment variables
	accessToken, _ = envsubst.String(accessToken)
	branchPattern, _ = envsubst.String(branchPattern)
	channels, _ = envsubst.String(channels)
	envVarContainingTemplate, _ = envsubst.String(envVarContainingTemplate)
	eventToSendMessage, _ = envsubst.String(eventToSendMessage)
	inlineTemplate, _ = envsubst.String(inlineTemplate)
	invertMatchStr, _ = envsubst.String(invertMatchStr)
	isDebugStr, _ = envsubst.String(isDebugStr)
	tagPattern, _ = envsubst.String(tagPattern)

	// Exit if required environment variables are not set
	if accessToken == "" {
		log.Fatalf(
			"In order to use the Slack Orb (v4 +), an OAuth token must be present via the SLACK_ACCESS_TOKEN environment variable." +
				"\nFollow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup.",
		)
	}
	if channels == "" {
		log.Fatalf(
			`No channel was provided. Please provide one or more channels using the "SLACK_PARAM_CHANNEL" environment variable or the "channel" parameter.`,
		)
	}

	// Exit if the job status does not match the message send event and the message send event is not set to "always"
	if !stringutils.IsEventMatchingStatus(eventToSendMessage, jobStatus) {
		message := fmt.Sprintf("The job status %q does not match the status set to send alerts %q.", jobStatus, eventToSendMessage)
		fmt.Println(message)
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Check if the branch and tag match their respective patterns and parse the invert match parameter
	branchMatches, err := stringutils.IsPatternMatchingString(branchPattern, jobBranch)
	if err != nil {
		log.Fatal("Error parsing the branch pattern:", err)
	}
	tagMatches, err := stringutils.IsPatternMatchingString(tagPattern, jobTag)
	if err != nil {
		log.Fatal("Error parsing the tag pattern:", err)
	}
	invertMatch, _ := strconv.ParseBool(invertMatchStr)
	if !stringutils.IsPostConditionMet(branchMatches, tagMatches, invertMatch) {
		fmt.Println("The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.")
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Build the message body
	template, err := jsonutils.DetermineTemplate(inlineTemplate, jobStatus, envVarContainingTemplate)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if template == "" {
		log.Fatalf("the template %q is empty. Exiting without posting to Slack...", template)
	}

	templateWithExpandedVars, err := jsonutils.ApplyFunctionToJSON(template, jsonutils.ExpandEnvVarsInInterface)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(templateWithExpandedVars)
}
