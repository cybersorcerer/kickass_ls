.const valid = $DEAD
.const invalid = $BUG
start:
    lda undefined_symbol
    dcp $ff