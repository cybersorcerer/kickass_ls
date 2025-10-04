#!/bin/bash

# Regression Test Runner for kickass_ls
# This script runs comprehensive tests to ensure no functionality is lost during redesign

echo "=== kickass_ls Regression Test Suite ==="
echo "Testing kickass_ls server functionality before context-aware redesign"
echo

# Check if server exists
if [ ! -f "./kickass_ls" ]; then
    echo "ERROR: kickass_ls server not found. Build it first with: go build -o kickass_ls ."
    exit 1
fi

# Check if test client exists
if [ ! -f "./kickass_cl/kickass_cl" ]; then
    echo "ERROR: kickass_cl test client not found. Build it first with: cd kickass_cl && go build -o kickass_cl ."
    exit 1
fi

# Check if comprehensive test file exists
if [ ! -f "./comprehensive-test.asm" ]; then
    echo "ERROR: comprehensive-test.asm not found"
    exit 1
fi

# Create required test files
echo "Creating test files..."

cat > test-cases/test_simple.asm << 'EOF'
// Simple test file for completion and hover tests
*= $0801
start:
    lda #$00
    sta $d020
    rts
EOF

cat > test-cases/test_builtins.asm << 'EOF'
// Test file for built-in function completion
.const angle = 45
.const radius = sin(angle) * 100
.const file_size = size("data.bin")
EOF

cat > test-cases/test_signature.asm << 'EOF'
// Test file for signature help
*= $0801
.function calculateDistance(x1, y1, x2, y2) {
    .return sqrt(pow(x2-x1, 2) + pow(y2-y1, 2))
}
EOF

cat > test-cases/test_comprehensive.asm << 'EOF'
// Large test file for performance testing
.const NUM_SPRITES = 8
*= $0801
.for (var i = 0; i < NUM_SPRITES; i++) {
    .for (var j = 0; j < 21; j++) {
        .byte $00
    }
}
EOF

cat > test-cases/test_illegal.asm << 'EOF'
// Test file for illegal opcode detection
*= $0801
    lda #$00
    dcp $ff      // Illegal opcode
    sax $80      // Illegal opcode
    rra $c000    // Illegal opcode
EOF

echo "Test files created."
echo

# Function to run a test and capture results
run_test() {
    local test_name="$1"
    local test_file="$2"
    local expected_patterns="$3"

    echo "Running: $test_name"

    # Run the test with timeout
    timeout 30s ./kickass_cl/kickass_cl -suite "$test_file" -server "./kickass_ls" > "test_output_${test_name// /_}.log" 2>&1
    local exit_code=$?

    if [ $exit_code -eq 124 ]; then
        echo "  ❌ TIMEOUT (30s exceeded)"
        return 1
    elif [ $exit_code -ne 0 ]; then
        echo "  ❌ FAILED (exit code: $exit_code)"
        return 1
    fi

    # Check for expected patterns
    if [ -n "$expected_patterns" ]; then
        IFS='|' read -ra PATTERNS <<< "$expected_patterns"
        for pattern in "${PATTERNS[@]}"; do
            if ! grep -q "$pattern" "test_output_${test_name// /_}.log"; then
                echo "  ❌ FAILED (missing pattern: $pattern)"
                return 1
            fi
        done
    fi

    echo "  ✅ PASSED"
    return 0
}

# Run individual tests
echo "=== Running Individual Feature Tests ==="
echo

passed=0
failed=0

# Test 1: Basic LSP Communication
if run_test "Basic LSP" "test-cases/test-kickass-ls-basic.json" "kickass_ls"; then
    ((passed++))
else
    ((failed++))
fi

# Test 2: Dead Code Detection
if run_test "Dead Code Detection" "test-cases/test-dead-code.json" "Dead code"; then
    ((passed++))
else
    ((failed++))
fi

# Test 3: Range Validation
if run_test "Range Validation" "test-cases/test-range-validation.json" "out of.*range"; then
    ((passed++))
else
    ((failed++))
fi

# Test 4: Zero Page Hints
if run_test "Zero Page Hints" "test-cases/test-zero-page-hints.json" "zero page"; then
    ((passed++))
else
    ((failed++))
fi

# Test 5: For Loop Processing
if run_test "For Loop Processing" "test-cases/test-comprehensive-for-loop.json" ""; then
    ((passed++))
else
    ((failed++))
fi

# Test 6: Illegal Opcodes
if run_test "Illegal Opcodes" "test-cases/test-simple-illegal.json" "illegal"; then
    ((passed++))
else
    ((failed++))
fi

echo
echo "=== Running Comprehensive Regression Test ==="
echo

# Run the full regression test
timeout 60s ./kickass_cl/kickass_cl -suite "test-cases/kickass-ls-full-regression-test.json" -server "./kickass_ls" > regression_test_output.log 2>&1
regression_exit_code=$?

if [ $regression_exit_code -eq 124 ]; then
    echo "❌ Comprehensive regression test TIMEOUT (60s exceeded)"
    ((failed++))
elif [ $regression_exit_code -eq 0 ]; then
    echo "✅ Comprehensive regression test PASSED"
    ((passed++))
else
    echo "❌ Comprehensive regression test FAILED (exit code: $regression_exit_code)"
    ((failed++))
fi

echo
echo "=== Performance Baseline Test ==="
echo

# Performance test
start_time=$(date +%s%N)
timeout 30s ./kickass_cl/kickass_cl -suite "test-cases/performance-baseline-test.json" -server "./kickass_ls" > performance_test_output.log 2>&1
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds

if [ $? -eq 0 ]; then
    echo "✅ Performance test PASSED (${duration}ms)"
    ((passed++))
else
    echo "❌ Performance test FAILED"
    ((failed++))
fi

echo
echo "=== Test Summary ==="
echo "Passed: $passed"
echo "Failed: $failed"
echo "Total:  $((passed + failed))"
echo

if [ $failed -eq 0 ]; then
    echo "🎉 ALL TESTS PASSED - kickass_ls is ready for redesign!"
    exit 0
else
    echo "❌ Some tests failed - fix issues before proceeding with redesign"
    echo
    echo "Check log files for details:"
    ls -la test_output_*.log regression_test_output.log performance_test_output.log 2>/dev/null
    exit 1
fi