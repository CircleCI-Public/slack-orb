package slack

import (
	"errors"
	"fmt"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

type Notification struct {
	Status                   string
	Branch                   string
	Tag                      string
	Event                    string
	BranchPattern            string
	TagPattern               string
	InvertMatch              bool
	Template                 string
	InlineTemplate           string
	EnvVarContainingTemplate string
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
	template, err := utils.DetermineTemplate(j.InlineTemplate, j.Status, j.EnvVarContainingTemplate)
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

	// Add a "channel" property with a nested "myChannel" property
	modifiedJSON, err := utils.ApplyFunctionToJSON(templateWithExpandedVars,
		utils.AddRootProperty("channel", "my_channel"))
	if err != nil {
		return "", err
	}

	if !j.IsEventMatchingStatus() {
		return "", ErrStatusMismatch
	}

	if !j.IsPostConditionMet() {
		return "", ErrPostConditionNotMet
	}

	return modifiedJSON, nil
}
