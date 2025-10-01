#!/bin/bash
# comprehensive-regression-test.sh - Run after each kickass-plan.md phase

set -e

echo "🧪 Comprehensive Regression Test Suite"
echo "======================================="

TEST_DIR="test-cases"
SERVER_PATH="/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server"
RESULTS_DIR="regression-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create results directory
mkdir -p "$RESULTS_DIR/$TIMESTAMP"

# Test categories - All 19 comprehensive tests
TESTS=(
    "lsp-lifecycle-test.json"
    "configuration-loading-test.json"
    "directive-processing-test.json"
    "symbol-table-test.json"
    "context-aware-completion-test.json"
    "performance-baseline-test.json"
    "memory-leak-test.json"
    "memory-completion-test.json"
    "debug-completion-test.json"
    "basic-completion.json"
    "lda-hash-test.json"
    "lexer-parser-integration-tests.json"
    "inc-fix-test.json"
    "arithmetic-expression-test.json"
    "builtin-function-validation-test.json"
    "advanced-for-directive-test.json"
    "for-scope-management-test.json"
    "conditional-directive-test.json"
    "type-system-integration-test.json"
    "comprehensive-regression-test.json"
)

# Results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_TEST_NAMES=()

echo "Starting regression test execution..."
echo ""

for test in "${TESTS[@]}"; do
    echo "Running: $test"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if ./test-client/test-client -suite "$TEST_DIR/$test" -server "$SERVER_PATH" \
       -output "$RESULTS_DIR/$TIMESTAMP/${test%.json}_result.json" > /dev/null 2>&1; then
        echo "✅ PASS: $test"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "❌ FAIL: $test"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test")
    fi
done

echo ""
echo "📊 Regression Test Summary"
echo "========================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Success Rate: $((PASSED_TESTS * 100 / TOTAL_TESTS))%"

# Show failed tests if any
if [ $FAILED_TESTS -gt 0 ]; then
    echo ""
    echo "❌ Failed Tests:"
    for failed_test in "${FAILED_TEST_NAMES[@]}"; do
        echo "  - $failed_test"
    done
fi

# Performance baseline comparison (if baseline exists)
if [ -f "baseline-performance.json" ]; then
    echo ""
    echo "📈 Performance Comparison"
    echo "========================"
    echo "Comparing current results with baseline..."

    # Extract performance data from performance-baseline-test results
    if [ -f "$RESULTS_DIR/$TIMESTAMP/performance-baseline-test_result.json" ]; then
        echo "Performance test results saved for comparison."
        echo "📋 Check: $RESULTS_DIR/$TIMESTAMP/performance-baseline-test_result.json"
    else
        echo "⚠️  Performance test results not found."
    fi
else
    echo ""
    echo "📋 Creating Performance Baseline"
    echo "==============================="
    if [ -f "$RESULTS_DIR/$TIMESTAMP/performance-baseline-test_result.json" ]; then
        cp "$RESULTS_DIR/$TIMESTAMP/performance-baseline-test_result.json" "baseline-performance.json"
        echo "✅ Performance baseline created from this run."
    fi
fi

# Critical functionality validation
echo ""
echo "🔍 Critical Functionality Validation"
echo "===================================="

CRITICAL_TESTS=(
    "memory-completion-test.json"
    "directive-processing-test.json"
    "symbol-table-test.json"
    "lsp-lifecycle-test.json"
    "comprehensive-regression-test.json"
)

CRITICAL_PASSED=0
for critical_test in "${CRITICAL_TESTS[@]}"; do
    if [[ " ${FAILED_TEST_NAMES[@]} " =~ " ${critical_test} " ]]; then
        echo "❌ CRITICAL FAILURE: $critical_test"
    else
        echo "✅ CRITICAL PASS: $critical_test"
        CRITICAL_PASSED=$((CRITICAL_PASSED + 1))
    fi
done

echo ""
echo "Critical Tests: ${#CRITICAL_TESTS[@]}"
echo "Critical Passed: $CRITICAL_PASSED"

# Generate detailed report
echo ""
echo "📋 Detailed Report"
echo "=================="
echo "Results saved to: $RESULTS_DIR/$TIMESTAMP/"
echo "Individual test results:"
for test in "${TESTS[@]}"; do
    result_file="$RESULTS_DIR/$TIMESTAMP/${test%.json}_result.json"
    if [ -f "$result_file" ]; then
        echo "  📄 ${test%.json}_result.json"
    fi
done

# Phase gate decision
echo ""
echo "🚦 Phase Gate Decision"
echo "====================="

if [ $FAILED_TESTS -eq 0 ]; then
    if [ $CRITICAL_PASSED -eq ${#CRITICAL_TESTS[@]} ]; then
        echo "🎉 ALL TESTS PASSED! ✅"
        echo "✅ All critical functionality validated"
        echo "✅ Safe to proceed with kickass-plan.md implementation"
        echo ""
        echo "📋 Next Steps:"
        echo "1. Begin kickass-plan.md Phase 1 implementation"
        echo "2. Run this script after each major change"
        echo "3. Never proceed to next phase without 100% pass rate"
        exit 0
    else
        echo "⚠️  CRITICAL TESTS FAILED - IMPLEMENTATION BLOCKED"
        echo "❌ $((${#CRITICAL_TESTS[@]} - CRITICAL_PASSED)) critical test(s) failed"
        echo "🚫 DO NOT PROCEED with kickass-plan.md implementation"
        exit 1
    fi
else
    echo "⚠️  $FAILED_TESTS TEST(S) FAILED - REVIEW REQUIRED"
    echo "❌ Failed tests: ${FAILED_TEST_NAMES[*]}"
    echo "🚫 DO NOT PROCEED with kickass-plan.md implementation"
    echo ""
    echo "📋 Required Actions:"
    echo "1. Fix all failing tests"
    echo "2. Re-run this script"
    echo "3. Only proceed when success rate = 100%"
    exit 1
fi