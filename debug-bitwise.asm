.const base = $1000
.const shifted_left = base << 2
.const shifted_right = base >> 4
.const masked = base & $FF
.const ored = base | $0F
.const xored = base ^ $AA

start:
    lda undefined_symbol
    dcp $ff
    rts