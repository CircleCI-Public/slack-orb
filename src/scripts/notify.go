package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

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

func ExpandAndMarshalJSON(messageBody string) (string, error) {
	if messageBody == "" {
		return "", nil
	}

	var jsonTemplate map[string]interface{}
	err := json.Unmarshal([]byte(messageBody), &jsonTemplate)
	if err != nil {
		return "", err
	}

	expandedTemplate := expandEnvVarsInInterface(jsonTemplate).(map[string]interface{})

	result, err := json.Marshal(expandedTemplate)
	if err != nil {
		return "", err
	}

	return string(result), nil
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
	branchPatternStr, _ = envsubst.String(branchPatternStr)
	invertMatchStr, _ = envsubst.String(invertMatchStr)
	jobBranch, _ = envsubst.String(jobBranch)
	jobStatus, _ = envsubst.String(jobStatus)
	jobTag, _ = envsubst.String(jobTag)
	messageSendEvent, _ = envsubst.String(messageSendEvent)
	tagPatternStr, _ = envsubst.String(tagPatternStr)

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
		log.Fatal("Error parsing the branch pattern:", branchMatchesErr)
	}
	tagMatches, tagMatchesErr := IsPatternMatchingString(tagPatternStr, jobTag)
	if tagMatchesErr != nil {
		log.Fatal("Error parsing the tag pattern:", tagMatchesErr)
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
			log.Fatal("Error loading BASH_ENV file:", err)
		}
	}

	// Build the message body
	customMessageBodyStr := os.Getenv("SLACK_PARAM_CUSTOM")
	customMessageBody, _ := envsubst.String(customMessageBodyStr)
	messageBody := ""

	if customMessageBody != "" {
		messageBody = customMessageBody
	} else {
		templateNameStr := os.Getenv("SLACK_PARAM_TEMPLATE")
		templateName, _ := envsubst.String(templateNameStr)
		messageBody = os.Getenv(templateName)
	}

	result, err := ExpandAndMarshalJSON(messageBody)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
