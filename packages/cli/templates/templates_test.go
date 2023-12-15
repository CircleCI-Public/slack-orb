package templates

import (
	"os"
	"testing"
)

func TestDetermineMessageBody(t *testing.T) {
	// Set up mock environment variables for the test
	_ = os.Setenv("MY_ENV_VAR_TEMPLATE", `{ "messageKey": "messageValue" }`)

	tests := []struct {
		name           string
		templateVar    string
		templatePath   string
		templateInline string
		template       string
		jobStatus      string
		expected       string
		hasError       bool
	}{
		{
			name:           "use inline template",
			templateVar:    "",
			templatePath:   "",
			templateInline: `{ "customMessageKey": "customMessageValue" }`,
			template:       "",
			jobStatus:      "success",
			expected:       `{ "customMessageKey": "customMessageValue" }`,
			hasError:       false,
		},
		{
			name:           "use provided template from env var",
			templateVar:    "MY_ENV_VAR_TEMPLATE",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "pass",
			expected:       `{ "messageKey": "messageValue" }`,
			hasError:       false,
		},
		{
			name:           "use template inferred from job status",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "pass",
			expected:       ForStatus("pass"),
			hasError:       false,
		},
		{
			name:           "use alternate inferred template",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "fail",
			expected:       ForStatus("fail"),
			hasError:       false,
		},
		{
			name:           "error because the job status is invalid",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "unknown",
			expected:       "",
			hasError:       true,
		},
		{
			name:           "error with non-existent environment variable",
			templateVar:    "non_existent_env_var",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "pass",
			expected:       "",
			hasError:       true,
		},
		{
			name:           "error with empty template var and non-existent job status",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "nonexistent_status",
			expected:       "",
			hasError:       true,
		},
		{
			name:           "valid template name",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "basic_success_1",
			jobStatus:      "pass",
			expected:       ForName("basic_success_1"),
			hasError:       false,
		},
		{
			name:           "invalid template name",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "nonexistent_template_name",
			jobStatus:      "pass",
			expected:       "",
			hasError:       true,
		},
		{
			name:           "all parameters empty",
			templateVar:    "",
			templatePath:   "",
			templateInline: "",
			template:       "",
			jobStatus:      "",
			expected:       "",
			hasError:       true,
		},
		{
			name:           "inline template with job status",
			templateVar:    "",
			templatePath:   "",
			templateInline: `{ "inlineTemplateKey": "inlineTemplateValue" }`,
			template:       "",
			jobStatus:      "fail",
			expected:       `{ "inlineTemplateKey": "inlineTemplateValue" }`,
			hasError:       false,
		},
		{
			name:           "invalid template path",
			templateVar:    "",
			templatePath:   "./basic_potato.json",
			templateInline: "",
			template:       "",
			jobStatus:      "pass",
			expected:       "",
			hasError:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DetermineTemplate(test.templateVar, test.templatePath, test.templateInline, test.template, test.jobStatus)
			if test.hasError {
				if err == nil {
					t.Errorf("Expected an error but got %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s, error: %v", test.name, err)
				}
				if result != test.expected {
					t.Errorf("Expected %s, got %s", test.expected, result)
				}
			}
		})
	}

	// Clean up mock environment variables after the test
	_ = os.Unsetenv("MY_ENV_VAR_TEMPLATE")
}
