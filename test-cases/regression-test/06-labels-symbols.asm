// Regression Test 06: Labels and Symbols
// Purpose: Ensure label definition and usage work
// Status: Should PASS with 0 errors

* = $0801

// Basic labels
start:
    jmp main

// Label with assignment
.label data_ptr = $fb

// Subroutine labels
init:
    lda #$00
    sta $d020
    rts

clear_screen:
    lda #$20
    ldx #$00
loop:
    sta $0400, x
    sta $0500, x
    sta $0600, x
    sta $0700, x
    inx
    bne loop
    rts

// Local labels (if supported)
main:
    jsr init
    jsr clear_screen

    // Forward reference
    jmp end

data:
    .byte $00, $01, $02

end:
    rts

// Label usage in expressions
.const MAIN_ADDR = main
.const INIT_SIZE = clear_screen - init
