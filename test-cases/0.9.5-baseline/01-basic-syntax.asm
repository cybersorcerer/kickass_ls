// Test Case 01: Basic Syntax
// Purpose: Test fundamental assembler syntax elements

// Labels
start:
    nop

loop:
    lda #$00
    sta $d020
    jmp loop

end:
    rts

// Comments - testing different comment styles
// This is a valid comment

// Hex values
hex_test:
    lda #$ff
    sta $d021

// Binary values
binary_test:
    lda #%11111111
    sta $0400

// Decimal values
decimal_test:
    ldx #255
    ldy #128

// Character literals
char_test:
    lda #65     // ASCII 'A'
    sta $0400
