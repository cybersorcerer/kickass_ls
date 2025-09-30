unused_label:
    lda #$00
    sta $01FF
    nop
    nop
    lax #$00

start:
    jmp start