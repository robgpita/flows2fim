#!/usr/bin/env bash

#       Windows-oriented test harness for flows2fim
#       This script is meant to be usable on GitHub Actions `windows-latest` 
#       runners using the bundled Bash (Git Bash) or when running under WSL.
#       It attempts to be tolerant of differences in available utilities and 
#       provides helpful error messages when dependencies are missing.

set -euo pipefail

usage() {
    cat <<'EOF'
    This script tests the built flows2fim executable's methods on Windows
    runners. It must be run from the root directory of the flows2fim
    repository to ensure proper pathing for test data.

    Usage: ./scripts/test_suite_windows.sh [OPTIONS] [METHOD]

    OPTIONS:
        -d, --dev      Run tests using 'go run main.go' instead of compiled binary
        -h, --help     Show this help message

    METHODS:
        controls: Only issue the controls tests.
        fim: Only issue the fim tests.
        validate: Only issue the validate tests.

    Providing no method, or 'all', will test all flows2fim methods.
EOF
}

dev_mode=false
cmd="flows2fim"
CMD_EXEC=""
method="all"

while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--dev)
            dev_mode=true
            cmd="go run main.go"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -* )
            echo "Invalid option: $1"
            usage
            exit 1
            ;;
        *)
            method="$1"
            shift
            ;;
    esac
done

# Ensure script is run from repo root
# Normalize paths to handle Windows Git Bash path format differences
repo_root=$(git rev-parse --show-toplevel 2>/dev/null)
current_dir=$(pwd)

# Convert Windows-style path (D:/...) to Unix-style (/d/...) if needed
if [[ "$repo_root" =~ ^[A-Z]:/.*$ ]]; then
    drive_letter=$(echo "$repo_root" | cut -c1 | tr '[:upper:]' '[:lower:]')
    repo_root="/${drive_letter}${repo_root:2}"
fi

if [[ "$repo_root" != "$current_dir" ]]; then
    echo "Error: This script must be run from the root directory of the flows2fim repository."
    usage
    exit 1
fi

# Check command availability
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

if $dev_mode; then
    echo "Running in dev mode using: $cmd"
    CMD_EXEC="$cmd"
else
    if command_exists "$cmd"; then
        CMD_EXEC="$cmd"
    elif command_exists "$cmd.exe"; then
        CMD_EXEC="$cmd.exe"
    else
        echo "flows2fim binary not found or not executable. Make sure the artifact is downloaded and available in PATH or run with -d for dev mode."
        exit 1
    fi
    # verify
    if ! $CMD_EXEC --version >/dev/null 2>&1; then
        echo "Unable to execute $CMD_EXEC --version; ensure the binary is runnable on this runner."
        exit 1
    fi
fi

export CMD_EXEC

data_folder=testdata
ref_dir=$data_folder/reference_data
db_path=$ref_dir/db
controls_benchmark_dir=$ref_dir/controls
fim_benchmark_dir=$ref_dir/fim
flows_files_dir=$ref_dir/flows_files
library_benchmark=$ref_dir/library
start_reaches_dir=$ref_dir/start_reaches
fim_reference_output_formats=$ref_dir/fim_output_formats

test_outputs=$data_folder/test_out
control_test_outputs=$test_outputs/controls
fim_test_outputs=$test_outputs/fim
fim_test_output_formats=$test_outputs/fim_output_formats

total_count=0
total_passed=0

# Environment/runtime checks and helpers
print_env_info() {
    echo "Runtime info:"
    echo "  uname: $(uname -a 2>/dev/null || true)"
    echo "  Bash: $BASH_VERSION"
    echo "  pwd: $(pwd)"
}

