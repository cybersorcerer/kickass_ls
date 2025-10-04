// Test Case 02: Kick Assembler Directives
// Purpose: Test all major Kick Assembler directives

// Variable declarations
.var counter = 0
.var address = $d020
.const SCREEN = $0400
.const BORDER = $d020

// Program counter
* = $0801

// Basic program header
.byte $0c, $08, $0a, $00, $9e
.text "2061"
.byte $00, $00, $00

// Main code
* = $0810
start:
    lda #$00
    sta BORDER

// Label directive
.label loop = *
    inc $d020
    jmp loop

// Conditional assembly
.if (counter == 0) {
    nop
    nop
}

// Define directive
.define DEBUG

// Macro directive
.macro clearScreen() {
    lda #$20
    ldx #$00
loop:
    sta SCREEN, x
    inx
    bne loop
}

// Namespace
.namespace utils {
    delay:
        ldx #$ff
    wait:
        dex
        bne wait
        rts
}

// Pseudocommand
.pseudocommand add arg1 : arg2 {
    clc
    lda arg1
    adc arg2
}

// Enum
.enum Colors {
    BLACK = 0,
    WHITE = 1,
    RED = 2
}

// Encoding
.encoding "screencode_mixed"

// Text with encoding
.text "HELLO WORLD"

// Fill directive
.fill 10, $00

// Import (commented - no file)
// .import source "lib.asm"

// Align directive
.align $100
aligned_data:
    .byte $00

// Function directive
.function getColor(x) {
    .return Colors.RED
}
