package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEventMatchingStatus(t *testing.T) {
	tests := []struct {
		name   string
		event  string
		status string
		want   bool
	}{
		{
			name:   "matching event and status",
			event:  "push",
			status: "push",
			want:   true,
		},
		{
			name:   "non-matching event and status",
			event:  "push",
			status: "pull",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sn := Notification{Status: tt.status, Event: tt.event}
			got := sn.IsEventMatchingStatus()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPostConditionMet(t *testing.T) {
	tests := []struct {
		name          string
		branch        string
		tag           string
		branchPattern string
		tagPattern    string
		invertMatch   bool
		want          bool
	}{
		{
			name:          "matching branch and tag patterns",
			branch:        "main",
			tag:           "v1.0",
			branchPattern: "main",
			tagPattern:    "v1.*",
			invertMatch:   false,
			want:          true,
		},
		{
			name:          "non-matching branch and tag patterns",
			branch:        "dev",
			tag:           "v2.0",
			branchPattern: "main",
			tagPattern:    "v1.*",
			invertMatch:   false,
			want:          false,
		},
		{
			name:          "invert match",
			branch:        "dev",
			tag:           "v2.0",
			branchPattern: "main",
			tagPattern:    "v1.*",
			invertMatch:   true,
			want:          true,
		},
		{
			name:          "empty branch and tag",
			branch:        "",
			tag:           "",
			branchPattern: "main",
			tagPattern:    "v1.*",
			invertMatch:   false,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sn := Notification{
				Branch:        tt.branch,
				Tag:           tt.tag,
				BranchPattern: tt.branchPattern,
				TagPattern:    tt.tagPattern,
				InvertMatch:   tt.invertMatch,
			}
			got := sn.IsPostConditionMet()
			assert.Equal(t, tt.want, got)
		})
	}
}
