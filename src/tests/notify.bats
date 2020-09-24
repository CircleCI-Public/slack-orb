setup() {
    source ./src/scripts/notify.sh
    export INTRNL_SLACK_WEBHOOK="x"
    export SLACK_PARAM_BRANCHPATTERN=$(cat $BATS_TEST_DIRNAME/sampleBranchFilters.txt)
}

@test "1: Skip message on no event" {
    CCI_STATUS="success"
    SLACK_PARAM_EVENT="fail"
    echo "Running notify"
    run Notify
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

@test "5: Branch Filter - case 1 | match-all default" {
    SLACK_PARAM_BRANCHPATTERN=".+"
    CIRCLE_BRANCH="xyz-123"
    run BranchFilter
    echo "Error SLACK_PARAM_BRANCHPATTERN debug: $SLACK_PARAM_BRANCHPATTERN"
    echo "Error output debug: $output"
    [ "$output" == "" ] # Should match any branch: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "6: Branch Filter - case 2 | string" {
    CIRCLE_BRANCH="master"
    run BranchFilter
    echo "Error output debug: $output"
    [ "$output" == "" ] # "master" is in the list: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "7: Branch Filter - case 3 | regex numbers" {
    CIRCLE_BRANCH="pr-123"
    run BranchFilter
    echo "Error output debug: $output"
    [ "$output" == "" ] # "pr-[0-9]+" should match: No output error
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "8: Branch Filter - case 4 | non-match" {
    CIRCLE_BRANCH="x"
    run BranchFilter
    echo "Error output debug: $output"
    [[ "$output" =~ "NO SLACK ALERT" ]] # "x" is not included in the filter. Error message expected.
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}

@test "9: Branch Filter - case 5 | no partial-match" {
    CIRCLE_BRANCH="pr-"
    run BranchFilter
    echo "Error output debug: $output"
    [[ "$output" =~ "NO SLACK ALERT" ]] # Filter dictates that numbers should be included. Error message expected.
    [ "$status" -eq 0 ] # In any case, this should return a 0 exit as to not block a build/deployment.
}