// Test file for mnemonic type classification
// Tests: Branch Distance Validation, Jump Target Validation, Dead Code Detection, Completion Context

.const SCREEN = $0400
.const BORDER = $d020

start:
    lda #0
    sta SCREEN

    // TEST 1: Branch Distance Validation (should validate ±128 bytes)
    beq nearby_label      // OK - Branch type
    bne nearby_label      // OK - Branch type
    bcc nearby_label      // OK - Branch type

    // TEST 2: Jump Target Validation (should check forward references)
    jmp forward_label     // OK - Jump type, forward reference allowed
    jsr subroutine        // OK - Jump type, forward reference allowed

    // TEST 3: Dead Code Detection (after unconditional jumps)
    jmp end              // Unconditional Jump - dead code after this
    lda #$FF             // WARNING: Dead code after JMP

nearby_label:
    nop

    rts                  // Return type - dead code after this
    sta BORDER           // WARNING: Dead code after RTS

forward_label:
    nop

    rti                  // Return type - dead code after this
    lda #$00             // WARNING: Dead code after RTI

subroutine:
    lda #1
    rts                  // Return - ends subroutine

end:
    jmp end              // Infinite loop

// TEST 4: Completion Context (manual test in interactive mode)
// Position after "jmp " should suggest ONLY labels
// Position after "beq " should suggest ONLY labels
// Position after "lda " should suggest all operands
