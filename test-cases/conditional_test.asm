.const debug_mode = 1
.const release_build = 0

// Simple .if directive
.if (debug_mode) {
    lda #$ff
    sta $d020
} else {
    lda #$00
    sta $d020
}

// Dead code detection
.if (0) {
    // This code should be flagged as dead code
    lda #$dead
    sta $beef
}

// .ifdef directive
.ifdef debug_mode {
    .byte $de, $bu, $g1
}

// .ifndef directive  
.ifndef undefined_symbol {
    .byte $no, $tu, $nd
}

// Nested conditionals
.if (debug_mode) {
    .if (release_build) {
        // This should be dead code since release_build = 0
        .byte $never, $reached
    } else {
        .byte $debug, $active
    }
}

start:
    rts