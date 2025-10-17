// ============================================================================
// 07-document-symbols.asm - Document Symbols Test
// ============================================================================
// Test: LSP provides complete symbol list for outline view
// ============================================================================

// Constants should appear in symbols list
.const SCREEN = $0400
.const BORDER = $d020
.const BACKGROUND = $d021

// Variables should appear in symbols list
.var counter = 0
.var temp = $ff

// Functions should appear in symbols list
.function add(a, b) {
    .return a + b
}

.function multiply(x, y) {
    .return x * y
}

// Macros should appear in symbols list
.macro clearScreen() {
    lda #$20
    ldx #$00
loop:
    sta SCREEN,x
    inx
    bne loop
}

.macro setColor(color) {
    lda #color
    sta BORDER
}

// Pseudocommands should appear in symbols list
.pseudocommand mov src : dst {
    lda src
    sta dst
}

// Namespace should appear in symbols list with children
.namespace graphics {
    .const SPRITE_PTR = $07f8

    clear:
        nop
        rts
}

// Labels should appear in symbols list
start:
    jsr graphics.clear
    clearScreen()
    setColor(1)
    rts

subroutine:
    nop
    rts

// ============================================================================
// EXPECTED SYMBOLS LIST:
// ============================================================================
// Constants:
// - SCREEN (line 8)
// - BORDER (line 9)
// - BACKGROUND (line 10)
//
// Variables:
// - counter (line 13)
// - temp (line 14)
//
// Functions:
// - add (line 17)
// - multiply (line 21)
//
// Macros:
// - clearScreen (line 26)
// - setColor (line 35)
//
// Pseudocommands:
// - mov (line 41)
//
// Namespaces:
// - graphics (line 47)
//   - SPRITE_PTR (line 48)
//   - clear (line 50)
//
// Labels:
// - start (line 56)
// - subroutine (line 62)
//
// Total: ~15-17 symbols (depending on implementation)
// ============================================================================
