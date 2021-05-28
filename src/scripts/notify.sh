BuildMessageBody() {
    # Send message
    #   If sending message, default to custom template,
    #   if none is supplied, check for a pre-selected template value.
    #   If none, error.
    if [ -n "${SLACK_PARAM_CUSTOM:-}" ]; then
        ModifyCustomTemplate
        # shellcheck disable=SC2016
        CUSTOM_BODY_MODIFIED=$(echo "$CUSTOM_BODY_MODIFIED" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed 's/`/\\`/g')
        T2=$(eval echo \""$CUSTOM_BODY_MODIFIED"\")
    elif [ -n "${SLACK_PARAM_TEMPLATE:-}" ]; then
        TEMPLATE="\$$SLACK_PARAM_TEMPLATE"
        T1=$(eval echo "$TEMPLATE" | sed 's/"/\\"/g')
        T2=$(eval echo \""$T1"\")
    else
        echo "Error: No message template selected."
        echo "Select either a custom template or one of the pre-included ones via the 'custom' or 'template' parameters."
        exit 1
    fi
    # Insert the default channel. THIS IS TEMPORARY
    T2=$(echo "$T2" | jq ". + {\"channel\": \"$SLACK_DEFAULT_CHANNEL\"}")
    SLACK_MSG_BODY=$T2
}


PostToSlack() {
    # Post once per channel listed by the channel parameter
    #    The channel must be modified in SLACK_MSG_BODY

    # shellcheck disable=SC2001
    for i in $(eval echo \""$SLACK_PARAM_CHANNEL"\" | sed "s/,/ /g")
    do
        echo "Sending to Slack Channel: $i"
        SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg channel "$i" '.channel = $channel')
        SLACK_SENT_RESPONSE=$(curl -s -f -X POST -H 'Content-type: application/json' -H "Authorization: Bearer $SLACK_ACCESS_TOKEN" --data "$SLACK_MSG_BODY" https://slack.com/api/chat.postMessage)
        SLACK_ERROR_MSG=$(echo "$SLACK_SENT_RESPONSE" | jq '.error')
        if [ ! "$SLACK_ERROR_MSG" = "null" ]; then
            echo "Slack API returned an error message:"
            echo "$SLACK_ERROR_MSG"
            echo
            echo
            echo "View the Setup Guide: https://github.com/CircleCI-Public/slack-orb/wiki/Setup"
            if [ "$SLACK_PARAM_IGNORE_ERRORS" = "0" ]; then
                exit 1
            fi
        fi
    done
}

ModifyCustomTemplate() {
    # Inserts the required "text" field to the custom json template from block kit builder.
    if [ "$(echo "$SLACK_PARAM_CUSTOM" | jq '.text')" = "null" ]; then
        CUSTOM_BODY_MODIFIED=$(echo "$SLACK_PARAM_CUSTOM" | jq '. + {"text": ""}')
    else
        # In case the text field was set manually.
        CUSTOM_BODY_MODIFIED=$(echo "$SLACK_PARAM_CUSTOM" | jq '.')
    fi
}

InstallJq() {
    if uname -a | grep Darwin > /dev/null 2>&1; then
        echo "Checking For JQ + CURL: MacOS"
        command -v jq >/dev/null 2>&1 || HOMEBREW_NO_AUTO_UPDATE=1 brew install jq --quiet
        return $?

    elif cat /etc/issue | grep Debian > /dev/null 2>&1 || cat /etc/issue | grep Ubuntu > /dev/null 2>&1; then
        echo "Checking For JQ + CURL: Debian"
        if [ "$(id -u)" = 0 ]; then export SUDO=""; else # Check if we're root
            export SUDO="sudo";
        fi
        command -v jq >/dev/null 2>&1 || { $SUDO apt -qq update && $SUDO apt -qq install -y jq; }
        return $?

    elif cat /etc/issue | grep Alpine > /dev/null 2>&1; then
        echo "Checking For JQ + CURL: Alpine"
        command -v curl >/dev/null 2>&1 || { echo >&2 "SLACK ORB ERROR: CURL is required. Please install."; exit 1; }
        command -v jq >/dev/null 2>&1 || { echo >&2 "SLACK ORB ERROR: JQ is required. Please install"; exit 1; }
        return $?
    fi
}

FilterBy() {
    # If any pattern supplied matches the current branch or the current tag, proceed; otherwise, exit with message.
    FLAG_MATCHES_FILTER="false"
    for i in $(echo "$1" | sed "s/,/ /g")
    do
        if echo "$2" | grep -Eq "^${i}$"; then
            FLAG_MATCHES_FILTER="true"
            break
        fi
    done
    if [ "$FLAG_MATCHES_FILTER" = "false" ]; then
        return 1
    fi
}

CheckEnvVars() {
    if [ -z "${SLACK_ACCESS_TOKEN:-}" ]; then
        echo "In order to use the Slack Orb (v4 +), an OAuth token must be present via the SLACK_ACCESS_TOKEN environment variable."
        echo "Follow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup"
        exit 1
    fi
    # If no channel is provided, quit with error
    if [ -z "${SLACK_PARAM_CHANNEL:-}" ]; then
       echo "No channel was provided. Enter value for SLACK_DEFAULT_CHANNEL env var, or channel parameter"
       exit 1
    fi
    if [ -n "${SLACK_WEBHOOK:-}" ]; then
        echo "It appears you have a Slack Webhook token present in this job."
        echo "Please note, Webhooks are no longer used for the Slack Orb (v4 +)."
        echo "Follow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup"
    fi
}

AbortPost() {
    echo "NO SLACK ALERT"
    echo
    for s in "$@"; do
        echo "$s"
    done
    exit 0
}

ShouldPost() {
    if [ "$CCI_STATUS" != "$SLACK_PARAM_EVENT" ]; then
        if [ "$SLACK_PARAM_EVENT" != "always" ]; then
            AbortPost \
                "This command is set to send an alert on: $SLACK_PARAM_EVENT" \
                "Current status: ${CCI_STATUS}"
        fi
    fi

    # In the event the Slack notification would be sent, first ensure it is allowed to trigger
    # on this branch or this tag.
    if [ -n "${CIRCLE_BRANCH:-}" ]; then
        if ! FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"; then
            AbortPost \
                "Branch pattern does not match: ${SLACK_PARAM_BRANCHPATTERN}" \
                "CI was triggered by branch: ${CIRCLE_BRANCH}"
        fi
    elif [ -n "${CIRCLE_TAG:-}" ]; then
        if ! FilterBy "$SLACK_PARAM_TAGPATTERN" "$CIRCLE_TAG"; then
            AbortPost \
                "Tag pattern does not match: ${SLACK_PARAM_TAGPATTERN}" \
                "CI was triggered by tag: ${CIRCLE_TAG}"
        fi
    else
        AbortPost \
            "Neither CIRCLE_BRANCH nor CIRCLE_TAG was set" \
            "Unable to determine whether orb should post"
    fi

    echo "Posting Status"
}

# Will not run if sourced from another script.
# This is done so this script may be tested.
ORB_TEST_ENV="bats-core"
if [ "${0#*$ORB_TEST_ENV}" = "$0" ]; then
    CheckEnvVars
    . "/tmp/SLACK_JOB_STATUS"
    ShouldPost
    InstallJq
    BuildMessageBody
    PostToSlack

fi
