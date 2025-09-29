start:
    lda #$02
    sta $D020    ; Border color
    sta $D021    ; Background color
    lda $DC0D    ; CIA1 interrupt control
    sta $DD0D    ; CIA2 interrupt control