package acceptance

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/circleci/ex/testing/compiler"
	"github.com/circleci/ex/testing/testcontext"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/skip"

	"github.com/CircleCI-Public/slack-orb-go/packages/cli/internal/testing/fakeslack"
)

func TestSlackOrbBinary(t *testing.T) {
	skip.If(t, testing.Short, "Test compiles and executes local binaries")

	ctx := testcontext.Background()
	fix := setupE2E(ctx, t)

	tests := []struct {
		name                      string
		environment               map[string]string
		expectedExitCode          int
		expectedOutput            string
		expectedSlackAPICallCount int
	}{{
		name: "Basic success template",
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
			"SLACK_STR_CHANNEL":  "test-channel",
			"CCI_STATUS":         "pass",
			"SLACK_STR_EVENT":    "pass",
		},
		expectedExitCode:          0,
		expectedOutput:            "Successfully posted message to channel: test-channel",
		expectedSlackAPICallCount: 1,
	}, {
		name: "Debug flag enabled",
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
			"SLACK_STR_CHANNEL":  "test-channel",
			"CCI_STATUS":         "pass",
			"SLACK_STR_EVENT":    "pass",
			"SLACK_BOOL_DEBUG":   "true",
		},
		expectedExitCode:          0,
		expectedOutput:            "Posting the following JSON to Slack",
		expectedSlackAPICallCount: 1,
	}, {
		name: "Basic fail template",
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
			"SLACK_STR_CHANNEL":  "test-channel",
			"CCI_STATUS":         "fail",
			"SLACK_STR_EVENT":    "fail",
		},
		expectedExitCode:          0,
		expectedOutput:            "Successfully posted message to channel: test-channel",
		expectedSlackAPICallCount: 1,
	}, {
		name: "Missing slack token",
		environment: map[string]string{
			"SLACK_STR_CHANNEL": "test-channel",
		},
		expectedExitCode: 1,
		expectedOutput:   "In order to use the Slack Orb an OAuth token must be present via the SLACK_ACCESS_TOKEN environment variable.",
	}, {
		name: "Missing slack channel",
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
		},
		expectedExitCode: 1,
		expectedOutput:   `No channel was provided. Please provide one or more channels using the "SLACK_STR_CHANNEL" environment variable or the "channel" parameter.`,
	}, {
		name:             "Job status does not match",
		expectedExitCode: 0,
		expectedOutput:   `Exiting without posting to Slack: The job status "fail" does not match the status set to send alerts "pass".`,
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
			"SLACK_STR_CHANNEL":  "test-channel",
			"CCI_STATUS":         "fail",
			"SLACK_STR_EVENT":    "pass",
		},
	}, {
		name:             "Multiline string env var parsed correctly",
		expectedExitCode: 0,
		expectedOutput:   `This message should show over multiple lines: Line 1.\\nLine 2.\\nLine 3.`,
		environment: map[string]string{
			"SLACK_ACCESS_TOKEN": "test-token",
			"SLACK_STR_CHANNEL":  "test-channel",
			"CCI_STATUS":         "pass",
			"SLACK_STR_EVENT":    "pass",
			"SLACK_BOOL_DEBUG":   "true",
			"MULTILINE_STRING":   `Line 1.\nLine 2.\nLine 3.`,
			"SLACK_STR_TEMPLATE_INLINE": `{
"blocks": [
	{
		"type": "section",
		"text": {
			"type": "mrkdwn",
			"text": "This message should show over multiple lines: $MULTILINE_STRING"
		}
	}
]
}`,
		},
		expectedSlackAPICallCount: 1,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slackAPIServer := httptest.NewServer(fix.slackAPI.Handler())
			t.Cleanup(func() {
				fix.slackAPI.Reset()
				slackAPIServer.Close()
			})

			cmd := exec.Command(fix.slackOrbPath, "notify")

			comparableOutput := &strings.Builder{}
			w := io.MultiWriter(os.Stdout, comparableOutput)
			cmd.Stdout = w
			cmd.Stderr = w
			cmd.Env = append(cmd.Environ(), "TEST_SLACK_API_BASE_URL="+slackAPIServer.URL)
			for key, value := range tt.environment {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
			}

			assert.Assert(t, cmd.Start())
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
			})

			err := cmd.Wait()
			if tt.expectedExitCode == 0 {
				assert.NilError(t, err)
			}

			assert.Check(t, cmp.Equal(cmd.ProcessState.ExitCode(), tt.expectedExitCode))
			assert.Check(t, cmp.Equal(len(fix.slackAPI.AllRequests()), tt.expectedSlackAPICallCount))

			if tt.expectedOutput != "" {
				assert.Check(t, cmp.Contains(comparableOutput.String(), tt.expectedOutput))
			}
		})
	}
}

type e2eFixture struct {
	slackOrbPath string
	binariesDir  string

	slackAPI *fakeslack.API
}

func setupE2E(ctx context.Context, t *testing.T) *e2eFixture {
	slack := fakeslack.New(ctx)

	var slackOrbBinary string
	c := compiler.NewParallel(1)
	c.Add(compiler.Work{
		Result: &slackOrbBinary,
		Name:   "slack-orb-cli",
		Target: "../../",
		Source: "github.com/CircleCI-Public/slack-orb-go/packages/cli",
	})
	err := c.Run(ctx)
	assert.Assert(t, err)

	return &e2eFixture{
		slackOrbPath: slackOrbBinary,
		binariesDir:  c.Dir(),
		slackAPI:     slack,
	}
}
