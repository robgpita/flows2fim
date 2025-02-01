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

# Set paths
data_folder=testdata
benchmark_data=$data_folder/benchmark_data
collection_data_path=$data_folder/mip_17110008
flows_dir=$collection_data_path/flows
library_dir=$collection_data_path/library
library_extent_dir=$collection_data_path/library_extent
ref_dir=$data_folder/reference_data
test_outputs=$data_folder/test_out
control_test_outputs=$test_outputs/controls

controls_benchmark=$benchmark_data/controls


test_help_statement() {
    printf "\n\t>>>> Test the --help usage statement <<<<\n\n" 
    flows2fim --help
}

compare_directories() {
    # Set args
    local dir1=$1
    local dir2=$2
    # Loop through files in dir1 and compare with dir2
    for file in "$dir1"/*; do
        filename=$(basename "$file")
        filepath2="$dir2/$filename"
        # Check if the file exists in dir2
        if [ -f "$filepath2" ]; then
            # Compare the files
            if ! cmp -s "$file" "$filepath2"; then
                echo "Files differ: $filename"
            fi
        else
            echo "File not found in dir2: $filename"
        fi
    done
    # Loop through files in dir2 to find any not in dir1
    for file in "$dir2"/*; do
        filename=$(basename "$file")
        filepath1="$dir1/$filename"
        # Check if the file does not exist in dir1
        if [ ! -f "$filepath1" ]; then
            echo "File not found in $dir1: $filename"
        fi
    done
}

# Define functions for each flows2fim method
controls_test_cases() {
    num_test_cases_controls=5
    # If previous directory exists, remove it
    if [ -d "$control_test_outputs" ]; then
        rm -rf "$control_test_outputs"
    fi
    # Create new test output directory
    mkdir -p $control_test_outputs
    
    printf "\n\t###### TEST CASES FOR flows2fim controls  ######\n\n"

    # Define the recurrence interval array
    local recurrence_interval=(2 5 10 25 50 100)
    printf "(1/${num_test_cases_controls})\n\t>>>> Generate controls.csv files from "${recurrence_interval[@]}""
    printf " recurrence interval data <<<<\n\n"
    
        for interval in "${recurrence_interval[@]}"; do
            # Execute flows2fim with the current interval
            echo "Executing flows2fim with interval: $interval"
            flows2fim controls -db $collection_data_path/ripple.gpkg \
                -f $ref_dir/flows_${interval}year.csv \
                -o $control_test_outputs/controls_${interval}year.csv \
                -scsv $collection_data_path/start_reaches.csv
        done
   
    # Compare controls files from recently generated test data and benchmark data
    printf "(2/${num_test_cases_controls})\n\t>>>> Comparing controls files <<<<\n\n"
    
    # Capture the output of the comparion as a variable if there was no differnce
    diff_output=$(compare_directories "$control_test_outputs" "$controls_benchmark")

    if [ -z "$diff_output" ]; then
        printf "No difference in controls files. \n"
    else
        printf "$diff_output \n"
    fi

    
    printf "(3/${num_test_cases_controls})\n\t>>>> Assert Error thrown from no start reaches file<<<<\n\n"
        # Create and assign temp file
        tempfile=$(mktemp)
        # Test case
        flows2fim controls -db $collection_data_path/ripple.gpkg \
            -f $ref_dir/flows_2year.csv \
            -o $control_test_outputs/controls_2year.csv &> "$tempfile"
        # Here we use head -n 1 to capture the first line redirected to the tmp file
        first_line=$(head -n 1 "$tempfile")
        # Remove temp file 
        rm "$tempfile"
        # Assign error string
        assert_expected_output_1="Error: either a CSV file or start reach IDs and control stages must be provided"
        # Compare Error messaging and print
        if [ "$first_line" = "$assert_expected_output_1" ]; then
            printf "Correct error thrown! \n"
        else
            printf "Error messaging inconsistent \n"
        fi    

    printf "(4/${num_test_cases_controls})\n\t>>>> Assert Error thrown from no start reaches file, or output file<<<<\n\n"
        # Create and assign temp file
        tempfile=$(mktemp)
        # Test case
        flows2fim controls -db $collection_data_path/ripple.gpkg -f $ref_dir/flows_2year.csv &> "$tempfile"
        # Here we use head -n 1 to capture the first line redirected to the tmp file
        first_line=$(head -n 1 "$tempfile")
        # Remove temp file 
        rm "$tempfile"
        # Assign error string
        assert_expected_output_2="Missing required flags"
        # Compare Error messaging and print
        if [ "$first_line" = "$assert_expected_output_2" ]; then
            printf "Correct error thrown! \n"
        else
            printf "Error messaging inconsistent \n"
        fi  

    printf "(5/${num_test_cases_controls})\n\t>>>> Assert Error thrown from only providing db file<<<<\n\n"
        # Create and assign temp file
        tempfile=$(mktemp)
        # Test case
        flows2fim controls -db $collection_data_path/ripple.gpkg &> "$tempfile"
        # Here we use head -n 1 to capture the first line redirected to the tmp file
        first_line=$(head -n 1 "$tempfile")
        # Remove temp file 
        rm "$tempfile"
        # Assign error string
        assert_expected_output_3="Missing required flags"
        # Compare Error messaging and print
        if [ "$first_line" = "$assert_expected_output_3" ]; then
            printf "Correct error thrown \n"
        else
            printf "Error messaging inconsistent \n"
        fi  

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
    printf "\n\t>>>> TESTING ALL METHODS <<<<\n"
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