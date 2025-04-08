#!/usr/bin/env bash
# shellcheck disable=SC2016,SC3043
if [[ "$SLACK_PARAM_CUSTOM" == \$* ]]; then
    echo "Doing substitution custom"
    SLACK_PARAM_CUSTOM="$(eval echo "${SLACK_PARAM_CUSTOM}" | circleci env subst)"
fi
if [[ "$SLACK_PARAM_TEMPLATE" == \$* ]]; then
    echo "Doing substitution template"
    SLACK_PARAM_TEMPLATE="$(eval echo "${SLACK_PARAM_TEMPLATE}" | circleci env subst)"
fi

if [ "$SLACK_PARAM_DEBUG" -eq 1 ]; then
    set -x
fi

# Import utils.
eval "$SLACK_SCRIPT_UTILS"
JQ_PATH=/usr/local/bin/jq

BuildMessageBody() {
    # Send message
    #   If sending message, default to custom template,
    #   if none is supplied, check for a pre-selected template value.
    #   If none, error.
    if [ -n "${SLACK_PARAM_CUSTOM:-}" ]; then
        SanitizeVars "$SLACK_PARAM_CUSTOM"
        ModifyCustomTemplate
        # shellcheck disable=SC2016
        CUSTOM_BODY_MODIFIED=$(echo "$CUSTOM_BODY_MODIFIED" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed 's/`/\\`/g')
        T2="$(eval printf '%s' \""$CUSTOM_BODY_MODIFIED"\")"
    else
        # shellcheck disable=SC2154
        if [ -n "${SLACK_PARAM_TEMPLATE:-}" ]; then
            TEMPLATE="\$$SLACK_PARAM_TEMPLATE"
        elif [ "$CCI_STATUS" = "pass" ]; then
            TEMPLATE="\$basic_success_1"
        elif [ "$CCI_STATUS" = "fail" ]; then
            TEMPLATE="\$basic_fail_1"
        else
            echo "A template wasn't provided nor is possible to infer it based on the job status. The job status: '$CCI_STATUS' is unexpected."
            exit 1
        fi

        [ -z "${SLACK_PARAM_TEMPLATE:-}" ] && echo "No message template was explicitly chosen. Based on the job status '$CCI_STATUS' the template '$TEMPLATE' will be used."
        template_body="$(eval printf '%s' \""$TEMPLATE\"")"
        SanitizeVars "$template_body"

        # shellcheck disable=SC2016
        T1="$(printf '%s' "$template_body" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed 's/`/\\`/g')"
        T2="$(eval printf '%s' \""$T1"\")"
    fi

    # Insert the default channel. THIS IS TEMPORARY
    T2="$(printf '%s' "$T2" | jq ". + {\"channel\": \"$SLACK_DEFAULT_CHANNEL\"}")"
    SLACK_MSG_BODY="$T2"
}

NotifyWithRetries() {
    local success_request=false
    local retry_count=0
    while [ "$retry_count" -le "$SLACK_PARAM_RETRIES" ]; do
        if SLACK_SENT_RESPONSE=$(curl -s -f -X POST -H 'Content-type: application/json; charset=utf-8' -H "Authorization: Bearer $SLACK_ACCESS_TOKEN" --data "$SLACK_MSG_BODY" "$1"); then
            echo "Notification sent"
            success_request=true
            break
        else
            echo "Error sending notification. Retrying..."
            retry_count=$((retry_count + 1))
            sleep "$SLACK_PARAM_RETRY_DELAY"
        fi
    done
    if [ "$success_request" = false ]; then
        echo "Error sending notification. Max retries reached"
        exit 1
    fi
}

