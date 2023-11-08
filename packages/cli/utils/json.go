package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/TylerBrock/colorjson"
	"github.com/a8m/envsubst"
	"github.com/fatih/color"
)

func ExpandEnvVarsInInterface(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		expandedValue, _ := envsubst.String(v)
		return expandedValue
	case map[string]interface{}:
		for key, innerValue := range v {
			v[key] = ExpandEnvVarsInInterface(innerValue)
		}
	case []interface{}:
		for i, innerValue := range v {
			v[i] = ExpandEnvVarsInInterface(innerValue)
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
		return "", fmt.Errorf("%s: %w", "ApplyFunctionToJSON - Unmarshal", err)
	}

	modifiedTemplate := modifier(jsonTemplate)

	switch v := modifiedTemplate.(type) {
	case map[string]interface{}:
		result, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("%s: %w", "ApplyFunctionToJSON - Marshal", err)
		}
		return string(result), nil
	case string:
		return v, nil
	case float64: // JSON numbers are decoded into float64 in Go
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("unexpected type %T", v)
	}
}

func InferTemplateEnvVarFromStatus(jobStatus string) (string, error) {
	switch jobStatus {
	case "pass":
		return "basic_success_1", nil
	case "fail":
		return "basic_fail_1", nil
	default:
		return "", fmt.Errorf("the job status: %q is unexpected", jobStatus)
	}
}

func DetermineTemplate(inlineTemplate, jobStatus, envVarContainingTemplate string) (string, error) {
	if inlineTemplate != "" {
		return inlineTemplate, nil
	}

	if envVarContainingTemplate == "" {
		var err error
		envVarContainingTemplate, err = InferTemplateEnvVarFromStatus(jobStatus)
		if err != nil {
			return "", fmt.Errorf("%s: %w", "DetermineTemplate", err)
		}
	}
	template := os.Getenv(envVarContainingTemplate)
	if template == "" {
		return "", fmt.Errorf("the template %q is empty", template)
	}
	return template, nil
}

func ExtractRootProperty(propertyName string) func(interface{}) interface{} {
	return func(data interface{}) interface{} {
		jsonMap, ok := data.(map[string]interface{})
		if !ok {
			return data
		}

		propertyValue, exists := jsonMap[propertyName]
		if exists {
			return propertyValue
		}
		return ""
	}
}

func AddRootProperty(propertyName string, propertyValue interface{}) func(interface{}) interface{} {
	return func(data interface{}) interface{} {
		jsonMap, ok := data.(map[string]interface{})
		if !ok {
			// If the type assertion fails, just return the original data
			return data
		}

		// Add the property
		jsonMap[propertyName] = propertyValue

		return jsonMap
	}
}

func ColorizeJSON(jsonStr string) (string, error) {
	// CircleCI supports color output.
	if os.Getenv("CI") == "true" {
		color.NoColor = false
	}

	var input map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &input)
	if err != nil {
		return "", err
	}

	f := colorjson.NewFormatter()
	f.Indent = 2
	colorized, err := f.Marshal(input)
	if err != nil {
		return "", err
	}
	return string(colorized), nil
}
