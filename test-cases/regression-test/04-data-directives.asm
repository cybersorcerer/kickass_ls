// Regression Test 04: Data Directives
// Purpose: Ensure .byte, .word, .text, .fill work correctly
// Status: Should PASS with 0 errors

* = $0801

start:
    jmp main

// .byte directive - single and multiple values
data_bytes:
    .byte $00
    .byte $01, $02, $03
    .byte $ff, $fe, $fd, $fc

// .word directive - 16-bit values
data_words:
    .word $0000
    .word $1234, $5678
    .word $ffff, $abcd

// .text directive - strings
text_data:
    .text "HELLO WORLD"
    .text "C64 RULES!"

// .fill directive - fill memory
fill_data:
    .fill 10, $00
    .fill 5, $ff
    .fill 256, $20

// .align directive
    .align $100
aligned_data:
    .byte $00

main:
    lda #$00
    sta $d020
    rts
