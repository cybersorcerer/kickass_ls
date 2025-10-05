// Test Case 06: .pseudocommand Directive
// Purpose: Test colon-separated parameter syntax
// Issue #3: Directive Parameter Parsing

* = $0801

// Basic pseudocommand with colon-separated parameters
.pseudocommand add src : dest {
    clc
    lda src
    adc dest
    sta dest
}

// Pseudocommand with multiple parameters
.pseudocommand move16 source : destination {
    lda source
    sta destination
    lda source+1
    sta destination+1
}

// Pseudocommand with immediate and memory parameters
.pseudocommand inc16 addr {
    inc addr
    bne skip
    inc addr+1
skip:
}

// Complex pseudocommand
.pseudocommand call routine : param1 : param2 {
    lda #param1
    ldx #param2
    jsr routine
}

start:
    // Use pseudocommands
    add #$10 : $80
    move16 $fb : $fd
    inc16 $1000
    call subroutine : $05 : $0a
    rts

subroutine:
    rts

// Invalid pseudocommand (missing parameters, should error)
test:
    add #$10
