package templates

import _ "embed"

var (
	//go:embed basic_fail_1.json
	basicFail string

	//nolint:unused
	//go:embed basic_on_hold_1.json
	basicOnHold string

	//go:embed basic_success_1.json
	basicSuccess string

	//go:embed success_tagged_deploy_1.json
	successTaggedDeploy string

	prepared = map[string]string{
		"pass": basicSuccess,
		"fail": basicFail,
	}
)

// ForStatus returns the default template body for the provided status if it exists,
// or the empty string if there is no default.
func ForStatus(status string) string {
	return prepared[status]
}
