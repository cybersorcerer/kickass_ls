.const MAX_SPRITES = 8
.var count = 0

start:
    lda #MAX_SPRITES
    sta count
    rts