// Test Case 05: .namespace Directive
// Purpose: Test namespace scope management
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic namespace
.namespace graphics {
    .const SCREEN = $0400
    .const COLOR_RAM = $d800

    clear:
        lda #$20
        ldx #$00
    loop:
        sta SCREEN, x
        inx
        bne loop
        rts
}

// Another namespace
.namespace sound {
    .const SID_BASE = $d400

    init:
        lda #$00
        sta SID_BASE
        rts
}

// Nested namespace
.namespace utils {
    .namespace math {
        .function abs(x) {
            .return (x < 0) ? -x : x
        }
    }

    delay:
        ldx #$ff
    wait:
        dex
        bne wait
        rts
}

start:
    // Use namespaced symbols
    jsr graphics.clear
    jsr sound.init
    jsr utils.delay

    lda #graphics.SCREEN
    sta $fb
    rts

// Duplicate namespace (should warn)
.namespace graphics {
    .const DUPLICATE = $1000
}
