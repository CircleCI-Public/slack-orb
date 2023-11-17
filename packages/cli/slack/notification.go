package slack

import (
	"errors"
	"fmt"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/templates"
	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

type Notification struct {
	Status         string
	Branch         string
	Tag            string
	Event          string
	BranchPattern  string
	TagPattern     string
	InvertMatch    bool
	TemplateVar    string
	TemplatePath   string
	TemplateInline string
	TemplateName   string
}

func (j *Notification) IsEventMatchingStatus() bool {
	return j.Status == j.Event || j.Event == "always"
}

func (j *Notification) IsPostConditionMet() bool {
	branchMatches, _ := utils.IsPatternMatchingString(j.BranchPattern, j.Branch)
	tagMatches, _ := utils.IsPatternMatchingString(j.TagPattern, j.Tag)
	return (branchMatches || tagMatches) != j.InvertMatch

}

var (
	ErrStatusMismatch      = errors.New("job status does not match configured trigger")
	ErrPostConditionNotMet = errors.New("post condition is not met")
)

func (j *Notification) BuildMessageBody() (string, error) {
	// Build the message body
	template, err := templates.DetermineTemplate(j.TemplateVar, j.TemplatePath, j.TemplateInline, j.TemplateName, j.Status)
	if err != nil {
		return "", err
	}
	if template == "" {
		return "", fmt.Errorf("the template %q is empty. Exiting without posting to Slack", template)
	}

	// Expand environment variables in the template
	templateWithExpandedVars, err := utils.ApplyFunctionToJSON(template, utils.ExpandEnvVarsInInterface)
	if err != nil {
		return "", err
	}

	if !j.IsEventMatchingStatus() {
		return "", ErrStatusMismatch
	}

	if !j.IsPostConditionMet() {
		return "", ErrPostConditionNotMet
	}

	return templateWithExpandedVars, nil
}