detect_dependencies() {
    print_env_info

    # GDAL core binary
    if command_exists gdalinfo; then
        echo "Found gdalinfo -> OK"
    else
        echo "Warning: 'gdalinfo' not found. Some tests may fail."
    fi

    # Python (used for gdalcompare.py and helpers)
    if command_exists python3; then
        PY=python3
    elif command_exists python; then
        PY=python
    else
        echo "Warning: Python not found. gdalcompare.py checks may be unavailable."
        PY=""
    fi

    # Find gdalcompare.py: try direct command, then python module locations
    GDALCOMPARE_CMD=""
    if command_exists gdalcompare.py; then
        GDALCOMPARE_CMD="gdalcompare.py"
    elif [[ -n "$PY" ]]; then
        # Try importing osgeo_utils.gdalcompare (GDAL >=3)
        if $PY - <<PYCODE 2>/dev/null
import importlib.util
import sys
spec = importlib.util.find_spec('osgeo_utils.gdalcompare')
print(bool(spec))
PYCODE
        then
            GDALCOMPARE_CMD="$PY -m osgeo_utils.gdalcompare"
        else
            # Try old script name
            if $PY - <<'PYCODE' 2>/dev/null
import importlib.util
import sys
spec = importlib.util.find_spec('gdal')
print(bool(spec))
PYCODE
            then
                # gdal Python bindings available; try calling gdalcompare via script path
                if $PY -c "import shutil,sys; p=shutil.which('gdalcompare.py'); print(p or '')" 2>/dev/null | grep -q -v '^$'; then
                    GDALCOMPARE_CMD="gdalcompare.py"
                fi
            fi
        fi
    fi

    if [[ -n "$GDALCOMPARE_CMD" ]]; then
        echo "Using gdalcompare: $GDALCOMPARE_CMD"
    else
        echo "gdalcompare not found; falling back to byte-wise comparisons where appropriate."
    fi

    export PY GDALCOMPARE_CMD
}

detect_dependencies

