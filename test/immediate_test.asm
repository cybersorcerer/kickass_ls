start:
    lda #$01
    ldx #$FF
    ldy #MAX_SPRITES
    cmp #%01010101
    and #&77
    rts