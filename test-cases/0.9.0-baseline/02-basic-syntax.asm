// ============================================================================
// 02-basic-syntax.asm - Basic Syntax Parsing Test
// ============================================================================
// Test: Parser correctly handles all basic syntax elements
// Expected: No parsing errors, clean AST generation
// ============================================================================

// ----------------------------------------------------------------------------
// 1. COMMENTS - All comment styles
// ----------------------------------------------------------------------------
// Single-line comment with //

/* Multi-line comment
   spanning multiple lines
   should be handled correctly */

/* Nested comment test:
   // Single-line inside multi-line
   Still a comment
*/

// ----------------------------------------------------------------------------
// 2. LABELS - All label styles
// ----------------------------------------------------------------------------
simple_label:
    nop

labelWithUnderscore_123:
    nop

// Label at start of line
start:
    nop

    // Label with indent
    loop:
        nop

// ----------------------------------------------------------------------------
// 3. MNEMONICS - All standard 6502 mnemonics
// ----------------------------------------------------------------------------
all_mnemonics:
    // Load/Store
    lda #$00
    ldx #$00
    ldy #$00
    sta $1000
    stx $1001
    sty $1002

    // Transfer
    tax
    tay
    txa
    tya
    tsx
    txs

    // Stack
    pha
    php
    pla
    plp

    // Arithmetic
    adc #$01
    sbc #$01
    inc $1000
    inx
    iny
    dec $1000
    dex
    dey

    // Logical
    and #$0f
    ora #$80
    eor #$ff
    bit $1000

    // Shift/Rotate
    asl
    asl $1000
    lsr
    lsr $1000
    rol
    rol $1000
    ror
    ror $1000

    // Compare
    cmp #$00
    cpx #$00
    cpy #$00

    // Branch
    bcc *+2
    bcs *+2
    beq *+2
    bne *+2
    bmi *+2
    bpl *+2
    bvc *+2
    bvs *+2

    // Jump/Subroutine
    jmp loop
    jsr subroutine
    rts

    // Flags
    clc
    sec
    cli
    sei
    clv
    cld
    sed

    // System
    brk
    nop
    rti

// ----------------------------------------------------------------------------
// 4. ADDRESSING MODES - All variations
// ----------------------------------------------------------------------------
addressing_modes:
    lda #$42                     // Immediate
    lda $42                      // Zero Page
    lda $42,x                    // Zero Page, X
    ldy $42,x                    // Zero Page, X (with LDY)
    ldx $42,y                    // Zero Page, Y
    lda $1234                    // Absolute
    lda $1234,x                  // Absolute, X
    lda $1234,y                  // Absolute, Y
    lda ($42,x)                  // Indexed Indirect (X)
    lda ($42),y                  // Indirect Indexed (Y)
    jmp ($1234)                  // Indirect (JMP only)
    asl                          // Accumulator
    nop                          // Implied

// ----------------------------------------------------------------------------
// 5. NUMBER FORMATS - Hex, Decimal, Binary, Character
// ----------------------------------------------------------------------------
number_formats:
    .byte $ff                    // Hexadecimal
    .byte 255                    // Decimal
    .byte %11111111              // Binary
    .byte 'A'                    // Character literal

    .word $1234                  // Hex word
    .word 4660                   // Decimal word
    .word %0001001000110100      // Binary word

    // Mixed formats
    .byte $00, 0, %00000000, 'X'

// ----------------------------------------------------------------------------
// 6. STRINGS - All string types
// ----------------------------------------------------------------------------
strings:
    .text "Hello World!"         // Text directive
    .text "Line 1\n"             // With escape (if supported)
    .text 'Single quotes'        // Single-quoted string
    .text ""                     // Empty string

// ----------------------------------------------------------------------------
// 7. DIRECTIVES - Basic data directives
// ----------------------------------------------------------------------------
directives:
    .byte $00, $01, $02          // Byte directive
    .word $1234, $5678           // Word directive
    .fill 256, $00               // Fill directive
    .fill 16, i                  // Fill with counter

// ----------------------------------------------------------------------------
// 8. CONSTANTS AND VARIABLES
// ----------------------------------------------------------------------------
.const SCREEN = $0400
.const BORDER = $d020
.const BACKGROUND = $d021

.var counter = 0
.var temp = $ff

// ----------------------------------------------------------------------------
// 9. EXPRESSIONS - All operators
// ----------------------------------------------------------------------------
expressions:
    .byte 5 + 3                  // Addition
    .byte 10 - 4                 // Subtraction
    .byte 6 * 7                  // Multiplication
    .byte 20 / 4                 // Division
    .byte 17 % 5                 // Modulo
    .byte $ff & $0f              // Bitwise AND
    .byte $f0 | $0f              // Bitwise OR
    .byte $ff ^ $aa              // Bitwise XOR
    .byte $01 << 4               // Left Shift
    .byte $80 >> 2               // Right Shift
    .byte <$1234                 // Low Byte
    .byte >$1234                 // High Byte
    .byte (5 + 3) * 2            // Parentheses
    .byte -42                    // Negation

// ----------------------------------------------------------------------------
// 10. PROGRAM COUNTER EXPRESSIONS
// ----------------------------------------------------------------------------
.pc = $0810 "Start"

pc_expressions:
    .var here = *                // PC as value
    beq *+5                      // Forward relative
    bne *-3                      // Backward relative
    .byte <*, >*                 // PC in expressions

// ----------------------------------------------------------------------------
// 11. SIMPLE MACRO
// ----------------------------------------------------------------------------
.macro nop2() {
    nop
    nop
}

// ----------------------------------------------------------------------------
// 12. SIMPLE FUNCTION
// ----------------------------------------------------------------------------
.function double(x) {
    .return x * 2
}

// ----------------------------------------------------------------------------
// 13. SIMPLE NAMESPACE
// ----------------------------------------------------------------------------
.namespace test {
    .const VALUE = 42

    label:
        nop
}

// ----------------------------------------------------------------------------
// 14. SIMPLE ENUM
// ----------------------------------------------------------------------------
.enum {
    FIRST = 0,
    SECOND = 1,
    THIRD = 2
}

// ----------------------------------------------------------------------------
// 15. MAIN PROGRAM USING ALL SYNTAX ELEMENTS
// ----------------------------------------------------------------------------
main:
    // Use constants
    lda #FIRST
    sta BORDER

    // Use variables
    lda #counter
    sta temp

    // Use expressions
    lda #(SCREEN + 40)

    // Use PC expressions
    bne *+3
    nop

    // Use namespace
    jsr test.label

    // Use macro
    nop2()

    // Use function
    .var result = double(21)

    rts

subroutine:
    nop
    rts

// ============================================================================
// EXPECTED RESULTS:
// ============================================================================
// When testing with: ./kickass_cl 02-basic-syntax.asm
//
// ✅ Parser successfully creates AST
// ✅ No parsing errors
// ✅ All tokens correctly identified
// ✅ All directives recognized
// ✅ All mnemonics valid
// ✅ All addressing modes parsed
// ✅ All number formats accepted
// ✅ All expressions evaluated
// ✅ No unexpected diagnostics
// ============================================================================
