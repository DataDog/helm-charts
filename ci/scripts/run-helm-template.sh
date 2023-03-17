#!/bin/bash

# boolean: indicating whether the operation will produce allure-results or not
ALLURE_RESULTS=false

# boolean: indicating whether or not the operation passed or failed (set to the exit code of the operation)
# Set to fail by default (i.e. non-zero exit code)
EXECUTION_PASSED=1

# This function runs the tests.  It needs to set EXECUTION_PASSED to true if the tests pass or false otherwise.
function run_template() {
   helm template datadog charts/datadog
   EXECUTION_PASSED=$?
}

# Record the results for the parent workflow to read.
function record_results() {
   echo $EXECUTION_PASSED > /results/exit-code
   echo $ALLURE_RESULTS > /results/allure-enabled
}

function main() {
   run_template
   record_results
}

main "$@"
