// Test: JMP and RTS SHOULD cause dead code warnings
// Ensure that actual dead code detection still works

*=$0800

start:
    lda #$00
    sta $d020
    jmp end      // Unconditional jump

    // These lines SHOULD be marked as unreachable (dead code)
    lda #$01     // Line 11 - should warn
    sta $d021    // Line 12 - should warn
    nop          // Line 13 - should warn

someLabel:       // Label makes code reachable again
    lda #$02
    sta $d020

end:
    rts

    // Code after RTS SHOULD be unreachable
    lda #$ff     // Line 23 - should warn
    nop          // Line 24 - should warn
