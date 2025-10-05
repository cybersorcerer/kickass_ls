// Regression Test 08: Expressions and Operators
// Purpose: Ensure expression evaluation works correctly
// Status: Should PASS with 0 errors

* = $0801

// Arithmetic operators
.const ADD_RESULT = 10 + 20
.const SUB_RESULT = 100 - 50
.const MUL_RESULT = 5 * 8
.const DIV_RESULT = 100 / 10

// Bitwise operators
.const AND_RESULT = $ff & $0f
.const OR_RESULT = $f0 | $0f
.const XOR_RESULT = $ff ^ $aa
.const NOT_RESULT = ~$00

// Shift operators
.const SHL_RESULT = $01 << 4
.const SHR_RESULT = $80 >> 4

// Comparison operators (if supported)
// .const CMP_RESULT = (5 > 3)

// Byte extraction
.const HIGH_BYTE = >$1234
.const LOW_BYTE = <$1234

start:
    // Use expression results
    lda #ADD_RESULT
    sta $80

    lda #HIGH_BYTE
    sta $81

    lda #LOW_BYTE
    sta $82

    // Expressions in instructions
    lda #10 + 5
    sta $d020

    lda #$ff & $0f
    sta $d021

    // Complex expressions
    lda #((100 + 50) / 3)
    sta $83

    rts

// Expression in label
.label result = ADD_RESULT + SUB_RESULT
