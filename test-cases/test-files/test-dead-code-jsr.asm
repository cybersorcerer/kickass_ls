// Test: JSR should NOT cause dead code warnings
// Bug fix: v1.0.1 - JSR was incorrectly treated as unconditional jump

*=$0800

start:
    jsr subroutine
    // These lines should NOT be marked as unreachable
    // JSR is a subroutine call that returns
    lda #$01
    sta $d020
    rts

subroutine:
    lda #$00
    sta $d021
    rts
