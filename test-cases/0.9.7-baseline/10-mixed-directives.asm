// Test Case 10: Mixed Directive Tests
// Purpose: Test combinations of new directive features
// Comprehensive test for v0.9.7

* = $0801

.encoding "petscii_upper"
.define DEBUG_MODE

// Namespace with functions and macros
.namespace lib {
    .function calculate(x, y) {
        .return x * y + 10
    }

    .macro debug(msg) {
        .ifdef DEBUG_MODE
            // Debug output would go here
            nop
        .endif
    }

    .enum Status {
        OK = 0,
        ERROR = 1,
        PENDING = 2
    }
}

// Pseudocommand using namespace
.pseudocommand setStatus value : addr {
    lda #lib.Status.value
    sta addr
}

// Function with enum return
.function getDefaultColor() {
    .enum DefaultColors {
        BACKGROUND = 0,
        BORDER = 6
    }
    .return DefaultColors.BORDER
}

start:
    // PC expressions
    .label start_addr = *

    // Use all features together
    lib.debug("Starting")

    lda #lib.calculate(5, 3)
    sta $d020

    setStatus OK : $80

    lda #getDefaultColor()
    sta $d021

    // Relative branch with PC
    beq * + 8
    nop
    nop

    rts

.const PROGRAM_SIZE = * - start_addr
