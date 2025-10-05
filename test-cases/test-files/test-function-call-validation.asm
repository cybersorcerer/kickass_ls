// Test function call argument count validation

.function add(a, b) {
    .return a + b
}

.function square(x) {
    .return x * x
}

.function noParams() {
    .return 42
}

* = $0801

// Test function calls with wrong argument counts
.const RESULT1 = add(5, 10)        // ✅ Correct: 2 arguments
.const RESULT2 = add(5)            // ❌ Should error: too few arguments (1 instead of 2)
.const RESULT3 = add(5, 10, 15)    // ❌ Should error: too many arguments (3 instead of 2)
.const RESULT4 = square(5)         // ✅ Correct: 1 argument
.const RESULT5 = square()          // ❌ Should error: too few arguments (0 instead of 1)
.const RESULT6 = noParams()        // ✅ Correct: 0 arguments
.const RESULT7 = noParams(5)       // ❌ Should error: too many arguments (1 instead of 0)

start:
    lda #$00
    rts
