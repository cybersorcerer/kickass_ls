start:
    lda undefined_symbol
    sta $INVALID
    bne too_far_branch
    xyz #$00

too_far_branch: