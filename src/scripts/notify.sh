echo "Sending Notification"

SetEnvVars() {
    INTRNL_SLACK_WEBHOOK=$(eval echo "$SLACK_PARAM_WEBHOOK")
}

BuildMessageBody() {
    # Send message
    #   If sending message, default to custom template,
    #   if none is supplied, check for a pre-selected template value.
    #   If none, error.
    if [ -n "$SLACK_PARAM_CUSTOM" ]; then
        SLACK_MSG_BODY="$SLACK_PARAM_CUSTOM"
    elif [ -n "$SLACK_PARAM_TEMPLATE" ]; then
        SLACK_MSG_BODY="$SLACK_PARAM_TEMPLATE"
    else
        echo "Error: No message template selected."
        echo "Select either a custom template or one of the pre-included ones via the 'custom' or 'template' parameters."
        exit 1
    fi
}

PostToSlack() {
    curl -X POST -H 'Content-type: application/json' \
        --data \
        "$SLACK_MSG_BODY" "$INTRNL_SLACK_WEBHOOK"
}

Notify() {
    if [[ "$CCI_STATUS" == "$SLACK_PARAM_EVENT" || "$SLACK_PARAM_EVENT" == "always" ]]; then
    PostToSlack
    else
        # dont send message.
        echo "NO SLACK ALERT"
        echo
        echo "This command is set to send an alert on: $SLACK_PARAM_EVENT"
        echo "Current status: $CCI_STATUS"
        exit 0
    fi
}

# Will not run if sourced from another script.
# This is done so this script may be tested.
if [[ "$_" == "$0" ]]; then
    SetEnvVars
    BuildMessageBody
    Notify
fi
