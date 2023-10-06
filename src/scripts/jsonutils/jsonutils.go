package jsonutils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/a8m/envsubst"
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
		return "", err
	}

	modifiedTemplate := modifier(jsonTemplate).(map[string]interface{})

	result, err := json.Marshal(modifiedTemplate)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func InferTemplateEnvVarFromStatus(jobStatus string) (string, error) {
	switch jobStatus {
	case "success":
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
			return "", err
		}
	}
	template := os.Getenv(envVarContainingTemplate)
	if template == "" {
		return "", fmt.Errorf("the template %q is empty", template)
	}
	return template, nil
}
