// Test Case 07: Diagnostic Detection
// Purpose: Test error and warning detection

* = $0810

// Valid code (no errors)
start:
    lda #$00
    sta $d020

// Error 1: Mnemonic without operand
    lda
    sta

// Error 2: Invalid hex value
    lda #$XY
    sta #$GG

// Error 3: Invalid addressing mode for mnemonic
    jmp #$1000      // JMP doesn't support immediate

// Error 4: Missing operand with trailing spaces
    ldx
    ldy

// Valid immediate mode
    lda #$20

// Error 5: Invalid binary value
    lda #%2222

// Valid code
    ldx #$ff
    jsr $1000

// Error 6: Undefined label reference (warning/hint)
loop:
    jmp undefined_label

// Valid label
end:
    rts

// Error 7: Invalid character literal
    lda #''
    sta #'AB'

// Valid character literal
    lda #'A'
