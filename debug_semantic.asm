.const base = $1000

start:
    lda undefined_symbol
    dcp $ff
    rts