package config

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/joho/godotenv"

	"github.com/CircleCI-Public/slack-orb-go/src/scripts/ioutils"
)

// Config represents the configuration loaded from environment variables.
type Config struct {
	AccessToken              string
	BranchPattern            string
	ChannelsStr              string
	EnvVarContainingTemplate string
	EventToSendMessage       string
	InlineTemplate           string
	InvertMatchStr           string
	JobBranch                string
	JobStatus                string
	JobTag                   string
	TagPattern               string
	IgnoreErrorsStr          string
}

// NewConfig loads configuration from environment variables.
func NewConfig() *Config {
	return &Config{
		AccessToken:              os.Getenv("SLACK_ACCESS_TOKEN"),
		BranchPattern:            os.Getenv("SLACK_PARAM_BRANCHPATTERN"),
		ChannelsStr:              os.Getenv("SLACK_PARAM_CHANNEL"),
		EnvVarContainingTemplate: os.Getenv("SLACK_PARAM_TEMPLATE"),
		EventToSendMessage:       os.Getenv("SLACK_PARAM_EVENT"),
		InlineTemplate:           os.Getenv("SLACK_PARAM_CUSTOM"),
		InvertMatchStr:           os.Getenv("SLACK_PARAM_INVERT_MATCH"),
		JobBranch:                os.Getenv("CIRCLE_BRANCH"),
		JobStatus:                os.Getenv("CCI_STATUS"),
		JobTag:                   os.Getenv("CIRCLE_TAG"),
		TagPattern:               os.Getenv("SLACK_PARAM_TAGPATTERN"),
		IgnoreErrorsStr:          os.Getenv("SLACK_PARAM_IGNORE_ERRORS"),
	}
}

type EnvVarError struct {
	VarName string
}

func (e *EnvVarError) Error() string {
	return fmt.Sprintf("environment variable not set: %s", e.VarName)
}

type ExpansionError struct {
	FieldName string
	Err       error
}

func (e *ExpansionError) Error() string {
	return fmt.Sprintf("error expanding %s: %v", e.FieldName, e.Err)
}

// ExpandEnvVariables expands environment variables in the configuration values.
func (c *Config) ExpandEnvVariables() error {
	fields := map[string]*string{
		"AccessToken":              &c.AccessToken,
		"BranchPattern":            &c.BranchPattern,
		"ChannelsStr":              &c.ChannelsStr,
		"EnvVarContainingTemplate": &c.EnvVarContainingTemplate,
		"EventToSendMessage":       &c.EventToSendMessage,
		"InvertMatchStr":           &c.InvertMatchStr,
		"IgnoreErrorsStr":          &c.IgnoreErrorsStr,
		"TagPattern":               &c.TagPattern,
	}

	for fieldName, fieldValue := range fields {
		val, err := envsubst.String(*fieldValue)

		if err != nil {
			return &ExpansionError{FieldName: fieldName, Err: err}
		}
		*fieldValue = val
	}

	return nil
}

// Validate checks whether the necessary environment variables are set.
func (c *Config) Validate() error {
	if c.AccessToken == "" {
		return &EnvVarError{VarName: "SLACK_ACCESS_TOKEN"}
	}
	if c.ChannelsStr == "" {
		return &EnvVarError{VarName: "SLACK_PARAM_CHANNEL"}
	}
	return nil
}

// handleOSSpecifics checks and applies OS-specific modifications to the file.
func handleOSSpecifics(filePath string) (string, error) {
	if runtime.GOOS == "windows" {
		filePath = strings.Replace(filePath, "/tmp", "C:/Users/circleci/AppData/Local/Temp", 1)

		err := ConvertFileToCRLF(filePath)
		if err != nil {
			return "", fmt.Errorf("error converting file to CRLF: %w", err)
		}
	}
	return filePath, nil
}

// LoadEnvFromFile loads environment variables from a specified file.
func LoadEnvFromFile(filePath string) error {
	fmt.Println("Starting to load environment variables from file.")

	modifiedPath, err := handleOSSpecifics(filePath)
	if err != nil {
		return fmt.Errorf("OS-specific handling failed: %v", err)
	}

	if !ioutils.FileExists(modifiedPath) {
		fmt.Printf("File %q does not exist. Skipping...\n", modifiedPath)
		return nil
	}

	fmt.Printf("Loading %q into the environment...\n", modifiedPath)
	if err := godotenv.Load(modifiedPath); err != nil {
		return fmt.Errorf("error loading %q file: %v", modifiedPath, err)
	}

	fmt.Println("Environment variables loaded successfully.")
	return nil
}

// ConvertFileToCRLF converts line endings in a file to CRLF.
var (
	CRLF = []byte{13, 10}
	LF   = []byte{10}
)

// ConvertFileToCRLF converts line endings in a file to CRLF.
func ConvertFileToCRLF(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content = bytes.ReplaceAll(content, CRLF, LF)
	content = bytes.ReplaceAll(content, LF, CRLF)

	if err := os.WriteFile(filePath, content, 0); err != nil {
		return err
	}

	return nil
}
