package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/fatih/color"
)

func TestAddRootProperty(t *testing.T) {
	tests := []struct {
		inputJSON map[string]interface{}
		key       string
		value     interface{}
		expected  map[string]interface{}
	}{
		{
			inputJSON: map[string]interface{}{"name": "John"},
			key:       "city",
			value:     "New York",
			expected:  map[string]interface{}{"name": "John", "city": "New York"},
		},
		{
			inputJSON: map[string]interface{}{},
			key:       "age",
			value:     30,
			expected:  map[string]interface{}{"age": 30},
		},
		{
			inputJSON: map[string]interface{}{"country": "USA"},
			key:       "isStudent",
			value:     false,
			expected:  map[string]interface{}{"country": "USA", "isStudent": false},
		},
		{
			inputJSON: map[string]interface{}{"name": "John", "city": "Los Angeles"},
			key:       "city",
			value:     "New York",
			expected:  map[string]interface{}{"name": "John", "city": "New York"}, // Overwrites existing key
		},
	}

	for _, test := range tests {
		modifier := AddRootProperty(test.key, test.value)
		result := modifier(test.inputJSON).(map[string]interface{})
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %+v, got %+v", test.expected, result)
		}
	}
}

func TestExtractRootProperty(t *testing.T) {
	tests := []struct {
		inputJSON map[string]interface{}
		key       string
		expected  interface{}
	}{
		{
			inputJSON: map[string]interface{}{"name": "John", "city": "New York"},
			key:       "city",
			expected:  "New York",
		},
		{
			inputJSON: map[string]interface{}{"age": 30},
			key:       "age",
			expected:  30,
		},
		{
			inputJSON: map[string]interface{}{"isStudent": false},
			key:       "isStudent",
			expected:  false,
		},
		{
			inputJSON: map[string]interface{}{"name": "John", "city": "Los Angeles"},
			key:       "country",
			expected:  "", // Key doesn't exist
		},
		{
			inputJSON: map[string]interface{}{},
			key:       "city",
			expected:  "", // Empty JSON
		},
	}

	for _, test := range tests {
		modifier := ExtractRootProperty(test.key)
		result := modifier(test.inputJSON)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %+v, got %+v", test.expected, result)
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

		result := ExpandEnvVarsInInterface(test.input)

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

		resultStr, err := ApplyFunctionToJSON(test.messageBody, ExpandEnvVarsInInterface)

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

func TestSpecialCharsInTemplate(t *testing.T) {
	tests := []struct {
		messageBody string
		envVars     map[string]string
		expected    string
		hasError    bool
	}{
		{
			messageBody: `{"multiline": "$MULTILINE_STRING", "quotes": "$NESTED_QUOTES"}`,
			envVars: map[string]string{
				"MULTILINE_STRING": "This is a string\nwith multiple\nlines.",
				"NESTED_QUOTES":    `This is a "quote inside" a quote.`,
			},
			expected: `{"multiline": "This is a string\nwith multiple\nlines.", "quotes": "This is a \"quote inside\" a quote."}`,
			hasError: false,
		},
		{
			messageBody: `{"specialChars": "$SPECIAL_CHARS"}`,
			envVars: map[string]string{
				"SPECIAL_CHARS": `\b\f\n\r\t`,
			},
			expected: `{"specialChars": "\\b\\f\\n\\r\\t"}`,
			hasError: false,
		},
		{
			messageBody: `{"invalidJSON": "$INVALID_JSON"}`,
			envVars: map[string]string{
				"INVALID_JSON": `{invalid: json}`,
			},
			expected: `{"invalidJSON": "{invalid: json}"}`,
			hasError: false,
		},
	}

	for _, test := range tests {
		// Set environment variables
		for key, value := range test.envVars {
			os.Setenv(key, value)
		}

		resultStr, err := ApplyFunctionToJSON(test.messageBody, ExpandEnvVarsInInterface)

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

		// Check if the result is a valid JSON
		if !json.Valid([]byte(resultStr)) {
			t.Errorf("The result is not a valid JSON for messageBody: %s", test.messageBody)
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

func TestColorize(t *testing.T) {
	testJSON := `{ "key": "value" }`
	color.NoColor = false
	colored, err := ColorizeJSON(testJSON)
	if err != nil {
		t.Errorf("Error colorizing JSON: %v", err)
	}
	expectedBytes := []byte{
		123, 10, 32, 32, // `{`, newline, space, space
		27, 91, 51, 55, 109, // ESC [ 37 m (white color)
		34, 107, 101, 121, 34, // "key"
		58, 32, // `:`, space
		27, 91, 48, 109, // ESC [ 0 m (reset color)
		27, 91, 51, 50, 109, // ESC [ 32 m (green color)
		34, 118, 97, 108, 117, 101, 34, // "value"
		27, 91, 48, 109, // ESC [ 0 m (reset color)
		10, 125, // newline, `}`
	}
	if !bytes.Equal(expectedBytes, []byte(colored)) {
		t.Errorf("Colorized JSON did not match expected.\nExpected: %s\nGot: %s", expectedBytes, colored)
	}
}
