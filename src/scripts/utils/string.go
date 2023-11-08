package utils

import (
	"fmt"
	"regexp"
)

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
