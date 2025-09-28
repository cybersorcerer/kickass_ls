// Comprehensive regression test
.const SCREEN = $0400
.var counter = 0

.macro SetColor(color) {
    lda #color
    sta $d020
}

.function Add(a, b) {
    .return a + b
}

.namespace Graphics {
    .const BACKGROUND = $d021
    .label clear:
        lda #0
        sta BACKGROUND
        rts
}

main:
    lda #Add(5, 3)
    sta counter
    +SetColor(1)
    jsr Graphics.clear

    // Test immediate addressing
    lda #$FF
    ldx #%11110000
    ldy #&777

    // Test branches
    bne main
    rts