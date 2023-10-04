package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

func ExecCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("command exited with error: %w\nstderr: %s", exitError, string(exitError.Stderr))
		}
		return "", fmt.Errorf("failed to run command: %w", err)
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func IsPatternMatchingString(patternStr string, matchString string) (bool, error) {
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		return false, fmt.Errorf("error compiling pattern %q: %w", patternStr, err)
	}
	return pattern.MatchString(matchString), nil
}

func IsEventMatchingStatus(messageSendEvent string, jobStatus string) bool {
	return jobStatus == messageSendEvent || messageSendEvent == "always"
}

func IsPostConditionMet(branchMatches bool, tagMatches bool, invertMatch bool) bool {
	return (branchMatches || tagMatches) != invertMatch
}

func main() {
	// Fetch environment variables
	branchPatternStr := os.Getenv("SLACK_PARAM_BRANCHPATTERN")
	invertMatchStr := os.Getenv("SLACK_PARAM_INVERT_MATCH")
	jobBranch := os.Getenv("CIRCLE_BRANCH")
	jobStatus := os.Getenv("CCI_STATUS")
	jobTag := os.Getenv("CIRCLE_TAG")
	messageSendEvent := os.Getenv("SLACK_PARAM_EVENT")
	tagPatternStr := os.Getenv("SLACK_PARAM_TAGPATTERN")

	// Expand environment variables
	branchPatternStr = os.ExpandEnv(branchPatternStr)
	invertMatchStr = os.ExpandEnv(invertMatchStr)
	jobBranch = os.ExpandEnv(jobBranch)
	jobStatus = os.ExpandEnv(jobStatus)
	jobTag = os.ExpandEnv(jobTag)
	messageSendEvent = os.ExpandEnv(messageSendEvent)
	tagPatternStr = os.ExpandEnv(tagPatternStr)

	// Exit if the job status does not match the message send event and the message send event is not set to "always"
	if !IsEventMatchingStatus(messageSendEvent, jobStatus) {
		message := fmt.Sprintf("The job status %q does not match the status set to send alerts %q.", jobStatus, messageSendEvent)
		fmt.Println(message)
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Parse the environment variables to proper types
	branchMatches, branchMatchesErr := IsPatternMatchingString(branchPatternStr, jobBranch)
	if branchMatchesErr != nil {
		fmt.Println("Error parsing the branch pattern:", branchMatchesErr)
		fmt.Println("Please check the branch pattern in the parameters and try again.")
		os.Exit(1)
	}
	tagMatches, tagMatchesErr := IsPatternMatchingString(tagPatternStr, jobTag)
	if tagMatchesErr != nil {
		fmt.Println("Error parsing the tag pattern:", tagMatchesErr)
		fmt.Println("Please check the tag pattern in the parameters and try again.")
		os.Exit(1)
	}
	invertMatch, _ := strconv.ParseBool(invertMatchStr)

	if !IsPostConditionMet(branchMatches, tagMatches, invertMatch) {
		fmt.Println("The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.")
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Load the environment variables from the configuration file
	bashEnv := os.Getenv("BASH_ENV")
	if FileExists(bashEnv) {
		fmt.Println("Loading BASH_ENV into the environment...")
		if err := godotenv.Load(bashEnv); err != nil {
			fmt.Println("Error loading BASH_ENV file:", err)
			fmt.Println("Please check the BASH_ENV variable and try again.")
			os.Exit(1)
		}
	}

	// Build the message body
	customMessageBodyStr := os.Getenv("SLACK_PARAM_CUSTOM")
	customMessageBody := os.ExpandEnv(customMessageBodyStr)

	if customMessageBody != "" {
		fmt.Println("Sending custom message to Slack...")
	} else {
		templateNameStr := os.Getenv("SLACK_PARAM_TEMPLATE")
		templateName := os.ExpandEnv(templateNameStr)
		template := os.Getenv(templateName)
		if template == "" {
			fmt.Println("Error loading the template:", templateName)
			fmt.Println("Please check the template name and try again.")
			os.Exit(1)
		}

		fmt.Println("Expanding the template...")
		expandedTemplate := os.ExpandEnv(template)
		fmt.Println("Template:", expandedTemplate)
	}
}
