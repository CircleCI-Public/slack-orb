#!/usr/bin/env bats

# load custom assertions and functions
load bats_helper

# setup is run before each test
function setup {
    INPUT_PROJECT_CONFIG=${BATS_TMPDIR}/input_config-${BATS_TEST_NUMBER}
    PROCESSED_PROJECT_CONFIG=${BATS_TMPDIR}/packed_config-${BATS_TEST_NUMBER} 
    JSON_PROJECT_CONFIG=${BATS_TMPDIR}/json_config-${BATS_TEST_NUMBER} 
    echo "#using temp file ${BATS_TMPDIR}/"

  # the name used in example config files.
  INLINE_ORB_NAME="slack"

}

@test "1: Basic expansion works" {
  # given
  process_config_with tests/cases/simple.yml

  # when
  assert_jq_match '.jobs | length' 1 #only 1 job

}