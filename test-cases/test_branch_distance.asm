; Test file for branch distance validation
start:
    lda #$00
    ; This should trigger a branch distance error - too far
target:
    .fill 200, $00  ; Fill 200 bytes to create distance > 127
    bne start       ; This branch should be out of range
    rts