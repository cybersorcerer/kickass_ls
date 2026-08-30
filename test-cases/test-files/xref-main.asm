#import "xref-constants.asm"

*=$0801

start:
        lda #$00
        sta BORDER_COLOR
        sta SCREEN_RAM
        rts
