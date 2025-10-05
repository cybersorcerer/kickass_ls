// Test Case 02: .define Directive
// Purpose: Test symbol-only directive parsing (no value)
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic define directive (symbol only, no value)
.define DEBUG
.define RELEASE_MODE
.define ENABLE_SOUND

// Conditional compilation based on defines
.ifdef DEBUG
    nop
    nop
.endif

.ifndef RELEASE_MODE
    lda #$ff
    sta $d020
.endif

// Test .undef
.undef DEBUG

// Redefinition (should warn)
.define DEBUG

start:
    lda #$00
    rts
