// Test pseudocommand call argument count validation

.pseudocommand mov src : dst {
    lda src
    sta dst
}

.pseudocommand add16 src1 : src2 : dst {
    clc
    lda src1
    adc src2
    sta dst
}

* = $0801

start:
    mov #$05 : $80           // ✅ Correct: 2 arguments
    mov #$05                 // ❌ Should error: too few arguments (1 instead of 2)
    mov #$05 : $80 : $81     // ❌ Should error: too many arguments (3 instead of 2)

    add16 #$10 : #$20 : $90  // ✅ Correct: 3 arguments
    add16 #$10 : #$20        // ❌ Should error: too few arguments (2 instead of 3)

    rts
