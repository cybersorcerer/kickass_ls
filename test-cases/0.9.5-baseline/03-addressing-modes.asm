// Test Case 03: Addressing Modes
// Purpose: Test all 6510 addressing modes

* = $0810

// Immediate addressing
    lda #$00
    ldx #$ff
    ldy #$80
    cmp #$42

// Zero Page
    lda $80
    sta $90
    ldx $a0
    stx $b0

// Zero Page, X
    lda $80, x
    sta $90, x
    ldy $a0, x
    sty $b0, x

// Zero Page, Y
    ldx $80, y
    stx $90, y

// Absolute
    lda $d020
    sta $d021
    jmp $0810
    jsr $1000

// Absolute, X
    lda $1000, x
    sta $2000, x
    inc $d000, x

// Absolute, Y
    lda $1000, y
    sta $2000, y

// Indirect
    jmp ($0310)

// Indexed Indirect (X)
    lda ($80, x)
    sta ($90, x)
    cmp ($a0, x)

// Indirect Indexed (Y)
    lda ($80), y
    sta ($90), y
    cmp ($a0), y

// Accumulator
    asl a
    lsr a
    rol a
    ror a

// Implied
    nop
    rts
    rti
    tax
    tay
    txa
    tya
    txs
    tsx
    pha
    pla
    php
    plp
    clc
    sec
    cli
    sei
    clv
    cld
    sed

// Relative (branches)
    beq *+5
    bne *-3
    bcc forward
    bcs backward
backward:
    bmi forward
    bpl backward
    bvc forward
    bvs backward
forward:
    nop
