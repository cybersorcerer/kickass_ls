.const BORDER_COLOR = $D020
.const BACKGROUND = $D021

init:
    lda #$0F
    sta BORDER_COLOR
    sta BACKGROUND
    jsr setup_screen
    rts

setup_screen:
    lda #$00
    sta BORDER_COLOR
    jmp main_loop

main_loop:
    inc BACKGROUND
    jsr setup_screen
    jmp main_loop

end_program:
    brk