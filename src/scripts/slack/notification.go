package slack

import (
	"errors"
	"fmt"

	"github.com/CircleCI-Public/slack-orb-go/src/scripts/jsonutils"
	"github.com/CircleCI-Public/slack-orb-go/src/scripts/stringutils"
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
	branchMatches, _ := stringutils.IsPatternMatchingString(j.BranchPattern, j.Branch)
	tagMatches, _ := stringutils.IsPatternMatchingString(j.TagPattern, j.Tag)
	return (branchMatches || tagMatches) != j.InvertMatch

}

func (j *Notification) BuildMessageBody() (string, error) {
	// Build the message body
	template, err := jsonutils.DetermineTemplate(j.InlineTemplate, j.Status, j.EnvVarContainingTemplate)
	if err != nil {
		return "", err
	}
	if template == "" {
		return "", fmt.Errorf("the template %q is empty. Exiting without posting to Slack", template)
	}

	// Expand environment variables in the template
	templateWithExpandedVars, err := jsonutils.ApplyFunctionToJSON(template, jsonutils.ExpandEnvVarsInInterface)
	if err != nil {
		return "", err
	}

	// Add a "channel" property with a nested "myChannel" property
	modifiedJSON, err := jsonutils.ApplyFunctionToJSON(templateWithExpandedVars, jsonutils.AddRootProperty("channel", "my_channel"))
	if err != nil {
		return "", err
	}

	if !j.IsEventMatchingStatus() {
		message := fmt.Sprintf("Exiting without posting to Slack: The job status %q does not match the status set to send alerts %q.", j.Status, j.Event)
		err = errors.New(message)
		return "", err
	}

	if !j.IsPostConditionMet() {
		err = errors.New("Exiting without posting to Slack: The post condition is not met. Neither the branch nor the tag matches the pattern or the match is inverted.")
		return "", err
	}

	return modifiedJSON, nil
}
