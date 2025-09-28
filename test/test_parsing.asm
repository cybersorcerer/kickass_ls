.const MAX_SPRITES = 8
.var sprite_count = 0

start:
    lda #MAX_SPRITES
    sta sprite_count
    tax
    jmp end

end:
    rts

// Test illegal hex
lda #$40G