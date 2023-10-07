package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/EricRibeiro/slack-orb-go/src/scripts/jsonutils"
	"github.com/EricRibeiro/slack-orb-go/src/scripts/stringutils"
)

func TestIsEventMatchingStatus(t *testing.T) {
	tests := []struct {
		jobStatus          string
		eventToSendMessage string
		result             bool
	}{
		{jobStatus: "pass", eventToSendMessage: "always", result: true},
		{jobStatus: "pass", eventToSendMessage: "pass", result: true},
		{jobStatus: "pass", eventToSendMessage: "fail", result: false},
		{jobStatus: "fail", eventToSendMessage: "always", result: true},
		{jobStatus: "fail", eventToSendMessage: "pass", result: false},
		{jobStatus: "fail", eventToSendMessage: "fail", result: true},
	}

	for _, test := range tests {
		result := stringutils.IsEventMatchingStatus(test.eventToSendMessage, test.jobStatus)
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
		result, err := stringutils.IsPatternMatchingString(test.patternStr, test.matchString)
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
		result := stringutils.IsPostConditionMet(test.branchMatches, test.tagMatches, test.invertMatch)
		if result != test.result {
			t.Errorf("For branchMatches: %v, tagMatches: %v, invertMatch: %v - expected %v, got %v", test.branchMatches, test.tagMatches, test.invertMatch, test.result, result)
		}
	}
}

func TestExpandEnvVarsInInterface(t *testing.T) {
	tests := []struct {
		input    interface{}
		envVars  map[string]string
		expected interface{}
	}{
		{
			input:    "Hello ${WORLD}",
			envVars:  map[string]string{"WORLD": "Earth"},
			expected: "Hello Earth",
		},
		{
			input: map[string]interface{}{
				"key": "value ${VAR}",
			},
			envVars: map[string]string{"VAR": "123"},
			expected: map[string]interface{}{
				"key": "value 123",
			},
		},
		{
			input: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value ${NESTED_VAR}",
				},
			},
			envVars: map[string]string{"NESTED_VAR": "456"},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value 456",
				},
			},
		},
	}

	for _, test := range tests {
		// Set environment variables
		for key, value := range test.envVars {
			os.Setenv(key, value)
		}

		result := jsonutils.ExpandEnvVarsInInterface(test.input)

		// Reset environment variables
		for key := range test.envVars {
			os.Unsetenv(key)
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For input: %+v, expected %+v, got %+v", test.input, test.expected, result)
		}
	}
}

