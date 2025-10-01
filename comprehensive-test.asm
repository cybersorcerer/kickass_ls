// Comprehensive test file for 6510 LSP server
// This file tests various features and edge cases

.const magic_number = $c000
.const zero_page_addr = $80
.const release_build = 0
.const debug_mode = 1

*= $0801
// BASIC SYS 2064
.byte $0c, $08, $0a, $00, $9e, $20, $32, $30, $36, $34, $00, $00, $00

*= $0810
start:
    // Zero page optimization hints
    lda $80          // Should suggest zp addressing
    sta $81          // Should suggest zp addressing

    // Range validation tests
    .const valid_byte = $FF
    .const invalid_byte = $100        // Should warn: out of byte range
    .const valid_word = $FFFF
    .const invalid_word = $10000      // Should warn: out of word range

    // Illegal opcodes
    dcp $ff          // Should warn: illegal opcode

    // Magic number detection
    lda #$c000       // Should hint: matches C64 VIC-II start

    // Branch distance
loop:
    nop
    nop
    // ... many nops to make branch too far
    .fill 130, $EA  // 130 NOPs
    bne loop        // Should warn: branch too far

    // .for loop test
    .for (var i = 0; i < 3; i++) {
        .byte i + $40
    }

    // Memory layout warnings
    sta $d020       // Should be fine (VIC-II)
    sta $a000       // Should warn: ROM area

    // Hardware bug detection
    jmp ($10ff)     // Should warn: JMP indirect bug

end_main:
    rts

.if (debug_mode) {
    .if (release_build) {
        .byte $NE, $VE, $R1                        // Dead: debug_mode=1 AND release_build=0
    } else {
        .byte $OK, $DEBUG                          // This should execute
        sta $d020
    }
} else {
    .byte $RELEASE                                 // Dead: debug_mode=1
}

// Dead code detection - this should be flagged
.if (0) {
    .byte $DE, $AD, $C0, $DE                       // Dead code - never executed
    lda #$dead
    sta $beef
}

// Nested conditionals with complex logic
.if (debug_mode) {
    .if (release_build) {
        .byte $NE, $VE, $R1                        // Dead: debug_mode=1 AND release_build=0
    } else {
        .byte $OK, $DEBUG                          // This should execute
    }
} else {
    .byte $NO, $DEBUG                              // Dead: debug_mode=1
}

// Test data directives with range validation
data_section:
    .byte $00, $FF, $100                          // Last value should warn: out of byte range (line 123)
    .word $0000, $FFFF, $10000                    // Last value should warn: out of word range (line 124)

    // More complex expressions
    .byte magic_number & $FF                      // Should be valid
    .word magic_number | $0F                      // Should be valid

unused_label:
    nop
    rts