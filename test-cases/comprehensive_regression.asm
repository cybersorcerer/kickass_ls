.const base = $1000
.const unused_const = 42

// Illegal characters from bitwise operators
.const shifted = base << 2
.const masked = base & $FF
.const ored = base | $0F
.const xored = base ^ $AA

start:
    lda #$01
    sta base
    
    // These should generate errors
    lda undefined_symbol
    jmp unknown_label
    
    // This should generate warnings
    dcp $ff          // Illegal opcode
    
    // Unused symbols should generate hints
    
    rts

// Unreferenced label should generate warnings
unreferenced_label:
    nop
    rts