func TestApplyFunctionToJSON(t *testing.T) {
	tests := []struct {
		messageBody string
		envVars     map[string]string
		expected    string
		hasError    bool
	}{
		{
			messageBody: `invalid json`,
			envVars:     map[string]string{"SOME_ENV": "expandedValue"},
			expected:    "",
			hasError:    true,
		},
		{
			messageBody: ``,
			envVars:     map[string]string{},
			expected:    ``,
			hasError:    false,
		},
		{
			messageBody: `{"key": "value", "number": 123, "array": [1, 2, 3], "bool": true}`,
			envVars:     map[string]string{},
			expected:    `{"key": "value", "number": 123, "array": [1, 2, 3], "bool": true}`,
			hasError:    false,
		},
		{
			messageBody: `{"key": "Hello ${WORLD}"}`,
			envVars:     map[string]string{"WORLD": "Earth"},
			expected:    `{"key": "Hello Earth"}`,
			hasError:    false,
		},
		{
			messageBody: `{"nested": {"key": "value ${NESTED_VAR}"}}`,
			envVars:     map[string]string{"NESTED_VAR": "456"},
			expected:    `{"nested": {"key": "value 456"}}`,
			hasError:    false,
		},
		{
			messageBody: `{"nestedDoubleQuotes": {"key": "${STRING_WITH_DOUBLE_QUOTES}"}}`,
			envVars:     map[string]string{"STRING_WITH_DOUBLE_QUOTES": `Do you prefer "tomato" or "potato"?`},
			expected:    `{"nestedDoubleQuotes": {"key": "Do you prefer \"tomato\" or \"potato\"?"}}`,
			hasError:    false,
		},
	}

	for _, test := range tests {
		// Set environment variables
		for key, value := range test.envVars {
			os.Setenv(key, value)
		}

		resultStr, err := jsonutils.ApplyFunctionToJSON(test.messageBody, jsonutils.ExpandEnvVarsInInterface)

		// Reset environment variables
		for key := range test.envVars {
			os.Unsetenv(key)
		}

		if test.hasError {
			if err == nil {
				t.Errorf("Expected an error for messageBody: %s", test.messageBody)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for messageBody: %s, error: %v", test.messageBody, err)
			continue
		}

		// Parse the result string into a map
		var resultMap map[string]interface{}
		if resultStr != "" {
			err = json.Unmarshal([]byte(resultStr), &resultMap)
			if err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				continue
			}
		} else {
			resultMap = nil
		}

		// Parse the expected string into a map
		var expectedMap map[string]interface{}
		if test.expected != "" {
			err = json.Unmarshal([]byte(test.expected), &expectedMap)
			if err != nil {
				t.Errorf("Failed to unmarshal expected result: %v", err)
				continue
			}
		} else {
			expectedMap = nil
		}

		// Compare the parsed structures
		if !reflect.DeepEqual(resultMap, expectedMap) {
			t.Errorf("For messageBody: %s, expected %+v, got %+v", test.messageBody, expectedMap, resultMap)
		}
	}
}

func TestInferTemplateEnvVarFromStatus(t *testing.T) {
	tests := []struct {
		jobStatus string
		expected  string
		hasError  bool
	}{
		// for the job status "success" the template name "basic_success_1" is returned
		{"success", "basic_success_1", false},
		// for the job status "fail" the template name "basic_fail_1" is returned
		{"fail", "basic_fail_1", false},
		// error because the job status is invalid.
		{"unknown", "", true},
	}

	for _, test := range tests {
		result, err := jsonutils.InferTemplateEnvVarFromStatus(test.jobStatus)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected an error for jobStatus: %s", test.jobStatus)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for jobStatus: %s, error: %v", test.jobStatus, err)
			continue
		}
		if result != test.expected {
			t.Errorf("For jobStatus: %s, expected %s, got %s", test.jobStatus, test.expected, result)
		}
	}
}

func TestDetermineMessageBody(t *testing.T) {
	// Set up mock environment variables for the test
	os.Setenv("basic_success_1", `{"text":"CircleCI job succeeded!","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Succeeded. :white_check_mark:","emoji":true}}]}`)
	os.Setenv("basic_fail_1", `{"text":"CircleCI job failed.","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Failed. :red_circle:","emoji":true}}]}`)

	tests := []struct {
		inlineTemplate           string
		jobStatus                string
		envVarContainingTemplate string
		expected                 string
		hasError                 bool
	}{
		// use custom message body
		{`{ "customMessageKey": "customMessageValue" }`, "success", "", `{ "customMessageKey": "customMessageValue" }`, false},
		// use basic_success_1 template because it was explicitly provided
		{"", "success", "basic_success_1", `{"text":"CircleCI job succeeded!","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Succeeded. :white_check_mark:","emoji":true}}]}`, false},
		// use basic_success_1 template because it was inferred from the job status
		{"", "success", "", `{"text":"CircleCI job succeeded!","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Succeeded. :white_check_mark:","emoji":true}}]}`, false},
		// use basic_fail_1 template because it was explicitly provided
		{"", "fail", "basic_fail_1", `{"text":"CircleCI job failed.","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Failed. :red_circle:","emoji":true}}]}`, false},
		// use basic_fail_1 template because it was inferred from the job status
		{"", "fail", "", `{"text":"CircleCI job failed.","blocks":[{"type":"header","text":{"type":"plain_text","text":"Job Failed. :red_circle:","emoji":true}}]}`, false},
		// error because the job status is invalid.
		{"", "unknown", "", "", true},
		// error because the template is empty
		{"", "success", "some_template_name", "", true},
	}

	for _, test := range tests {
		result, err := jsonutils.DetermineTemplate(test.inlineTemplate, test.jobStatus, test.envVarContainingTemplate)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected an error but got %s", result)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for %+v, error: %v", test, err)
			continue
		}
		if result != test.expected {
			t.Errorf("For %+v, got %s", test.inlineTemplate, result)
		}
	}

	// Clean up mock environment variables after the test
	os.Unsetenv("basic_success_1")
	os.Unsetenv("basic_fail_1")
}

