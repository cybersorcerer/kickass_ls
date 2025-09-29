; Comprehensive test file for all LSP features
; This file tests symbols, references, and goto-definition

.const SCREEN_WIDTH = 320
.var screen_height = 200

main:
    lda #sin(45)
    sta screen_buffer
    jsr initialize_screen
    jsr main_loop
    rts

initialize_screen:
    lda #BLACK
    sta $d020
    lda #BLUE
    sta $d021
    rts

main_loop:
    jsr update_display
    jsr handle_input
    jmp main_loop

update_display:
    lda screen_height
    cmp #200
    beq done
    inc screen_height
done:
    rts

handle_input:
    lda $dc01
    and #$1f
    rts

screen_buffer:
    .byte 0

unused_label:
    nop