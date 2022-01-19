setup() {
    source ./src/scripts/notify.sh
    export SLACK_PARAM_BRANCHPATTERN=$(cat $BATS_TEST_DIRNAME/sampleBranchFilters.txt)
}

setupModifyMentions() {
    T2=$(cat $BATS_TEST_DIRNAME/sampleBasicFail.json)
    T2=$(echo $T2 | jq -r --arg mentions "*Mentions*: $1" \
            '(.blocks[] | select(.type == "section").fields[].text | select(contains("Mentions"))) = $mentions')
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

@test "13: ModifyMentions - Should replace environment variable with its value" {
    TEMPLATE_MENTIONS_PARAM="\$MY_MENTIONS"
    MY_MENTIONS="I should show up in the template!"

    setupModifyMentions "$TEMPLATE_MENTIONS_PARAM"

    echo "T2 pre-function: $T2"
    ModifyMentions
    echo "T2 post-function: $T2"

    MENTIONS_POST_FUNCTION=$(echo $T2 | jq -r '.blocks[] | select(.type == "section").fields[].text | select(contains("Mentions"))' | sed -e "s/\*Mentions\*: //")
    echo "Mentions post-function: $MENTIONS_POST_FUNCTION" 
    
    [[ $MENTIONS_POST_FUNCTION == $MY_MENTIONS ]]
}

@test "14: ModifyMentions - Should leave the mentions untouched" {
    TEMPLATE_MENTIONS_PARAM="Hello there! I should not change!"

    setupModifyMentions "$TEMPLATE_MENTIONS_PARAM"

    MENTIONS_PRE_FUNCTION=$(echo $T2 | jq -r '.blocks[] | select(.type == "section").fields[].text | select(contains("Mentions"))' | sed -e "s/\*Mentions\*: //")
    echo "Mentions pre-function: $MENTIONS_PRE_FUNCTION"

    echo "T2 pre-function: $T2"
    ModifyMentions
    echo "T2 post-function: $T2"

    MENTIONS_POST_FUNCTION=$(echo $T2 | jq -r '.blocks[] | select(.type == "section").fields[].text | select(contains("Mentions"))' | sed -e "s/\*Mentions\*: //")
    echo "Mentions post-function: $MENTIONS_POST_FUNCTION" 
    
    [[ $MENTIONS_PRE_FUNCTION == $MENTIONS_POST_FUNCTION ]]
}

@test "15: ModifyMentions - Mixing variables and text should work" {
    TEMPLATE_MENTIONS_PARAM="Hello there! Pinging \$MY_MENTIONS to take a look at this!"
    MY_MENTIONS="@orbs and @potato"
    EXPECTED_MENTIONS_POST_FUNCTION="Hello there! Pinging @orbs and @potato to take a look at this!"

    setupModifyMentions "$TEMPLATE_MENTIONS_PARAM"

    echo "T2 pre-function: $T2"
    ModifyMentions
    echo "T2 post-function: $T2"

    MENTIONS_POST_FUNCTION=$(echo $T2 | jq -r '.blocks[] | select(.type == "section").fields[].text | select(contains("Mentions"))' | sed -e "s/\*Mentions\*: //")
    echo "Mentions post-function: $MENTIONS_POST_FUNCTION" 
    
    [[ $MENTIONS_POST_FUNCTION == $EXPECTED_MENTIONS_POST_FUNCTION ]]
}