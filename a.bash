
SLACK_PARAM_CUSTOM=$(cat b.json)

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
        echo $CUSTOM_BODY_MODIFIED
        T2="$(eval printf '%s' \""$CUSTOM_BODY_MODIFIED"\")"
        echo $T2
    fi

    # Insert the default channel. THIS IS TEMPORARY
    T2="$(printf '%s' "$T2" | jq ". + {\"channel\": \"$SLACK_DEFAULT_CHANNEL\"}")"
    SLACK_MSG_BODY="$T2"
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

# $1: Template with environment variables to be sanitized.
SanitizeVars() {

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

BuildMessageBody