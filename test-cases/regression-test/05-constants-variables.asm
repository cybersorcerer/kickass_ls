// Regression Test 05: Constants and Variables
// Purpose: Ensure .const and .var directives work
// Status: Should PASS with 0 errors

* = $0801

// Constants (immutable)
.const SCREEN = $0400
.const COLOR_RAM = $d800
.const BORDER = $d020
.const BACKGROUND = $d021

// Variables (mutable in Kick Assembler)
.var counter = 0
.var address = $1000
.var color = 14

// Expressions in constants
.const SPRITE_0 = $07f8
.const SPRITE_BASE = SCREEN + $3f8
.const DATA_SIZE = 256

start:
    // Use constants
    lda #$00
    sta BORDER
    sta BACKGROUND

    // Use in addressing
    lda SCREEN
    sta COLOR_RAM

    // Use variables
    lda #counter
    sta $80

    lda #color
    sta BORDER

    rts
