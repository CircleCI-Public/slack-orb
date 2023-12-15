package templates

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

var (
	//go:embed basic_fail_1.json
	basicFail string

	//go:embed basic_success_1.json
	basicSuccess string

	//go:embed success_tagged_deploy_1.json
	successTaggedDeploy string
)

var (
	prepared = map[string]string{
		"pass": basicSuccess,
		"fail": basicFail,
	}

	preparedNames = map[string]string{
		"basic_fail_1":            basicFail,
		"basic_success_1":         basicSuccess,
		"success_tagged_deploy_1": successTaggedDeploy,
	}
)

var (
	ErrFileNotExist     = errors.New("the file does not exist")
	ErrFileRead         = errors.New("could not read the file")
	ErrInvalidStatus    = errors.New("the status provided is invalid")
	ErrTemplateEmpty    = errors.New("the environment variable provided is empty")
	ErrTemplateNotExist = errors.New("the template does not exist")
)

// ForStatus returns the default template body for the provided status if it exists,
// or the empty string if there is no default.
func ForStatus(status string) string {
	return prepared[status]
}

// ForName returns the default template body for the provided name if it exists,
// or the empty string if there is no default.
func ForName(name string) string {
	return preparedNames[name]
}

// DetermineTemplate returns the template to use for the notification.
// The order of precedence is templateVar, templatePath, templateInline, templateName.
// If none of these are provided the default template for the job status is used.
func DetermineTemplate(templateVar, templatePath, templateInline, templateName, jobStatus string) (string, error) {
	switch {
	case templateVar != "":
		return getTemplateFromEnv(templateVar)
	case templatePath != "":
		return getTemplateFromFile(templatePath)
	case templateInline != "":
		return templateInline, nil
	case templateName != "":
		return getTemplateByName(templateName)
	default:
		return getTemplateForStatus(jobStatus)
	}
}

func getTemplateFromEnv(templateVar string) (string, error) {
	template := os.Getenv(templateVar)
	if template == "" {
		return "", fmt.Errorf("%w: %s", ErrTemplateEmpty, templateVar)
	}
	return template, nil
}

func getTemplateFromFile(templatePath string) (string, error) {
	if !utils.FileExists(templatePath) {
		return "", fmt.Errorf("%w: %s", ErrFileNotExist, templatePath)
	}
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrFileRead, templatePath)
	}
	return string(template), nil
}

func getTemplateByName(templateName string) (string, error) {
	template := ForName(templateName)
	if template == "" {
		return "", fmt.Errorf("%w: %s", ErrTemplateNotExist, templateName)
	}
	return template, nil
}

func getTemplateForStatus(jobStatus string) (string, error) {
	template := ForStatus(jobStatus)
	if template == "" {
		return "", fmt.Errorf("%w: %s", ErrInvalidStatus, jobStatus)
	}
	return template, nil
}
