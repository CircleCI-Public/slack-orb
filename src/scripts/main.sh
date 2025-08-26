#!/usr/bin/env bash

# Workaround for Windows Support
# For details, see: https://github.com/CircleCI-Public/slack-orb/pull/380
# shellcheck source=/dev/null
if [ "$SLACK_PARAM_DEBUG" -eq 1 ]; then
    set -x
fi

eval printf '%s' "$SLACK_SCRIPT_NOTIFY"