func TestAddRootProperty(t *testing.T) {
	tests := []struct {
		messageBody string
		propertyKey string
		propertyVal interface{}
		expected    string
		hasError    bool
	}{
		{
			messageBody: `{"name": "John"}`,
			propertyKey: "city",
			propertyVal: "New York",
			expected:    `{"name": "John", "city": "New York"}`,
			hasError:    false,
		},
		{
			messageBody: `invalid json`,
			propertyKey: "city",
			propertyVal: "New York",
			expected:    "",
			hasError:    true,
		},
		{
			messageBody: `{}`,
			propertyKey: "city",
			propertyVal: "New York",
			expected:    `{"city": "New York"}`,
			hasError:    false,
		},
	}

	for _, test := range tests {
		resultStr, err := jsonutils.ApplyFunctionToJSON(test.messageBody, jsonutils.AddRootProperty(test.propertyKey, test.propertyVal))

		if test.hasError {
			if err == nil {
				t.Errorf("Expected an error for messageBody: %s", test.messageBody)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for messageBody: %s, error: %v", test.messageBody, err)
			continue
		}

		// Parse the result string into a map
		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		if err != nil {
			t.Errorf("Failed to unmarshal result: %v", err)
			continue
		}

		// Parse the expected string into a map
		var expectedMap map[string]interface{}
		err = json.Unmarshal([]byte(test.expected), &expectedMap)
		if err != nil {
			t.Errorf("Failed to unmarshal expected result: %v", err)
			continue
		}

		// Compare the parsed structures
		if !reflect.DeepEqual(resultMap, expectedMap) {
			t.Errorf("For messageBody: %s, expected %+v, got %+v", test.messageBody, expectedMap, resultMap)
		}
	}
}

func TestExtractRootProperty(t *testing.T) {
	tests := []struct {
		messageBody string
		propertyKey string
		expected    interface{}
		hasError    bool
	}{
		{
			messageBody: `{"name": "John", "city": "New York", "age": 30, "isStudent": false}`,
			propertyKey: "city",
			expected:    "New York",
			hasError:    false,
		},
		{
			messageBody: `{"name": "John", "city": "New York", "age": 30, "isStudent": false}`,
			propertyKey: "age",
			expected:    "30",
			hasError:    false,
		},
		{
			messageBody: `{"name": "John", "city": "New York", "age": 30, "isStudent": false}`,
			propertyKey: "isStudent",
			expected:    "false",
			hasError:    false,
		},
		{
			messageBody: `invalid json`,
			propertyKey: "city",
			expected:    "",
			hasError:    true,
		},
		{
			messageBody: `{}`,
			propertyKey: "city",
			expected:    "",
			hasError:    false,
		},
	}

	for _, test := range tests {
		result, err := jsonutils.ApplyFunctionToJSON(test.messageBody, jsonutils.ExtractRootProperty(test.propertyKey))

		if test.hasError {
			if err == nil {
				t.Errorf("Expected an error for messageBody: %s", test.messageBody)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error for messageBody: %s, error: %v", test.messageBody, err)
			continue
		}

		// Compare the results
		if result != test.expected {
			t.Errorf("For messageBody: %s, expected %v, got %v", test.messageBody, test.expected, result)
		}
	}
}
