// Regression Test 09: Comment Support
// Purpose: Ensure all comment styles are supported
// Status: Should PASS with 0 errors

* = $0801

// Line comment style 1 (C++)

; Line comment style 2 (Assembly)

/* Block comment style 1 */

/*
 * Multi-line block comment
 * with multiple lines
 */

start:
    lda #$00    // End of line comment
    sta $d020   ; Another end of line comment

    /* Inline block comment */ nop

    // Code with comments
    ldx #$00    // Counter
loop:           // Loop label
    inx         // Increment
    cpx #$10    // Compare with 16
    bne loop    // Branch if not equal

    rts         // Return

// Commented-out code should not execute
// lda #$ff
// sta $d020

; Another commented section
; .byte $00, $01, $02

/*
 * Complex multi-line comment
 * with various content:
 * - Instructions: lda #$00
 * - Directives: .byte $ff
 * - Expressions: 10 + 20
 */

end:
    rts
