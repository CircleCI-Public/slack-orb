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
	"github.com/buger/jsonparser"
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

func processKeyVal(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (interface{}, error) {
	switch dataType {
	case jsonparser.Object:
		// Recursively process nested objects
		resultObj := make(map[string]interface{})
		jsonparser.ObjectEach(value, func(k []byte, v []byte, dataType jsonparser.ValueType, offset int) error {
			processedVal, err := processKeyVal(k, v, dataType, offset)
			if err != nil {
				return err
			}
			resultObj[string(k)] = processedVal
			return nil
		})
		return resultObj, nil

	case jsonparser.Array:
		// Process array elements
		var resultArr []interface{}
		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				log.Fatal(err)
			}
			processedVal, err := processKeyVal(nil, value, dataType, offset)
			if err != nil {
				log.Fatal(err)
			}
			resultArr = append(resultArr, processedVal)
		})
		return resultArr, nil

	case jsonparser.String:
		// Process and escape the string value
		expandedString, err := envsubst.String(string(value))
		return expandedString, err

	default:
		// Other JSON types (Number, Boolean, Null)
		var result interface{}
		err := json.Unmarshal(value, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
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

	// Run message body through processKeyVal
	processedData, err := processKeyVal(nil, []byte(messageBody), jsonparser.Object, 0)
	if err != nil {
		log.Fatal(err)
	}	
	result, err := json.Marshal(processedData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(result))

	// Don't run message body through processKeyVal
	unprocessedData, _ := envsubst.String(messageBody)
	result, err = json.Marshal(unprocessedData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(result))
}
