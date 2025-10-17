// ============================================================================
// 06-goto-definition.asm - Go to Definition / Find References Test
// ============================================================================
// Test: LSP correctly finds definitions and references
// ============================================================================

.const SCREEN = $0400       // Definition of SCREEN (line 7)
.var counter = 0            // Definition of counter (line 8)

start:                      // Definition of start (line 10)
    lda #SCREEN             // Reference to SCREEN - should jump to line 7
    sta counter             // Reference to counter - should jump to line 8
    jsr subroutine          // Reference to subroutine - should jump to line 16
    jmp loop                // Reference to loop - should jump to line 14

loop:                       // Definition of loop (line 14)
    rts

subroutine:                 // Definition of subroutine (line 16)
    lda #SCREEN             // Reference to SCREEN - should jump to line 7
    sta counter             // Reference to counter - should jump to line 8
    rts

// ============================================================================
// EXPECTED BEHAVIOR:
// ============================================================================
// Go to Definition:
// - Click on SCREEN at line 11 → Jump to line 7
// - Click on counter at line 12 → Jump to line 8
// - Click on subroutine at line 13 → Jump to line 16
// - Click on loop at line 13 → Jump to line 14
//
// Find References (for SCREEN):
// - Line 7 (definition)
// - Line 11 (usage)
// - Line 17 (usage)
// Total: 3 locations
//
// Find References (for counter):
// - Line 8 (definition)
// - Line 12 (usage)
// - Line 18 (usage)
// Total: 3 locations
// ============================================================================
