start:
    lda #$1C
    sta $D400    ; SID Voice 1 frequency low
    lda #$10
    sta $D401    ; SID Voice 1 frequency high
    lda #%00010001
    sta $D404    ; SID Voice 1 control
    lda #$0F
    sta $D418    ; SID Volume