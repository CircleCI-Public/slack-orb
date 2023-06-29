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

@test "6: FilterBy - match-all default" {
    SLACK_PARAM_BRANCHPATTERN=".+"
    CIRCLE_BRANCH="xyz-123"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error SLACK_PARAM_BRANCHPATTERN debug: $SLACK_PARAM_BRANCHPATTERN"
    echo "Error output debug: $output"
    [ "$output" == "" ] # Should match any branch: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "7: FilterBy - string" {
    CIRCLE_BRANCH="master"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [ "$output" == "" ] # "master" is in the list: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "8: FilterBy - regex numbers" {
    CIRCLE_BRANCH="pr-123"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [ "$output" == "" ] # "pr-[0-9]+" should match: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "9: FilterBy - non-match" {
    CIRCLE_BRANCH="x"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [[ "$output" =~ "NO SLACK ALERT" ]] # "x" is not included in the filter. Error message expected.
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "10: FilterBy - no partial-match" {
    CIRCLE_BRANCH="pr-"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [[ "$output" =~ "NO SLACK ALERT" ]] # Filter dictates that numbers should be included. Error message expected.
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "11: FilterBy - SLACK_PARAM_BRANCHPATTERN is empty" {
    unset SLACK_PARAM_BRANCHPATTERN
    CIRCLE_BRANCH="master"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "12: FilterBy - CIRCLE_BRANCH is empty" {
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "13: FilterBy - match and SLACK_PARAM_INVERT_MATCH is set" {
    CIRCLE_BRANCH="pr-123"
    SLACK_PARAM_INVERT_MATCH="1"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [[ "$output" =~ "NO SLACK ALERT" ]] # "pr-[0-9]+" should match but inverted: Error message expected.
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "14: FilterBy - non-match and SLACK_PARAM_INVERT_MATCH is set" {
    CIRCLE_BRANCH="foo"
    SLACK_PARAM_INVERT_MATCH="1"
    run FilterBy "$SLACK_PARAM_BRANCHPATTERN" "$CIRCLE_BRANCH"
    echo "Error output debug: $output"
    [ "$output" == "" ] # Nothing should match but inverted: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
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
