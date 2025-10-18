#!/bin/bash

# Regression Test Runner for kickass_ls
# Runs all test suites in test-cases/ to ensure no functionality is lost

set -e

echo "========================================"
echo " kickass_ls Regression Test Suite"
echo "========================================"
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

# Check if server exists
if [ ! -f "build/kickass_ls" ]; then
    echo -e "${RED}ERROR: kickass_ls server not found.${RESET}"
    echo "Build it first with: make build"
    exit 1
fi

# Check if test client exists
if [ ! -f "build/kickass_cl" ]; then
    echo -e "${RED}ERROR: kickass_cl test client not found.${RESET}"
    echo "Build it first with: make build"
    exit 1
fi

# Find all test suite JSON files
TEST_SUITES=(
    "test-cases/regression-test/regression-suite.json"
    "test-cases/0.9.0-baseline/baseline-suite.json"
    "test-cases/0.9.7-baseline/baseline-suite.json"
    "test-cases/test-files/test-encoding-suite.json"
)

# Counters
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0
TIMEOUT_SUITES=0

# Test results
FAILED_SUITE_NAMES=()

# Run each test suite
for suite in "${TEST_SUITES[@]}"; do
    if [ ! -f "$suite" ]; then
        echo -e "${YELLOW}⚠ Skipping: $suite (not found)${RESET}"
        continue
    fi

    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    suite_name=$(basename "$suite" .json)
    suite_dir=$(dirname "$suite")

    echo -e "${CYAN}Running: $suite_dir/$suite_name${RESET}"

    # Run test with timeout
    if timeout 60s build/kickass_cl -suite "$suite" -server "build/kickass_ls" > "test_output_${suite_name}.log" 2>&1; then
        echo -e "${GREEN}✓ PASSED${RESET}"
        PASSED_SUITES=$((PASSED_SUITES + 1))
    else
        exit_code=$?
        if [ $exit_code -eq 124 ]; then
            echo -e "${RED}✗ TIMEOUT (60s exceeded)${RESET}"
            TIMEOUT_SUITES=$((TIMEOUT_SUITES + 1))
            FAILED_SUITE_NAMES+=("$suite_name (timeout)")
        else
            echo -e "${RED}✗ FAILED (exit code: $exit_code)${RESET}"
            echo -e "${YELLOW}  See: test_output_${suite_name}.log${RESET}"
            FAILED_SUITES=$((FAILED_SUITES + 1))
            FAILED_SUITE_NAMES+=("$suite_name (failed)")

            # Show last 10 lines of error
            echo -e "${YELLOW}  Last 10 lines:${RESET}"
            tail -10 "test_output_${suite_name}.log" | sed 's/^/    /'
        fi
    fi
    echo
done

# Summary
echo "========================================"
echo " Test Summary"
echo "========================================"
echo -e "Total test suites:  ${CYAN}${TOTAL_SUITES}${RESET}"
echo -e "Passed:             ${GREEN}${PASSED_SUITES}${RESET}"
echo -e "Failed:             ${RED}${FAILED_SUITES}${RESET}"
echo -e "Timeout:            ${YELLOW}${TIMEOUT_SUITES}${RESET}"
echo

if [ ${#FAILED_SUITE_NAMES[@]} -gt 0 ]; then
    echo -e "${RED}Failed suites:${RESET}"
    for failed in "${FAILED_SUITE_NAMES[@]}"; do
        echo -e "  ${RED}✗${RESET} $failed"
    done
    echo
fi

# Exit code
if [ $FAILED_SUITES -gt 0 ] || [ $TIMEOUT_SUITES -gt 0 ]; then
    echo -e "${RED}Regression tests FAILED${RESET}"
    exit 1
else
    echo -e "${GREEN}All regression tests PASSED${RESET}"
    exit 0
fi
