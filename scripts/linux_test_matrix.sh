#!/bin/bash

# This script tests the built flows2fim executable's methods with a combination of parameters
usage() {
    echo "
    This script tests the built flows2fim executable's methods with a combination of parameters.
    It must be run from the root directory of the flows2fim repository to ensure proper pathing for 
    test data.

    Usage: ./scripts/linux_test_matrix.sh 

    Providing no options, or 'all', will test all flows2fim methods- controls, fim, & validate.

    OPTIONS:
        controls: Only issue the controls tests. 
            e.g.: ./scripts/linux_test_matrix.sh controls
        fim: Only issue the fim tests. 
            e.g.: ./scripts/linux_test_matrix.sh fim
        validate: Only issue the validate tests. 
            e.g.: ./scripts/linux_test_matrix.sh validate
    "
}

# Check if this script is being issued from the repo's root to ensure consistent filepaths
if [ "$(git rev-parse --show-toplevel)" != "$(pwd)" ]; then
    echo "Error: This script must be run from the root directory of the flows2fim repository."
    usage
    exit 1
fi

# Ensure flows2fim is installed, available, and executable
if flows2fim --version > /dev/null; then
    printf "\n\t>>>> Testing availability of flows2fim executable. Output of flows2fim --version: <<<<\n\n"
    flows2fim --version
else
    echo "flows2fim is not available"
    usage
    exit 1
fi


test_help_statement() {
    printf "\n\t>>>> Test the --help usage statement <<<<\n\n" 
    flows2fim --help
}

# Define functions for each flows2fim method
controls_test_cases() {
    num_test_cases_controls=4
    printf "\n\t###### TEST CASES FOR flows2fim controls  ######\n\n"
    printf "(1/${num_test_cases_controls})\n\t>>>> Test controls  <<<<\n\n"
    # flows2fim controls -db

}

fim_test_cases() {
    num_test_cases_fim=4
    printf "\n\t###### TEST CASES FOR flows2fim fim  ######\n\n"
    printf "(1/${num_test_cases_fim})\n\t>>>> Test fim <<<<\n\n"
    # flows2fim fim 

}

validate_test_cases() {
    num_test_cases_validate=4
    printf "\n\t###### TEST CASES FOR flows2fim validate  ######\n\n"
    printf "(1/${num_test_cases_validate})\n\t>>>> Test validate <<<<\n\n"
    # flows2fim validate

}

# Assign the command line argument if given, else 'all' is the default
method=${1:-all}

# Control flow for methods and test cases
if [ "$method" = "all" ]; then
    printf "\n\t>>>> TESTING ALL METHODS <<<<\n\n"
    test_help_statement
    controls_test_cases
    fim_test_cases
    validate_test_cases
elif [ "$method" = "controls" ]; then
    controls_test_cases

elif [ "$method" = "fim" ]; then
    fim_test_cases

elif [ "$method" = "validate" ]; then
    validate_test_cases
else
    printf "\n\t>>>> NO TEST CASES WERE RUN <<<<\n\n"
    usage
fi