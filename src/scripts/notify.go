package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

func IsPatternMatchingString(patternStr string, matchString string) (bool, error) {
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		return false, fmt.Errorf("error compiling pattern %q: %w", patternStr, err)
	}
	return pattern.MatchString(matchString), nil
}

func IsEventMatchingStatus(eventToSendMessage string, jobStatus string) bool {
	return jobStatus == eventToSendMessage || eventToSendMessage == "always"
}

func IsPostConditionMet(branchMatches bool, tagMatches bool, invertMatch bool) bool {
	return (branchMatches || tagMatches) != invertMatch
}

func expandEnvVarsInInterface(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		expandedValue, _ := envsubst.String(v)
		return expandedValue
	case map[string]interface{}:
		for key, innerValue := range v {
			v[key] = expandEnvVarsInInterface(innerValue)
		}
	case []interface{}:
		for i, innerValue := range v {
			v[i] = expandEnvVarsInInterface(innerValue)
		}
	}
	return value
}

func ApplyFunctionToJSON(messageBody string, modifier func(interface{}) interface{}) (string, error) {
	if messageBody == "" {
		return "", nil
	}

	var jsonTemplate map[string]interface{}
	err := json.Unmarshal([]byte(messageBody), &jsonTemplate)
	if err != nil {
		return "", err
	}

	modifiedTemplate := modifier(jsonTemplate).(map[string]interface{})

	result, err := json.Marshal(modifiedTemplate)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func inferTemplateEnvVarFromStatus(jobStatus string) (string, error) {
	switch jobStatus {
	case "success":
		return "basic_success_1", nil
	case "fail":
		return "basic_fail_1", nil
	default:
		return "", fmt.Errorf("the job status: %q is unexpected", jobStatus)
	}
}

func determineTemplate(inlineTemplate, jobStatus, envVarContainingTemplate string) (string, error) {
	if inlineTemplate != "" {
		return inlineTemplate, nil
	}

	if envVarContainingTemplate == "" {
		var err error
		envVarContainingTemplate, err = inferTemplateEnvVarFromStatus(jobStatus)
		if err != nil {
			return "", err
		}
	}
	template := os.Getenv(envVarContainingTemplate)
	if template == "" {
		return "", fmt.Errorf("the template %q is empty", template)
	}
	return template, nil
}

func main() {
	// Load the environment variables from the configuration file
	// This has to be done before anything else to ensure that the environment variables modified by the configuration file are available
	bashEnv := os.Getenv("BASH_ENV")
	if FileExists(bashEnv) {
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
	inlineTemplate := os.Getenv("SLACK_PARAM_CUSTOM")
	invertMatchStr := os.Getenv("SLACK_PARAM_INVERT_MATCH")
	jobBranch := os.Getenv("CIRCLE_BRANCH")
	jobStatus := os.Getenv("CCI_STATUS")
	jobTag := os.Getenv("CIRCLE_TAG")
	eventToSendMessage := os.Getenv("SLACK_PARAM_EVENT")
	tagPattern := os.Getenv("SLACK_PARAM_TAGPATTERN")

	// Expand environment variables
	accessToken, _ = envsubst.String(accessToken)
	branchPattern, _ = envsubst.String(branchPattern)
	channels, _ = envsubst.String(channels)
	envVarContainingTemplate, _ = envsubst.String(envVarContainingTemplate)
	inlineTemplate, _ = envsubst.String(inlineTemplate)
	invertMatchStr, _ = envsubst.String(invertMatchStr)
	eventToSendMessage, _ = envsubst.String(eventToSendMessage)
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
	if !IsEventMatchingStatus(eventToSendMessage, jobStatus) {
		message := fmt.Sprintf("The job status %q does not match the status set to send alerts %q.", jobStatus, eventToSendMessage)
		fmt.Println(message)
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Check if the branch and tag match their respective patterns and parse the invert match parameter
	branchMatches, err := IsPatternMatchingString(branchPattern, jobBranch)
	if err != nil {
		log.Fatal("Error parsing the branch pattern:", err)
	}
	tagMatches, err := IsPatternMatchingString(tagPattern, jobTag)
	if err != nil {
		log.Fatal("Error parsing the tag pattern:", err)
	}
	invertMatch, _ := strconv.ParseBool(invertMatchStr)
	if !IsPostConditionMet(branchMatches, tagMatches, invertMatch) {
		fmt.Println("The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.")
		fmt.Println("Exiting without posting to Slack...")
		os.Exit(0)
	}

	// Build the message body
	template, err := determineTemplate(inlineTemplate, jobStatus, envVarContainingTemplate)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if template == "" {
		log.Fatalf("the template %q is empty. Exiting without posting to Slack...", template)
	}

	templateWithExpandedVars, err := ApplyFunctionToJSON(template, expandEnvVarsInInterface)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(templateWithExpandedVars)
}
