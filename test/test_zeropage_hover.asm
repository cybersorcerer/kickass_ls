start:
    lda #$37
    sta $0001    ; Processor port
    lda #$37
    sta $0000    ; Processor port direction