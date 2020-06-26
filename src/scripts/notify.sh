echo "Sending Notification"

if [[ "$CCI_STATUS" == "<<parameters.event>>" || "<<parameters.event>>" == "always" ]]; then
    # send message
    # If sending message, default to custom template, if none is supplied, check for a pre-selected template value. If none, error.

    if [ -n "$SLACK_PARAM_CUSTOM" ]; then
        SLACK_MSG_BODY="$SLACK_PARAM_CUSTOM"
    elif [ -n "$SLACK_PARAM_TEMPLATE" ]; then
        SLACK_MSG_BODY="$SLACK_PARAM_TEMPLATE"
    else
        echo "Error: No message template selected."
        echo "Select either a custom template or one of the pre-included ones via the 'custom' or 'template' parameters."
        exit 0
    fi

    curl -X POST -H 'Content-type: application/json' \
        --data \
        "$SLACK_MSG_BODY" "<<parameters.webhook>>"
else
    # dont send message.
    echo "NO SLACK ALERT"
    echo
    echo "This command is set to send an alert on: <<parameters.event>>"
    echo "Current status: $CCI_STATUS"
    exit 0
fi
