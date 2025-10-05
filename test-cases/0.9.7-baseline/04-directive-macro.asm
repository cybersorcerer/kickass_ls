// Test Case 04: .macro Directive
// Purpose: Test parameter list parsing for .macro
// Issue #3: Directive Parameter Parsing

* = $0801

// Macro with no parameters
.macro nopDelay() {
    nop
    nop
    nop
}

// Macro with single parameter
.macro storeValue(value) {
    lda #value
    sta $d020
}

// Macro with multiple parameters
.macro addValues(dest, val1, val2) {
    clc
    lda val1
    adc val2
    sta dest
}

// Macro with register parameters
.macro saveRegs(addr) {
    sta addr
    stx addr+1
    sty addr+2
}

start:
    // Use macros
    nopDelay()
    storeValue($0e)
    addValues($80, #$10, #$20)
    saveRegs($fb)
    rts

// Macro with unused parameter (should hint)
.macro unusedMacro(a, b, c) {
    lda #a
    sta $d020
}

// Macro parameter count mismatch (should error when called)
call_test:
    storeValue($01, $02)  // Too many arguments
