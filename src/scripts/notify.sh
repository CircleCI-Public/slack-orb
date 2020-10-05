BuildMessageBody() {
    # Send message
    #   If sending message, default to custom template,
    #   if none is supplied, check for a pre-selected template value.
    #   If none, error.
    if [ -n "$SLACK_PARAM_CUSTOM" ]; then
        ModifyCustomTemplate
        CUSTOM_BODY_MODIFIED=$(echo "$CUSTOM_BODY_MODIFIED" | sed 's/"/\\"/g' | sed 's/\\n/\\\\n/g' | sed 's/|/\\|/g' | sed 's/</\\</g' | sed 's/>/\\>/g')
        T2=$(eval echo $CUSTOM_BODY_MODIFIED)
    elif [ -n "$SLACK_PARAM_TEMPLATE" ]; then
        TEMPLATE="$(echo \$$SLACK_PARAM_TEMPLATE)"
        T1=$(eval echo $TEMPLATE | sed 's/"/\\"/g' | sed 's/\\n/\\\\n/g')
        T2=$(eval echo $T1)
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

    # If no channel is provided, quit with error
    [ "$SLACK_PARAM_CHANNEL" != "" ] || echo "No channel was provided. Enter value for SLACK_DEFAULT_CHANNEL env var, or channel parameter"; exit 0;
    for i in $(echo $(eval echo "$SLACK_PARAM_CHANNEL")  | sed "s/,/ /g")
    do
        echo "Sending to Slack Channel: $i"
        SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg channel "$i" '.channel = $channel')
        curl -s -f -X POST -H 'Content-type: application/json' \
        -H "Authorization: Bearer $SLACK_ACCESS_TOKEN" \
        --data \
        "$SLACK_MSG_BODY" https://slack.com/api/chat.postMessage > /dev/null
    done
}

Notify() {
    if [ "$CCI_STATUS" = "$SLACK_PARAM_EVENT" ] || [ "$SLACK_PARAM_EVENT" = "always" ]; then
    BranchFilter # In the event the Slack notification would be sent, first ensure it is allowed to trigger on this branch.
    PostToSlack
    echo "Sending Notification"
    else
        # dont send message.
        echo "NO SLACK ALERT"
        echo
        echo "This command is set to send an alert on: $SLACK_PARAM_EVENT"
        echo "Current status: ${CCI_STATUS}"
        exit 0
    fi
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
    if uname -a | grep darwin > /dev/null 2>&1; then
        echo "Installing JQ for Mac."
        command -v jq >/dev/null 2>&1 || HOMEBREW_NO_AUTO_UPDATE=1 brew install jq --quiet
        return $?

    elif cat /etc/issue | grep Debian > /dev/null 2>&1 || cat /etc/issue | grep Ubuntu > /dev/null 2>&1; then
        if [ "$(id -u)" = 0 ]; then export SUDO=""; else # Check if we're root
            export SUDO="sudo";
        fi
        echo "Installing JQ for Debian."
        $SUDO apt -qq update
        $SUDO apt -qq install -y jq
        return $?

    elif cat /etc/issue | grep Alpine > /dev/null 2>&1; then
        command -v curl >/dev/null 2>&1 || { echo >&2 "SLACK ORB ERROR: CURL is required. Please install."; exit 1; }
        command -v jq >/dev/null 2>&1 || { echo >&2 "SLACK ORB ERROR: JQ is required. Please install"; exit 1; }
        return $?
    fi
}

BranchFilter() {
    # If any pattern supplied matches the current branch, proceed; otherwise, exit with message.
    FLAG_MATCHES_FILTER="false"
    for i in $(echo "$SLACK_PARAM_BRANCHPATTERN" | sed "s/,/ /g")
    do
     if echo "$CIRCLE_BRANCH" | grep -Eq "^${i}$" ; then
        FLAG_MATCHES_FILTER="true"
        break
     fi
    done
    if [ "$FLAG_MATCHES_FILTER" = "false" ]; then
        # dont send message.
        echo "NO SLACK ALERT"
        echo
        echo 'Current branch does not match any item from the "branch_list" parameter'
        echo "Current branch: ${CIRCLE_BRANCH}"
        exit 0
    fi
}

# Will not run if sourced from another script.
# This is done so this script may be tested.
ORB_TEST_ENV="bats-core"
if [ "${0#*$ORB_TEST_ENV}" = "$0" ]; then
    . "/tmp/SLACK_JOB_STATUS"
    InstallJq
    BuildMessageBody
    Notify
fi
