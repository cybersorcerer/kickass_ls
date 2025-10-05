// Test Case 03: .function Directive
// Purpose: Test parameter list parsing for .function
// Issue #3: Directive Parameter Parsing

* = $0801

// Function with single parameter
.function square(x) {
    .return x * x
}

// Function with multiple parameters
.function add(a, b) {
    .return a + b
}

// Function with no parameters
.function getScreenStart() {
    .return $0400
}

// Function with complex expression
.function clamp(value, min, max) {
    .return max(min(value, max), min)
}

// Use functions in code
.const RESULT = square(5)
.const SUM = add(10, 20)
.const SCREEN = getScreenStart()

start:
    lda #RESULT
    sta $d020
    rts

// Function without .return statement (should warn)
.function noReturn(x) {
    .var temp = x * 2
}

// Unused parameter (should hint)
.function unusedParam(x, y) {
    .return x * x
}
