#!/usr/bin/env sh

# Workaround for Windows Support
# For details, see: https://github.com/CircleCI-Public/slack-orb/pull/380
# shellcheck source=/dev/null
eval printf '%s' "$SLACK_SCRIPT_NOTIFY"
