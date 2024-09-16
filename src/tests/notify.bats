#!/usr/bin/env bats

setup() {
    source ./src/scripts/notify.sh
    source ./src/scripts/utils.sh
    export SLACK_PARAM_BRANCHPATTERN=$(cat $BATS_TEST_DIRNAME/sampleBranchFilters.txt)
    SLACK_PARAM_INVERT_MATCH="0"
}

@test "1: Skip message on no event" {
    CCI_STATUS="success"
    SLACK_PARAM_EVENT="fail"
    echo "Running notify"
    run ShouldPost
    echo "test output status: $status"
    echo "Output:"
    echo "$output"
    echo " --- "
    [ "$status" -eq 0 ] # Check for no exit error
    [[ $output == *"NO SLACK ALERT"* ]] # Ensure output contains expected string
}
@test "2: ModifyCustomTemplate" {
    # Ensure a custom template has the text key automatically affixed.
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplate.json)
    ModifyCustomTemplate
    TEXTKEY=$(echo $CUSTOM_BODY_MODIFIED | jq '.text')
    [ "$TEXTKEY" == '""' ]
}

@test "3: ModifyCustomTemplate with existing Text key" {
    # Ensure a custom template has the text key automatically affixed with original contents.
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplateWithText.json)
    ModifyCustomTemplate
    TEXTKEY=$(echo $CUSTOM_BODY_MODIFIED | jq '.text')
    [ "$TEXTKEY" == '"User-Added text key"' ]
}

@test "4: ModifyCustomTemplate with environment variable in link" {
    TESTLINKURL="http://circleci.com"
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplateWithLink.json)
    SLACK_DEFAULT_CHANNEL="xyz"
    BuildMessageBody
    EXPECTED=$(echo "{ \"blocks\": [ { \"type\": \"section\", \"text\": { \"type\": \"mrkdwn\", \"text\": \"Sample link using environment variable in markdown <${TESTLINKURL}|LINK >\" } } ], \"text\": \"\", \"channel\": \"$SLACK_DEFAULT_CHANNEL\" }" | jq)
    [ "$SLACK_MSG_BODY" == "$EXPECTED" ]
}

@test "5: ModifyCustomTemplate special chars" {
    TESTLINKURL="http://circleci.com"
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplateWithSpecialChars.json)
    SLACK_DEFAULT_CHANNEL="xyz"
    BuildMessageBody
    EXPECTED=$(echo "{ \"blocks\": [ { \"type\": \"section\", \"text\": { \"type\": \"mrkdwn\", \"text\": \"These asterisks are not \`glob\`  patterns **t** (parentheses'). [Link](https://example.org)\" } } ], \"text\": \"\", \"channel\": \"$SLACK_DEFAULT_CHANNEL\" }" | jq)
    [ "$SLACK_MSG_BODY" == "$EXPECTED" ]
}

@test "15: Sanitize - Escape newlines in environment variables" {
    CIRCLE_JOB="$(printf "%s\\n" "Line 1." "Line 2." "Line 3.")"
    EXPECTED="Line 1.\\nLine 2.\\nLine 3."
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplate.json)
    SanitizeVars "$SLACK_PARAM_CUSTOM"
    printf '%s\n' "Expected: $EXPECTED" "Actual: $CIRCLE_JOB"
    [ "$CIRCLE_JOB" = "$EXPECTED" ] # Newlines should be literal and escaped
}

@test "16: Sanitize - Escape double quotes in environment variables" {
    CIRCLE_JOB="$(printf "%s\n" "Hello \"world\".")"
    EXPECTED="Hello \\\"world\\\"."
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplate.json)
    SanitizeVars "$SLACK_PARAM_CUSTOM"
    printf '%s\n' "Expected: $EXPECTED" "Actual: $CIRCLE_JOB"
    [ "$CIRCLE_JOB" = "$EXPECTED" ] # Double quotes should be escaped
}

@test "17: Sanitize - Escape backslashes in environment variables" {
    CIRCLE_JOB="$(printf "%s\n" "removed extra '\' from  notification template")"
    EXPECTED="removed extra '\\\' from  notification template"
    SLACK_PARAM_CUSTOM=$(cat $BATS_TEST_DIRNAME/sampleCustomTemplate.json)
    SanitizeVars "$SLACK_PARAM_CUSTOM"
    printf '%s\n' "Expected: $EXPECTED" "Actual: $CIRCLE_JOB"
    [ "$CIRCLE_JOB" = "$EXPECTED" ] # Backslashes should be escaped
}

@test "18: Sanitize - Remove carriage returns in environment variables" {
    MESSAGE="$(cat $BATS_TEST_DIRNAME/sampleVariableWithCRLF.txt)"
    SLACK_PARAM_CUSTOM='{"text": "${MESSAGE}"}'
    SLACK_DEFAULT_CHANNEL="xyz"
    BuildMessageBody

    EXPECTED=$(echo "{ \"text\": \"Multiline Message 1\nMultiline Message 2\", \"channel\": \"$SLACK_DEFAULT_CHANNEL\" }" | jq)
    printf '%s\n' "Expected: $EXPECTED" "Actual: $SLACK_MSG_BODY"
    [ "$SLACK_MSG_BODY" = "$EXPECTED" ] # CRLF should be escaped
}