PostToSlack() {
    # Post once per channel listed by the channel parameter
    #    The channel must be modified in SLACK_MSG_BODY
    #    If thread_id is used a file containing the initial message `thread_ts` per each channel is persisted
    #    /tmp/SLACK_THREAD_INFO/<channel_name> will contain:
    #    <thread_id>=12345.12345

    # shellcheck disable=SC2001
    for i in $(eval echo \""$SLACK_PARAM_CHANNEL"\" | sed "s/,/ /g"); do
        # replace non-alpha
        SLACK_PARAM_THREAD=$(echo "$SLACK_PARAM_THREAD" | sed -r 's/[^[:alpha:][:digit:].]/_/g')
        # check if the invoked `notify` command is intended to post threaded messages &
        # check for persisted thread info file for each channel listed in channel parameter
        if [ ! "$SLACK_PARAM_THREAD" = "" ] && [ -f "/tmp/SLACK_THREAD_INFO/$i" ]; then
            # get the initial message thread_ts targeting the correct channel and thread id
            # || [ "$?" = "1" ] - this is used to avoid exit status 1 if grep doesn't match anything
            # shellcheck disable=SC2002
            SLACK_THREAD_EXPORT=$(grep -m1 "$SLACK_PARAM_THREAD" /tmp/SLACK_THREAD_INFO/"$i" || [ "$?" = "1" ])
            if [ ! "$SLACK_THREAD_EXPORT" = "" ]; then
                # if there is an initial message with a thread id, load it into the environment
                # thread_id=12345.12345
                eval "$SLACK_THREAD_EXPORT"
            fi
            # get the value of the specified thread from the environment
            # SLACK_THREAD_TS=12345.12345
            SLACK_THREAD_TS=$(eval "echo \"\$$SLACK_PARAM_THREAD\"")
            if [ $SLACK_PARAM_UPDATE ]; then
                # when updating, the key ts is used to reference the message to be updated
                TS=$(eval "echo \"\$$SLACK_PARAM_THREAD\"")
                SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg ts "$TS" '.ts = $ts')
                SLACK_UPDATE=true
            else
                # append the thread_ts to the body for posting the message in the correct thread
                SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg thread_ts "$SLACK_THREAD_TS" '.thread_ts = $thread_ts')
            fi
        fi

        echo "Sending to Slack Channel: $i"
        SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg channel "$i" '.channel = $channel')
        if [ "$SLACK_PARAM_DEBUG" -eq 1 ]; then
            printf "%s\n" "$SLACK_MSG_BODY" >"$SLACK_MSG_BODY_LOG"
            echo "The message body being sent to Slack can be found below. To view redacted values, rerun the job with SSH and access: ${SLACK_MSG_BODY_LOG}"
            echo "$SLACK_MSG_BODY"
        fi

        if [ "${SLACK_PARAM_OFFSET:0}" -ne 0 ]; then
            if date --version >/dev/null 2>&1; then
                # GNU date function
                POST_AT=$(date -d "now + ${SLACK_PARAM_OFFSET} seconds" +%s)
            elif date -v+1S >/dev/null 2>&1; then
                # BSD date function
                POST_AT=$(date -v"+${SLACK_PARAM_OFFSET}S" +%s)
            else
                # Alpine
                POST_AT=$(date -u +%s | awk -v sec="$SLACK_PARAM_OFFSET" '{print $1 + sec}')
            fi
            SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq --arg post_at "$POST_AT" '.post_at = ($post_at|tonumber)')
            # text is required for scheduled messages
            SLACK_MSG_BODY=$(echo "$SLACK_MSG_BODY" | jq '.text = "Dummy fallback text"')
            
            NotifyWithRetries https://slack.com/api/chat.scheduleMessage
        elif [ $SLACK_UPDATE ]; then
            NotifyWithRetries https://slack.com/api/chat.update
        else
            NotifyWithRetries https://slack.com/api/chat.postMessage
        fi

        if [ "$SLACK_PARAM_DEBUG" -eq 1 ]; then
            printf "%s\n" "$SLACK_SENT_RESPONSE" >"$SLACK_SENT_RESPONSE_LOG"
            echo "The response from the API call to Slack can be found below. To view redacted values, rerun the job with SSH and access: ${SLACK_SENT_RESPONSE_LOG}"
            echo "$SLACK_SENT_RESPONSE"
        fi

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

        # check if the invoked `notify` command is intended to post messages in threads
        if [ ! "$SLACK_PARAM_THREAD" = "" ]; then
            # get message thread_ts from response
            SLACK_THREAD_TS=$(echo "$SLACK_SENT_RESPONSE" | jq '.ts')
            if [ ! "$SLACK_THREAD_TS" = "null" ]; then
                # store the thread_ts in the specified channel for the specified thread_id
                mkdir -p /tmp/SLACK_THREAD_INFO
                echo "$SLACK_PARAM_THREAD=$SLACK_THREAD_TS" >>/tmp/SLACK_THREAD_INFO/"$i"
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
    echo "Checking For JQ + CURL"
    if command -v curl >/dev/null 2>&1 && ! command -v jq >/dev/null 2>&1; then
        uname -a | grep Darwin >/dev/null 2>&1 && JQ_VERSION=jq-osx-amd64 || JQ_VERSION=jq-linux32
        curl -Ls -o "$JQ_PATH" https://github.com/stedolan/jq/releases/download/jq-1.6/"${JQ_VERSION}"
        chmod +x "$JQ_PATH"
        command -v jq >/dev/null 2>&1
        return $?
    else
        command -v curl >/dev/null 2>&1 || {
            echo >&2 "SLACK ORB ERROR: CURL is required. Please install."
            exit 1
        }
        command -v jq >/dev/null 2>&1 || {
            echo >&2 "SLACK ORB ERROR: JQ is required. Please install"
            exit 1
        }
        return $?
    fi
}

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

