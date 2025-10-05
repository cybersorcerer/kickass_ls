// Test Case 08: .import Directive
// Purpose: Test keyword + string parameter
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic import with source keyword
// .import source "lib/graphics.asm"
// .import source "lib/sound.asm"

// Import with binary keyword
// .import binary "data/sprites.bin"
// .import c64 "music/tune.sid"

// Import with conditional
.ifdef USE_CUSTOM_CHARSET
    // .import binary "charset.bin" at $2000
.endif

start:
    // Use imported symbols (would come from imported files)
    // jsr graphics.init
    // jsr sound.play
    lda #$00
    sta $d020
    rts

// Invalid import (file doesn't exist - should warn)
// .import source "nonexistent/file.asm"

// Import without quotes (should error)
// .import source filename.asm
