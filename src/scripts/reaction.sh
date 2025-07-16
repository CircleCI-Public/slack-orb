#!/usr/bin/env bash
if [ "$SLACK_PARAM_DEBUG" = 1 ]; then
    set -x
fi

SLACK_PARAM_CHANNEL="$(echo "${SLACK_PARAM_CHANNEL}" | circleci env subst)"

ReactToSlack() {
    # replace non-alpha
    SLACK_PARAM_THREAD=$(echo "$SLACK_PARAM_THREAD" | sed -r 's/[^[:alpha:][:digit:].]/_/g')

    if [ ! "$SLACK_PARAM_THREAD" = "" ] && [ -f "/tmp/SLACK_THREAD_INFO/$SLACK_PARAM_CHANNEL" ]; then
        # shellcheck disable=SC2002
        SLACK_THREAD_EXPORT=$(grep -m1 "$SLACK_PARAM_THREAD" /tmp/SLACK_THREAD_INFO/"${SLACK_PARAM_CHANNEL}" || [ "$?" = "1" ])
        if [ ! "$SLACK_THREAD_EXPORT" = "" ]; then
            # if there is an initial message with a thread id, load it into the environment
            # thread_id=12345.12345
            eval "$SLACK_THREAD_EXPORT"
        fi
        # get the value of the specified thread from the environment
        # SLACK_THREAD_TS=12345.12345
        SLACK_THREAD_TS=$(eval "echo \"\$$SLACK_PARAM_THREAD\"")
    fi

    if [ -n "${SLACK_PARAM_REMOVE_REACT_NAME}" ]; then
        echo "Remove reaction with name=${SLACK_PARAM_REMOVE_REACT_NAME} channel=${SLACK_PARAM_CHANNEL} thread_ts=${SLACK_THREAD_TS}"
        curl -X POST --location 'https://slack.com/api/reactions.remove' \
            --header 'Content-Type: application/x-www-form-urlencoded' \
            --header "Authorization: Bearer ${SLACK_ACCESS_TOKEN}" \
            --data-urlencode "channel=${SLACK_PARAM_CHANNEL}" \
            --data-urlencode "name=${SLACK_PARAM_REMOVE_REACT_NAME}" \
            --data-urlencode "timestamp=${SLACK_THREAD_TS}"
        echo
    fi
    if [ -n "${SLACK_PARAM_ADD_REACT_NAME}" ]; then
        echo "Add reaction with name=${SLACK_PARAM_ADD_REACT_NAME} channel=${SLACK_PARAM_CHANNEL} thread_ts=${SLACK_THREAD_TS}"
        curl -X POST --location 'https://slack.com/api/reactions.add' \
            --header 'Content-Type: application/x-www-form-urlencoded' \
            --header "Authorization: Bearer ${SLACK_ACCESS_TOKEN}" \
            --data-urlencode "channel=${SLACK_PARAM_CHANNEL}" \
            --data-urlencode "name=${SLACK_PARAM_ADD_REACT_NAME}" \
            --data-urlencode "timestamp=${SLACK_THREAD_TS}"
        echo
    fi
}

ReactToSlack
set +x