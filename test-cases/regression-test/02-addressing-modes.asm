// Regression Test 02: All 13 Addressing Modes
// Purpose: Ensure all addressing modes parse correctly
// Status: Should PASS with 0 errors

* = $0801

start:
    // 1. Immediate
    lda #$00
    ldx #$ff
    ldy #$10

    // 2. Zero Page
    lda $80
    ldx $81
    ldy $82

    // 3. Zero Page,X
    lda $80, x
    ldy $81, x
    sta $82, x

    // 4. Zero Page,Y
    ldx $80, y
    stx $81, y

    // 5. Absolute
    lda $d020
    sta $0400
    jmp $1000

    // 6. Absolute,X
    lda $1000, x
    sta $0400, x

    // 7. Absolute,Y
    lda $1000, y
    sta $0400, y

    // 8. Indexed Indirect (Zero Page,X)
    lda ($80, x)
    sta ($82, x)

    // 9. Indirect Indexed (Zero Page),Y
    lda ($80), y
    sta ($82), y

    // 10. Indirect (JMP only)
    jmp ($0310)

    // 11. Accumulator
    asl a
    lsr a
    rol a
    ror a

    // 12. Implied
    nop
    tax
    tay
    rts

    // 13. Relative (branches)
loop:
    dec $80
    bne loop

    rts
