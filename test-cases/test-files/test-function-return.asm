// Test .function with missing .return

.function hasReturn(x) {
    .return x * 2
}

.function noReturn(x) {
    .var temp = x * 2
}

* = $0801
start:
    lda #$00
    rts