compare_directories() {
    local dir1=$1
    local dir2=$2
    local fim_mode=${3:-}

    local any_diff=""

    for file in "$dir1"/*; do
        filename=$(basename "$file")
        filepath2="$dir2/$filename"
        if [[ "$filename" == "fim_2year_test_rel_false.vrt" ]]; then
            continue
        fi
        if [[ -f "$filepath2" && "$fim_mode" == "fim" ]]; then
            tempfile=$(mktemp)
            if [[ -n "${GDALCOMPARE_CMD:-}" ]]; then
                # If GDALCOMPARE_CMD is a python -m invocation, it may contain spaces
                eval "$GDALCOMPARE_CMD \"$file\" \"$filepath2\"" &> "$tempfile" || true
                gdalcompare_output=$(tail -n 1 "$tempfile" | grep -Eo "[0-9]+" | tail -n 1 || true)
                rm -f "$tempfile"
                gdalcompare_difference_tolerance=2
                if [[ -n "$gdalcompare_output" && $gdalcompare_output -gt $gdalcompare_difference_tolerance ]]; then
                    echo "Files differ: $filename"
                    any_diff=1
                fi
            else
                # fallback to binary compare for non-raster/text files
                if ! cmp -s "$file" "$filepath2"; then
                    echo "Files differ (cmp): $filename"
                    any_diff=1
                fi
            fi
        elif [[ -f "$filepath2" ]]; then
            # For CSV files, normalize line endings before comparison
            if [[ "$filename" == *.csv ]]; then
                # Compare after normalizing line endings (handle Windows CRLF vs Unix LF)
                if ! diff -q --strip-trailing-cr "$file" "$filepath2" &> /dev/null; then
                    echo "Files differ: $filename"
                    # Show first few differences for debugging
                    echo "  First 5 lines of differences:"
                    diff --strip-trailing-cr "$file" "$filepath2" | head -n 10 || true
                    any_diff=1
                fi
            else
                if ! cmp -s "$file" "$filepath2"; then
                    echo "Files differ: $filename"
                    any_diff=1
                fi
            fi
        else
            echo "File not found in dir2: $filename"
            any_diff=1
        fi
    done

    for file in "$dir2"/*; do
        filename=$(basename "$file")
        filepath1="$dir1/$filename"
        if [[ ! -f "$filepath1" ]]; then
            echo "File in $dir2 not found in $dir1: $filename"
            any_diff=1
        fi
    done

    if [[ -n "$any_diff" ]]; then
        return 1
    fi
    return 0
}

controls_test_cases() {
    # Test 1: Generate and compare 6 controls files (2, 5, 10, 25, 50, 100 year)
    # Test 2: Empty start reaches should produce header-only output
    # Test 3: Empty flows file should throw "Flow not found" error
    # Test 4: Swapped columns in flows file should throw "Flow not found" error
    # Test 5: Empty flow values should throw "Flow not found" error
    local num_test_cases_controls=5
    local failed_controls_testcases=0
    total_count=$(( total_count + num_test_cases_controls ))

    if [[ -d "$control_test_outputs" ]]; then
        rm -rf "$control_test_outputs"
    fi
    mkdir -p "$control_test_outputs"

    echo "Running controls test cases..."
    echo "Test 1: Generating controls files for 6 recurrence intervals..."

    local recurrence_interval=(2 5 10 25 50 100)
    for interval in "${recurrence_interval[@]}"; do
        $CMD_EXEC controls -db "$db_path/ripple.gpkg" \
            -f "$flows_files_dir/flows_${interval}year.csv" \
            -o "$control_test_outputs/controls_${interval}year.csv" \
            -scsv "$start_reaches_dir/start_reaches.csv"
    done

    if compare_directories "$control_test_outputs" "$controls_benchmark_dir"; then
        echo "Test 1 PASSED: No difference in controls files."
    else
        echo "Test 1 FAILED: Outputs differ for controls."
        failed_controls_testcases=$((failed_controls_testcases + 1))
    fi

    # Test 2: Empty start reaches should produce header-only output
    echo "Test 2: Checking empty start reaches behavior..."
    $CMD_EXEC controls -db "$db_path/ripple.gpkg" -f "$flows_files_dir/flows_2year.csv" -o "$control_test_outputs/controls_2year_empty.csv" -scsv "$start_reaches_dir/empty_start_reaches.csv" &> /dev/null || true
    # Check header only (empty outputs) - the output file should have just the header
    if [[ -f "$control_test_outputs/controls_2year_empty.csv" ]] && head -n 1 "$control_test_outputs/controls_2year_empty.csv" | grep -q "reach_id,flow,control_stage"; then
        # Check that file has only 1 line (header only, no data)
        line_count=$(wc -l < "$control_test_outputs/controls_2year_empty.csv" | tr -d '[:space:]')
        if [[ "$line_count" -eq 1 ]]; then
            echo "Test 2 PASSED: Empty start reaches produced header-only file"
        else
            echo "Test 2 FAILED: Empty start reaches file has $line_count lines (expected 1)"
            failed_controls_testcases=$((failed_controls_testcases + 1))
        fi
    else
        echo "Test 2 FAILED: Empty start reaches did not produce expected output file"
        failed_controls_testcases=$((failed_controls_testcases + 1))
    fi

    # Test 3: Empty flows file should throw "Flow not found" error
    echo "Test 3: Checking empty flows file error handling..."
    tempfile=$(mktemp)
    temp_out=$(mktemp)
    $CMD_EXEC controls -db "$db_path/ripple.gpkg" \
        -f "$flows_files_dir/empty_file.csv" \
        -o "$temp_out" \
        -scsv "$start_reaches_dir/start_reaches.csv" &> "$tempfile" || true
    if grep -q "Flow not found for reach" "$tempfile"; then
        echo "Test 3 PASSED: Flow not found error thrown for empty file"
    else
        echo "Test 3 FAILED: Flow not found error not thrown for empty file"
        cat "$tempfile"
        failed_controls_testcases=$((failed_controls_testcases + 1))
    fi
    rm -f "$temp_out" "$tempfile"

    # Test 4: Swapped columns in flows file should throw "Flow not found" error
    echo "Test 4: Checking swapped columns error handling..."
    tempfile=$(mktemp)
    temp_out=$(mktemp)
    $CMD_EXEC controls -db "$db_path/ripple.gpkg" \
        -f "$flows_files_dir/flows_2year_swapped.csv" \
        -o "$temp_out" \
        -scsv "$start_reaches_dir/start_reaches.csv" &> "$tempfile" || true
    if grep -q "Flow not found for reach" "$tempfile"; then
        echo "Test 4 PASSED: Flow not found error thrown for swapped columns"
    else
        echo "Test 4 FAILED: Flow not found error not thrown for swapped columns"
        cat "$tempfile"
        failed_controls_testcases=$((failed_controls_testcases + 1))
    fi
    rm -f "$temp_out" "$tempfile"

    # Test 5: Empty flow values should throw "Flow not found" error
    echo "Test 5: Checking empty flow values error handling..."
    tempfile=$(mktemp)
    temp_out=$(mktemp)
    $CMD_EXEC controls -db "$db_path/ripple.gpkg" \
        -f "$flows_files_dir/flows_2year_empty_values.csv" \
        -o "$temp_out" \
        -scsv "$start_reaches_dir/start_reaches.csv" &> "$tempfile" || true
    if grep -q "Flow not found for reach" "$tempfile"; then
        echo "Test 5 PASSED: Flow not found error thrown for empty flow values"
    else
        echo "Test 5 FAILED: Flow not found error not thrown for empty flow values"
        cat "$tempfile"
        failed_controls_testcases=$((failed_controls_testcases + 1))
    fi
    rm -f "$temp_out" "$tempfile"

    controls_passed=$(( num_test_cases_controls - failed_controls_testcases ))
    total_passed=$(( total_passed + controls_passed ))
}

fim_test_cases() {
    # Test 1: Generate fim files for 6 recurrence intervals (2, 5, 10, 25, 50, 100 year)
    # Test 2: Regression test comparing generated files
    # Test 3: Generate fim files in different output formats (gtiff, cog, vrt)
    # Test 4: Regression test for different output formats
    # Test 5: Assert error thrown from missing controls file parameter
    # Test 6: Assert error thrown from missing library parameter
    # Test 7: Assert error thrown from missing output file parameter
    # Test 8: Assert error thrown from empty controls file
    local num_test_cases_fim=8
    local failed_fim_testcases=0
    total_count=$(( total_count + num_test_cases_fim ))

    if [[ -d "$fim_test_outputs" ]]; then
        rm -rf "$fim_test_outputs"
        rm -rf "$fim_test_output_formats"
    fi
    mkdir -p "$fim_test_outputs"
    mkdir -p "$fim_test_output_formats"

    echo "Running fim test cases..."

    # Test 1: Generate fim files for 6 recurrence intervals
    echo "Test 1: Generating fim_year.tif files for 6 recurrence intervals..."
    local recurrence_interval=(2 5 10 25 50 100)
    local fim_file_format="GTiff"
    for interval in "${recurrence_interval[@]}"; do
        $CMD_EXEC fim -c "$controls_benchmark_dir/controls_${interval}year.csv" -fmt "$fim_file_format" -lib "$library_benchmark" -type depth -o "$fim_test_outputs/fim_${interval}year.tif" &> /dev/null
    done

    # Test 2: Regression test for all recurrence interval fim.tif files
    echo "Test 2: Regression tests for all recurrence interval fim.tif files..."
    if compare_directories "$fim_test_outputs" "$fim_benchmark_dir" "fim"; then
        echo "Test 2 PASSED: No significant difference in fim.tif files."
    else
        echo "Test 2 FAILED: Outputs differ for fim files."
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi

    # Test 3: Generate fim files in different output formats
    echo "Test 3: Generating fim files in different output formats..."
    local file_formats=(gtiff cog vrt)
    for format in "${file_formats[@]}"; do
        local output_file="fim_2year.$format"
        $CMD_EXEC fim -c "$controls_benchmark_dir/controls_2year.csv" -fmt "$format" -lib "$library_benchmark" -type depth -o "$fim_test_output_formats/$output_file" &> /dev/null
    done

    # Test 4: Regression test for different output formats
    echo "Test 4: Regression tests for different output formats..."
    if compare_directories "$fim_reference_output_formats" "$fim_test_output_formats" "fim"; then
        echo "Test 4 PASSED: No differences in .cog, .vrt & .tif files."
    else
        echo "Test 4 FAILED: Outputs differ for different formats."
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi

    # Test 5: Assert error thrown from missing controls file parameter
    echo "Test 5: Checking missing controls file parameter error..."
    tempfile=$(mktemp)
    $CMD_EXEC fim -fmt GTiff -lib "$library_benchmark" -type depth -o "$fim_test_outputs/fim_test.tif" &> "$tempfile" || true
    if grep -q "missing required flags" "$tempfile"; then
        echo "Test 5 PASSED: Correct error thrown for missing controls parameter"
    else
        echo "Test 5 FAILED: Error messaging inconsistent"
        cat "$tempfile"
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi
    rm -f "$tempfile"

    # Test 6: Assert error thrown from missing library parameter
    echo "Test 6: Checking missing library parameter error..."
    tempfile=$(mktemp)
    $CMD_EXEC fim -c "$controls_benchmark_dir/controls_2year.csv" -fmt GTiff -type depth -o "$fim_test_outputs/fim_test.tif" &> "$tempfile" || true
    if grep -q "missing required flags" "$tempfile"; then
        echo "Test 6 PASSED: Correct error thrown for missing library parameter"
    else
        echo "Test 6 FAILED: Error messaging inconsistent"
        cat "$tempfile"
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi
    rm -f "$tempfile"

    # Test 7: Assert error thrown from missing output file parameter
    echo "Test 7: Checking missing output file parameter error..."
    tempfile=$(mktemp)
    $CMD_EXEC fim -c "$controls_benchmark_dir/controls_2year.csv" -fmt GTiff -type depth -lib "$library_benchmark" &> "$tempfile" || true
    if grep -q "missing required flags" "$tempfile"; then
        echo "Test 7 PASSED: Correct error thrown for missing output parameter"
    else
        echo "Test 7 FAILED: Error messaging inconsistent"
        cat "$tempfile"
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi
    rm -f "$tempfile"

    # Test 8: Assert error thrown from empty controls file
    echo "Test 8: Checking empty controls file error..."
    tempfile=$(mktemp)
    $CMD_EXEC fim -c "$flows_files_dir/empty_file.csv" -fmt GTiff -lib "$library_benchmark" -type depth -o "$fim_test_outputs/fim_test.tif" &> "$tempfile" || true
    if grep -q "no records in controls file" "$tempfile"; then
        echo "Test 8 PASSED: Correct error thrown for empty controls file"
    else
        echo "Test 8 FAILED: Error messaging inconsistent"
        cat "$tempfile"
        failed_fim_testcases=$((failed_fim_testcases + 1))
    fi
    rm -f "$tempfile"

    fim_passed=$(( num_test_cases_fim - failed_fim_testcases ))
    total_passed=$(( total_passed + fim_passed ))
}

validate_test_cases() {
    echo "Validate test cases are not fully implemented for Windows."
    echo "Note: On Windows, validate functionality may be limited without GDAL runtime libraries."
    echo "Skipping validate tests on this platform."
}

case "$method" in
    controls)
        controls_test_cases
        ;;
    fim)
        fim_test_cases
        ;;
    validate)
        validate_test_cases
        ;;
    all)
        test_help_statement() { echo; }
        controls_test_cases
        fim_test_cases
        ;;
    *)
        echo "Unknown method: $method"
        usage
        exit 1
        ;;
esac

echo "Total tests run: $total_count"
echo "Total passed (approx): $total_passed"

if [[ $total_passed -lt $total_count ]]; then
    echo "Some tests failed or comparisons flagged differences. Inspect output under $test_outputs"
    exit 1
fi

exit 0
