// ============================================================================
// 05-hover.asm - Hover Information Test
// ============================================================================
// Test: LSP provides hover info for symbols and mnemonics
// Note: Test interactively or check hover content
// ============================================================================

.const SCREEN = $0400       // Hover: Should show "SCREEN = $0400"
.var counter = 0            // Hover: Should show "counter = 0"

start:                      // Hover: Should show "Label: start"
    lda #$00                // Hover on 'lda': Should show "LDA - Load Accumulator"
    sta SCREEN              // Hover on 'SCREEN': Should show constant value
    jsr subroutine          // Hover on 'subroutine': Should show label info
    rts

subroutine:
    nop
    rts

// ============================================================================
// EXPECTED HOVER CONTENT:
// ============================================================================
// Line 8 (SCREEN): "Constant: SCREEN = $0400"
// Line 9 (counter): "Variable: counter = 0"
// Line 11 (start): "Label: start at line 11"
// Line 12 (lda): "LDA - Load Accumulator\nAffects: N, Z"
// Line 13 (SCREEN): "Constant: SCREEN = $0400"
// Line 14 (subroutine): "Label: subroutine at line 18"
// ============================================================================
