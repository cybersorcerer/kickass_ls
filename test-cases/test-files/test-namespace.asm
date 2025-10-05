// Test .namespace directive parsing

.namespace utils {
    .const SCREEN = $0400

    clear:
        lda #$20
        ldx #$00
    loop:
        sta SCREEN,x
        inx
        bne loop
        rts
}

.namespace graphics {
    .const COLORS = 16

    setColor:
        sta $d020
        rts
}

// Duplicate namespace (should warn)
.namespace utils {
    another:
        rts
}

* = $0801
start:
    jsr utils.clear
    lda #14
    jsr graphics.setColor
    rts
