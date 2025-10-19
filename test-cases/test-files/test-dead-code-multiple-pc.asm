// Test: Multiple *= or .pc directives should not cause false warnings
// Each PC directive starts a new code section

*=$0800 "BASIC Start"
    .byte $00
    .byte $0E, $08
    .byte $0A, $00
    .byte $9E
    .byte $20, $28, $34, $30, $39, $36, $29
    .byte $00, $00, $00

.const SCREEN = $0400

*=$1000 "Main Start"
    // This code should NOT be marked as unreachable
    // even though there's no JMP from the previous section
    lda #$00
    sta $d020
    jsr subroutine
    rts

subroutine:
    lda #$01
    sta $d021
    rts

.pc = $2000 "Data Section"
    // This should also NOT be unreachable
    .byte $00, $01, $02, $03
