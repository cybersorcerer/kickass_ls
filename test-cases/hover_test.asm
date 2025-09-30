.const SCREEN_COLOR = $D021

start:
    lda #$FF
    sta SCREEN_COLOR
    jmp loop

loop:
    inc $D020
    jmp loop