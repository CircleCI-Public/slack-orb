package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-multierror"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/utils"
)

var SlackConfig Config

// Config represents the configuration loaded from environment variables.
type Config struct {
	// Required configuration
	AccessToken string
	Channels    string

	// Trigger matching
	BranchPattern      string
	TagPattern         string
	EventToSendMessage string
	JobBranch          string
	JobStatus          string
	JobTag             string

	// Flags
	Debug        bool
	IgnoreErrors string
	InvertMatch  string

	// Message template
	TemplateInline string
	TemplateName   string
	TemplatePath   string
	TemplateVar    string

	// Overridable for testing
	SlackAPIBaseUrl string
}

// InitConfig will initialize the configuration from environment variables.
// This will set 'SlackConfig' to the loaded configuration.
func InitConfig() error {
	if err := bindEnv(); err != nil {
		return err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return errors.New("unable to bind configuration")
	}

	SlackConfig = cfg
	return nil
}

func bindEnv() error {
	// Load environment variables from BASH_ENV and SLACK_JOB_STATUS files
	// This has to be done before loading the configuration because the configuration
	// depends on the environment variables loaded from these files
	if err := loadEnvFromFile(os.Getenv("BASH_ENV")); err != nil {
		return err
	}
	if err := loadEnvFromFile("/tmp/SLACK_JOB_STATUS"); err != nil {
		return err
	}

	var errs error
	for k, v := range map[string]string{
		"AccessToken":        "SLACK_ACCESS_TOKEN",
		"Channels":           "SLACK_STR_CHANNEL",
		"BranchPattern":      "SLACK_STR_BRANCHPATTERN",
		"EventToSendMessage": "SLACK_STR_EVENT",
		"IgnoreErrors":       "SLACK_BOOL_IGNORE_ERRORS",
		"InvertMatch":        "SLACK_BOOL_INVERT_MATCH",
		"JobBranch":          "CIRCLE_BRANCH",
		"JobStatus":          "CCI_STATUS",
		"JobTag":             "CIRCLE_TAG",
		"SlackAPIBaseUrl":    "TEST_SLACK_API_BASE_URL",
		"TagPattern":         "SLACK_STR_TAGPATTERN",
		"TemplateInline":     "SLACK_STR_TEMPLATE_INLINE",
		"TemplateName":       "SLACK_STR_TEMPLATE",
		"TemplatePath":       "SLACK_STR_TEMPLATE_PATH",
		"TemplateVar":        "SLACK_STR_TEMPLATE_VAR",
		"Debug":              "SLACK_BOOL_DEBUG",
	} {
		errs = multierror.Append(errs, viper.BindEnv(k, v))
	}

	return errs.(*multierror.Error).ErrorOrNil()
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

// expandEnvVariables expands environment variables in the configuration values.
func (c *Config) expandEnvVariables() error {
	fields := map[string]*string{
		"AccessToken":        &c.AccessToken,
		"BranchPattern":      &c.BranchPattern,
		"Channels":           &c.Channels,
		"EventToSendMessage": &c.EventToSendMessage,
		"IgnoreErrors":       &c.IgnoreErrors,
		"InvertMatch":        &c.InvertMatch,
		"TagPattern":         &c.TagPattern,
		"TemplateName":       &c.TemplateName,
		"TemplatePath":       &c.TemplatePath,
		"TemplateVar":        &c.TemplateVar,
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
	if err := c.expandEnvVariables(); err != nil {
		return fmt.Errorf("error expanding environment variables: %v", err)
	}
	if c.AccessToken == "" {
		return &EnvVarError{VarName: "SLACK_ACCESS_TOKEN"}
	}
	if c.Channels == "" {
		return &EnvVarError{VarName: "SLACK_STR_CHANNEL"}
	}
	if c.JobStatus != "pass" && c.JobStatus != "fail" {
		return fmt.Errorf("invalid value for CCI_STATUS: %s", c.JobStatus)
	}
	return nil
}

// handleOSSpecifics checks and applies OS-specific modifications to the file.
func handleOSSpecifics(filePath string) (string, error) {
	if runtime.GOOS == "windows" {
		filePath = strings.Replace(filePath, "/tmp", "C:/Users/circleci/AppData/Local/Temp", 1)
		if !utils.FileExists(filePath) {
			err := ConvertFileToCRLF(filePath)
			if err != nil {
				return "", fmt.Errorf("error converting file to CRLF: %w", err)
			}
		}
	}
	return filePath, nil
}

// loadEnvFromFile loads environment variables from a specified file.
func loadEnvFromFile(filePath string) error {
	log.Debug("Starting to load environment variables from file.")

	modifiedPath, err := handleOSSpecifics(filePath)
	if err != nil {
		return fmt.Errorf("OS-specific handling failed: %v", err)
	}

	if !utils.FileExists(modifiedPath) {
		log.Debugf("File %q does not exist. Skipping...\n", modifiedPath)
		return nil
	}

	log.Debugf("Loading %q into the environment...\n", modifiedPath)
	if err := godotenv.Load(modifiedPath); err != nil {
		return fmt.Errorf("error loading %q file: %v", modifiedPath, err)
	}

	log.Debug("Environment variables loaded successfully.")
	return nil
}

var (
	CRLF = []byte{13, 10}
	LF   = []byte{10}
)

// ConvertFileToCRLF converts line endings in a file to CRLF.
func ConvertFileToCRLF(filePath string) error {
	//nolint:gosec // G304 path is validated elsewhere
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

func GetDebug() bool {
	debugBool, err := strconv.ParseBool(os.Getenv("SLACK_BOOL_DEBUG"))
	if err != nil {
		return false
	}
	return debugBool
}
