// Simple test for built-ins
.const VALUE = sin(PI)
.var color = RED

start:
    lda #floor(3.14)
    sta $02
    rts