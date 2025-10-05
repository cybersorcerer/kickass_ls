// Regression Test 01: Basic 6510 Mnemonics
// Purpose: Ensure all standard mnemonics are recognized
// Status: Should PASS with 0 errors

* = $0801

start:
    // Load/Store operations
    lda #$00
    ldx #$ff
    ldy #$10
    sta $d020
    stx $d021
    sty $0400

    // Transfer operations
    tax
    tay
    txa
    tya
    txs
    tsx

    // Stack operations
    pha
    pla
    php
    plp

    // Arithmetic
    adc #$01
    sbc #$01
    inc $80
    dec $80
    inx
    iny
    dex
    dey

    // Logic operations
    and #$0f
    ora #$f0
    eor #$ff
    bit $80

    // Shifts/rotates
    asl a
    lsr a
    rol a
    ror a
    asl $80
    lsr $80

    // Branch operations
    beq start
    bne start
    bcc start
    bcs start
    bpl start
    bmi start
    bvc start
    bvs start

    // Jump/call
    jmp end
    jsr subroutine

    // Flags
    clc
    sec
    cli
    sei
    cld
    sed
    clv

    // Compare
    cmp #$00
    cpx #$00
    cpy #$00

    // Misc
    nop
    brk

subroutine:
    rts

end:
    rts
