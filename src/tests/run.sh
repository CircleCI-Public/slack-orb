#!/usr/bin/env sh
# shellcheck disable=SC2046,SC2115,2155

export basic_success_1="$(cat ../message_templates/basic_success_1.json)"
export basic_fail_1="$(cat ../message_templates/basic_fail_1.json)"
export basic_on_hold_1="$(cat ../message_templates/basic_on_hold_1.json)"
export success_tagged_deploy_1="$(cat ../message_templates/success_tagged_deploy_1.json)"
export test_template="$(cat ../scripts/test_template.json)"
export DOUBLE_QUOTES_STRING=$(printf "%s\\n" "Hello There! My name is \"Potato\"")
export MULTILINE_STRING=$(printf "%s\\n" "Line 1." "Line 2." "Line 3.")
export $(grep -v '^#' .env | xargs)

go run ../scripts/main.go