SetupEnvVars() {
    echo "BASH_ENV file: $BASH_ENV"
    if [ -f "$BASH_ENV" ]; then
        echo "Exists. Sourcing into ENV"
        # shellcheck disable=SC1090
        . "$BASH_ENV"
    else
        echo "Does Not Exist. Skipping file execution"
    fi
}

CheckEnvVars() {
    if [ -n "${SLACK_WEBHOOK:-}" ]; then
        echo "It appears you have a Slack Webhook token present in this job."
        echo "Please note, Webhooks are no longer used for the Slack Orb (v4 +)."
        echo "Follow the setup guide available in the wiki: https://github.com/CircleCI-Public/slack-orb/wiki/Setup"
    fi
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

SetupLogs() {
    if [ "$SLACK_PARAM_DEBUG" -eq 1 ]; then
        LOG_PATH="$(mktemp -d 'slack-orb-logs.XXXXXX')"
        SLACK_MSG_BODY_LOG="$LOG_PATH/payload.json"
        SLACK_SENT_RESPONSE_LOG="$LOG_PATH/response.json"

        touch "$SLACK_MSG_BODY_LOG" "$SLACK_SENT_RESPONSE_LOG"
        chmod 0600 "$SLACK_MSG_BODY_LOG" "$SLACK_SENT_RESPONSE_LOG"
    fi
}

# $1: Template with environment variables to be sanitized.
SanitizeVars() {
    [ -z "$1" ] && {
        printf '%s\n' "Missing argument."
        return 1
    }
    local template="$1"

    # Find all environment variables in the template with the format $VAR or ${VAR}.
    # The "|| true" is to prevent bats from failing when no matches are found.
    local variables
    variables="$(printf '%s\n' "$template" | grep -o -E '\$\{?[a-zA-Z_0-9]*\}?' || true)"
    [ -z "$variables" ] && {
        printf '%s\n' "Nothing to sanitize."
        return 0
    }

    # Extract the variable names from the matches.
    local variable_names
    variable_names="$(printf '%s\n' "$variables" | grep -o -E '[a-zA-Z0-9_]+' || true)"
    [ -z "$variable_names" ] && {
        printf '%s\n' "Nothing to sanitize."
        return 0
    }

    # Find out what OS we're running on.
    detect_os

    for var in $variable_names; do
        # The variable must be wrapped in double quotes before the evaluation.
        # Otherwise the newlines will be removed.
        local value
        value="$(eval printf '%s' \"\$"$var\"")"
        [ -z "$value" ] && {
            printf '%s\n' "$var is empty or doesn't exist. Skipping it..."
            continue
        }

        printf '%s\n' "Sanitizing $var..."

        local sanitized_value="$value"
        # Escape backslashes.
        sanitized_value="$(printf '%s' "$sanitized_value" | awk '{gsub(/\\/, "&\\"); print $0}')"
        # Escape newlines.
        sanitized_value="$(printf '%s' "$sanitized_value" | tr -d '\r' | awk 'NR > 1 { printf("\\n") } { printf("%s", $0) }')"
        # Escape double quotes.
        if [ "$PLATFORM" = "windows" ]; then
            sanitized_value="$(printf '%s' "$sanitized_value" | awk '{gsub(/"/, "\\\""); print $0}')"
        else
            sanitized_value="$(printf '%s' "$sanitized_value" | awk '{gsub(/\"/, "\\\""); print $0}')"
        fi

        # Write the sanitized value back to the original variable.
        # shellcheck disable=SC3045 # This is working on Alpine.
        printf -v "$var" "%s" "$sanitized_value"
    done

    return 0
}

# Will not run if sourced from another script.
# This is done so this script may be tested.
ORB_TEST_ENV="bats-core"
if [ "${0#*"$ORB_TEST_ENV"}" = "$0" ]; then
    # shellcheck source=/dev/null
    . "/tmp/SLACK_JOB_STATUS"
    ShouldPost
    SetupEnvVars
    SetupLogs
    CheckEnvVars
    InstallJq
    BuildMessageBody
    PostToSlack
fi
set +x
