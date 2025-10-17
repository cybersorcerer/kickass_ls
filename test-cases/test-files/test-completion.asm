// Test completion at various positions

.const SCREEN = $0400
.const BORDER = $d020

.macro clearScreen() {
    nop
}

start:
    // Position 1: After dot - should show ONLY directives
    .

    // Position 2: After space - should show mnemonics


    // Position 3: After mnemonic - should show addressing modes/operands
    lda

    // Position 4: After # - should show constants
    lda #

    // Position 5: After jmp/jsr - should show ONLY labels
    jmp

    // Position 6: In operand - should show constants/labels
    sta

loop:
    nop

subroutine:
    rts
