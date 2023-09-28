package main

import "testing"

func TestIsEventMatchingStatus(t *testing.T) {
	tests := []struct {
		jobStatus        string
		messageSendEvent string
		result           bool
	}{
		{jobStatus: "pass", messageSendEvent: "always", result: true},
		{jobStatus: "pass", messageSendEvent: "pass", result: true},
		{jobStatus: "pass", messageSendEvent: "fail", result: false},
		{jobStatus: "fail", messageSendEvent: "always", result: true},
		{jobStatus: "fail", messageSendEvent: "pass", result: false},
		{jobStatus: "fail", messageSendEvent: "fail", result: true},
	}

	for _, test := range tests {
		result := IsEventMatchingStatus(test.messageSendEvent, test.jobStatus)
		if result != test.result {
			t.Errorf("Expected %v, got %v", test.result, result)
		}
	}
}

func TestIsPatternMatchingString(t *testing.T) {
	tests := []struct {
		patternStr  string
		matchString string
		result      bool
	}{
		{patternStr: ".*", matchString: "myBranchName", result: true},
		{patternStr: ".*", matchString: "myTagName", result: true},
		{patternStr: "thisVerySpecificBranchName", matchString: "myBranchName", result: false},
		{patternStr: "thisVerySpecificBranchName", matchString: "thisVerySpecificBranchName", result: true},
		{patternStr: "thisVerySpecificTagName", matchString: "myTagName", result: false},
		{patternStr: "thisVerySpecificTagName", matchString: "thisVerySpecificTagName", result: true},
		{patternStr: "", matchString: "", result: true},                     // both empty
		{patternStr: "", matchString: "notEmpty", result: true},             // pattern empty, match string not empty
		{patternStr: "notEmpty", matchString: "", result: false},            // pattern not empty, match string empty
		{patternStr: "^[a-z]+$", matchString: "alllowercase", result: true}, // character class
		{patternStr: "^[a-zA-Z]+$", matchString: "MixEdCaSe", result: true}, // character class with upper and lower case
		{patternStr: "^[0-9]+$", matchString: "12345", result: true},        // numeric values
		{patternStr: "^\\d{2,4}$", matchString: "123", result: true},        // quantifier
		{patternStr: "apple|orange", matchString: "apple", result: true},    // alternation
		{patternStr: "apple|orange", matchString: "banana", result: false},
		{patternStr: "^a.c$", matchString: "abc", result: true}, // dot special character
		{patternStr: "^a.c$", matchString: "abbc", result: false},
	}

	for _, test := range tests {
		result, err := IsPatternMatchingString(test.patternStr, test.matchString)
		if err != nil {
			t.Errorf("For pattern %q and matchString %q, unexpected error: %v", test.patternStr, test.matchString, err)
		}
		if result != test.result {
			t.Errorf("For pattern %q and matchString %q, expected %v, got %v", test.patternStr, test.matchString, test.result, result)
		}
	}
}

func TestIsPostConditionMet(t *testing.T) {
	tests := []struct {
		branchMatches bool
		tagMatches    bool
		invertMatch   bool
		result        bool
	}{
		{branchMatches: true, tagMatches: true, invertMatch: false, result: true},
		{branchMatches: true, tagMatches: true, invertMatch: true, result: false},
		{branchMatches: true, tagMatches: false, invertMatch: false, result: true},
		{branchMatches: true, tagMatches: false, invertMatch: true, result: false},
		{branchMatches: false, tagMatches: true, invertMatch: false, result: true},
		{branchMatches: false, tagMatches: true, invertMatch: true, result: false},
		{branchMatches: false, tagMatches: false, invertMatch: false, result: false},
		{branchMatches: false, tagMatches: false, invertMatch: true, result: true},
	}

	for _, test := range tests {
		result := IsPostConditionMet(test.branchMatches, test.tagMatches, test.invertMatch)
		if result != test.result {
			t.Errorf("For branchMatches: %v, tagMatches: %v, invertMatch: %v - expected %v, got %v", test.branchMatches, test.tagMatches, test.invertMatch, test.result, result)
		}
	}
}
