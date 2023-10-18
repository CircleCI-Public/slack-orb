package config

import (
	"os"
	"testing"
)

func TestExpandEnvVariables(t *testing.T) {
	tests := []struct {
		configVar   string
		description string
		envVars     map[string]string
		expectedErr string
		expectedVal string
		fieldName   string
	}{
		{
			// This test case checks the basic functionality of variable expansion.
			// An environment variable "TEST_VAR" is set, and it is expected that
			// the variable in the configuration gets expanded to the value of "TEST_VAR".
			configVar:   "${TEST_VAR}",
			description: "BasicVariableExpansion",
			envVars:     map[string]string{"TEST_VAR": "value"},
			expectedErr: "",
			expectedVal: "value",
			fieldName:   "AccessToken",
		},
		{
			// This test case checks whether a suffix can be successfully appended
			// to the value of an environment variable after it gets expanded.
			configVar:   "${TEST_VAR}_suffix",
			description: "VariableWithSuffix",
			envVars:     map[string]string{"TEST_VAR": "value"},
			expectedErr: "",
			expectedVal: "value_suffix",
			fieldName:   "AccessToken",
		},
		{
			// This test case checks whether two environment variables can be
			// concatenated successfully after they get expanded.
			configVar:   "${TEST_VAR}_${ANOTHER_VAR}",
			description: "ConcatenateTwoVariables",
			envVars:     map[string]string{"TEST_VAR": "value", "ANOTHER_VAR": "another_value"},
			expectedErr: "",
			expectedVal: "value_another_value",
			fieldName:   "AccessToken",
		},
		{
			// This test case checks whether the default value is used when the environment
			// variable is set but empty. The default value "default_value" should be used.
			configVar:   "${TEST_VAR:-default_value}",
			description: "DefaultForEmptyVariable",
			envVars:     map[string]string{"TEST_VAR": ""},
			expectedErr: "",
			expectedVal: "default_value",
			fieldName:   "AccessToken",
		},
		{
			// This test case checks the behavior when an environment variable is unset.
			// The configuration should have an empty value as "UNSET_VAR" is not set.
			configVar:   "${UNSET_VAR}",
			description: "UnsetVariable",
			envVars:     map[string]string{},
			expectedErr: "",
			expectedVal: "",
			fieldName:   "AccessToken",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Setting environment variables
			for varName, val := range test.envVars {
				os.Setenv(varName, val)
				defer os.Unsetenv(varName)

			}

			config := &Config{AccessToken: test.configVar}
			err := config.ExpandEnvVariables()

			if err != nil {
				expErr, ok := err.(*ExpansionError)
				if ok {
					if expErr.FieldName != test.expectedErr {
						t.Errorf("Expected error field name: %q, got: %s", test.expectedErr, expErr.FieldName)
					}
				} else {
					t.Errorf("Expected ExpansionError, got: %v", err)
				}
			} else if test.expectedErr != "" {
				t.Errorf("Expected error for field name: %q, but got nil", test.expectedErr)
			} else {
				actualVal := config.AccessToken
				if actualVal != test.expectedVal {
					t.Errorf("Expected value %q, but got %s", test.expectedVal, actualVal)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		config      *Config
		description string
		expectedErr string // This holds the name of the field expected to error
	}{
		{
			// This test case checks the behavior when the access token is missing.
			config:      &Config{AccessToken: "", ChannelsStr: "channel"},
			description: "MissingAccessToken",
			expectedErr: "SLACK_ACCESS_TOKEN",
		},
		{
			// This test case checks the behavior when the channel string is missing.
			config:      &Config{AccessToken: "token", ChannelsStr: ""},
			description: "MissingChannelString",
			expectedErr: "SLACK_PARAM_CHANNEL",
		},
		{
			// This test case checks the behavior when nothing is missing.
			config:      &Config{AccessToken: "token", ChannelsStr: "channel"},
			description: "ValidConfig",
			expectedErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.config.Validate()

			if err != nil {
				envErr, ok := err.(*EnvVarError)
				if ok {
					if envErr.VarName != test.expectedErr {
						t.Errorf("Expected error var name: %s, got: %s", test.expectedErr, envErr.VarName)
					}
				} else {
					t.Errorf("Expected EnvVarError, got: %v", err)
				}
			} else if test.expectedErr != "" {
				t.Errorf("Expected error for field name: %s, but got nil", test.expectedErr)
			}
		})
	}
}

func TestLoadEnvFromFile(t *testing.T) {
	tests := []struct {
		description string
		envVarName  string
		envVarValue string
		expectedErr bool
		filePath    string
	}{
		{
			// This test case checks the behavior when the file does not exist.
			description: "FileDoesNotExist",
			filePath:    "/path/that/does/not/exist",
			envVarName:  "",
			envVarValue: "",
			expectedErr: false,
		},
		{
			// This test case checks the successful loading of environment variables from a file.
			description: "ValidFile",
			envVarName:  "TEST_VAR",
			envVarValue: "potato",
			expectedErr: false,
			filePath:    "testdata/valid_env_file",
		},
		{
			// This test case checks the behavior when the file is invalid.
			description: "InvalidFile",
			envVarName:  "TEST_VAR",
			envVarValue: "potato",
			expectedErr: true,
			filePath:    "testdata/invalid_env_file",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			err := LoadEnvFromFile(test.filePath)

			if (err != nil) != test.expectedErr {
				t.Errorf("Expected error: %v, got: %v", test.expectedErr, err)
			}

			if !test.expectedErr && test.envVarName != "" {
				val, present := os.LookupEnv(test.envVarName)
				if !present || val != test.envVarValue {
					t.Errorf("Expected env var value: %q, got: %q", test.envVarValue, val)
				}
			}
		})
	}
}
