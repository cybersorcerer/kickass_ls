// Test Case 09: Program Counter Expressions
// Purpose: Test * (PC) in expressions
// Issue #4: Program Counter Expressions

* = $0801

// Basic PC reference
.label current_address = *

// PC with offset (forward branch)
start:
    lda #$00
    beq *+5        // Branch forward 5 bytes
    nop
    nop
    rts

// PC with offset (backward branch)
loop:
    inc $d020
    dex
    bne *-5        // Branch backward to loop

    rts

// PC in expressions
.const CODE_START = *
    lda #$01
    sta $d020
.const CODE_END = *
.const CODE_SIZE = CODE_END - CODE_START

// PC with alignment
.align 256
page_aligned:
    .label page_start = *
    nop
    rts

// Relative addressing with PC
relative_jump:
    jmp * + 10
    nop
    nop
    nop
    nop
    nop
target:
    rts

// PC in data directives
.byte <*, >*      // Low and high byte of current PC
