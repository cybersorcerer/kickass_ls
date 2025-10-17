// ============================================================================
// 04-completion.asm - Code Completion Test
// ============================================================================
// Test: LSP provides correct completions in different contexts
// Note: Test interactively or via test suite JSON
// ============================================================================

.const SCREEN = $0400
.const BORDER = $d020

start:
    // TEST 1: After '.', should suggest directives only
    // Type: .
    // Expected: .byte, .const, .var, .macro, .function, etc.
    // NOT: lda, sta, SCREEN, etc.

    // TEST 2: After mnemonic space, should suggest addressing modes
    // Type: lda
    // Expected: #, $, (, label names, etc.

    // TEST 3: After #, should suggest constants
    // Type: lda #
    // Expected: SCREEN, BORDER, numbers

    // TEST 4: After jmp/jsr, should suggest labels only
    // Type: jsr
    // Expected: start, loop, subroutine
    // NOT: SCREEN, BORDER (constants)

loop:
    nop
    rts

subroutine:
    nop
    rts

// ============================================================================
// HOW TO TEST COMPLETION:
// ============================================================================
// Using test client:
// ./kickass_cl/kickass_cl completion-at test-cases/0.9.0-baseline/04-completion.asm 13 8
//
// Line 13 is the "// Type: ." line, position 8 is after the dot
//
// Expected: List of directives (.byte, .const, .var, etc.)
// ============================================================================
