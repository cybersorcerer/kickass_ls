// Test macro call argument count validation

.macro setColor(color) {
    lda #color
    sta $d020
}

.macro addValues(dest, val1, val2) {
    clc
    lda val1
    adc val2
    sta dest
}

* = $0801

start:
    setColor(1)              // ✅ Correct: 1 argument
    setColor(1, 2)           // ❌ Should error: too many arguments (2 instead of 1)
    addValues($80, #$10, #$20)  // ✅ Correct: 3 arguments
    addValues($80)           // ❌ Should error: too few arguments (1 instead of 3)
    rts
