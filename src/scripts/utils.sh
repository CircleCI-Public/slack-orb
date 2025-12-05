#!/bin/false
# shellcheck shell=bash
# shellcheck disable=SC2154

FilterBy() {
    if [ -z "$1" ]; then
        return
    fi
    # If any pattern supplied matches the current branch or the current tag, proceed; otherwise, exit with message.
    FLAG_MATCHES_FILTER="false"
    # shellcheck disable=SC2001
    for i in $(echo "$1" | sed "s/,/ /g"); do
        if echo "$2" | grep -Eq "^${i}$"; then
            FLAG_MATCHES_FILTER="true"
            break
        fi
    done
    # If the invert_match parameter is set, invert the match.
    if { [ "$FLAG_MATCHES_FILTER" = "false" ] && [ "$SLACK_PARAM_INVERT_MATCH" -eq 0 ]; } ||
        { [ "$FLAG_MATCHES_FILTER" = "true" ] && [ "$SLACK_PARAM_INVERT_MATCH" -eq 1 ]; }; then
        # dont send message.
        echo "NO SLACK ALERT"
        echo
        echo "Current reference \"$2\" does not match any matching parameter"
        echo "Current matching pattern: $1"
        exit 0
    fi
}

ShouldPost() {
    if [ "$CCI_STATUS" = "$SLACK_PARAM_EVENT" ] || [ "$SLACK_PARAM_EVENT" = "always" ]; then
        # In the event the Slack notification would be sent, first ensure it is allowed to trigger
        # on this branch or this tag.
        FilterBy "$SLACK_PARAM_BRANCHPATTERN" "${CIRCLE_BRANCH:-}"
        FilterBy "$SLACK_PARAM_TAGPATTERN" "${CIRCLE_TAG:-}"

        echo "Posting Status"
    else
        # dont send message.
        echo "NO SLACK ALERT"
        echo
        echo "This command is set to send an alert on: $SLACK_PARAM_EVENT"
        echo "Current status: ${CCI_STATUS}"
        exit 0
    fi
}

detect_os() {
  detected_platform="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case "$detected_platform" in
    linux*)
      PLATFORM=linux
      ;;
    darwin*)
      PLATFORM=macos
      ;;
    msys*|cygwin*)
      PLATFORM=windows
      ;;
    *)
      printf '%s\n' "Unsupported OS: \"$platform\"."
      exit 1
      ;;
  esac
  readonly PLATFORM
  export PLATFORM
}
