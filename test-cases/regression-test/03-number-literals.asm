// Regression Test 03: Number Literal Formats
// Purpose: Ensure all number formats are recognized
// Status: Should PASS with 0 errors

* = $0801

start:
    // Hexadecimal literals
    lda #$00
    lda #$ff
    lda #$d020
    lda #$FFFF

    // Decimal literals
    lda #0
    lda #10
    lda #255
    lda #65535

    // Binary literals
    lda #%00000000
    lda #%11111111
    lda #%10101010
    lda #%01010101

    // Octal literals (if supported)
    // lda #@777

    // Character literals
    lda #'A'
    lda #'Z'
    lda #'0'
    lda #' '

    // Mixed in expressions
    lda #$10 + 5
    lda #%11110000 | $0f
    lda #'A' + 1

    rts
