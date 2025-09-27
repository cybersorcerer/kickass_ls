// Enhanced Semantic Analysis Features Test File
// Test all new 6502/C64 specific analysis features

  *=$1000
// ===========================================
// SECTION 1: Branch Distance Validation
// ===========================================

start:
    lda #$01
    ldx #$02

// This should work fine (short distance)
short_loop:
    dex
    bne short_loop      // OK: Short branch distance

// Test long branch (should trigger warning)
    jmp far_away        // Jump to create distance

// Fill some space to create distance
.for (var i=0; i<100; i++) {
    nop
}

near_target:
    bne near_target     // OK: Very short branch

// Create even more distance
.for (var i=0; i<50; i++) {
    nop
}

far_away:
    // This branch will be too far (should trigger ERROR)
    bne start           // ERROR: Branch distance > 127 bytes

// ===========================================
// SECTION 2: Illegal Opcode Detection
// ===========================================

legal_code:
    lda #$ff           // OK: Legal opcode
    sta $d020          // OK: Legal opcode

illegal_code:
    SLO $d020          // WARNING: Illegal opcode (Shift Left then OR)
    LAX $d021          // WARNING: Illegal opcode (Load A and X)
    SAX $80            // WARNING: Illegal opcode (Store A AND X)
    DCP $81            // WARNING: Illegal opcode (Decrement then Compare)
    ISC $82            // WARNING: Illegal opcode (Increment then Subtract)

// ===========================================
// SECTION 3: 6502 Hardware Bug Detection
// ===========================================

hardware_bugs:
    // JMP indirect page boundary bug test
    jmp ($20ff)        // WARNING: Page boundary bug - reads from $20FF and $2000!
    jmp ($21fe)        // OK: No page boundary issue
    jmp ($30ff)        // WARNING: Another page boundary bug
    jmp (vector)       // OK: Regular indirect jump

vector: .word $1234

// ===========================================
// SECTION 4: Zero Page Optimization
// ===========================================

zero_page_tests:
    lda $0080          // HINT: Consider zero-page addressing for $80
    sta $00ff          // HINT: Consider zero-page addressing for $FF
    lda $0200          // OK: Not zero page

    // Already optimized zero page
    lda $80            // OK: Already zero page
    sta $ff            // OK: Already zero page

// ===========================================
// SECTION 5: Magic Number Detection
// ===========================================

magic_numbers:
    lda #$ff
    sta $d020          // HINT: Consider constant for Border color register
    sta $d021          // HINT: Consider constant for Background color register

    lda $0314          // HINT: Consider constant for IRQ vector (low byte)
    sta $0315          // HINT: Consider constant for IRQ vector (high byte)

    jmp $fffc          // HINT: Consider constant for Reset vector

    // Good practice - using constants
    .const BORDER_COLOR = $d020
    .const BG_COLOR = $d021

    lda #$01
    sta BORDER_COLOR   // OK: Using constant
    sta BG_COLOR       // OK: Using constant

// ===========================================
// SECTION 6: Dead Code Detection
// ===========================================

dead_code_test:
    lda #$01
    jmp end_section    // Unconditional jump

    // All code below should be marked as UNREACHABLE
    nop                // WARNING: Unreachable code after unconditional jump
    lda #$02           // WARNING: Unreachable code after unconditional jump
    sta $d020          // WARNING: Unreachable code after unconditional jump

reachable_again:       // OK: Label makes code reachable again
    lda #$03
    rts                // Another unconditional exit

    nop                // WARNING: Unreachable code after unconditional jump

end_section:           // OK: Jump target, code is reachable

// ===========================================
// SECTION 7: Memory Layout Awareness Tests
// ===========================================

memory_tests:
    // These should work but generate info messages
    lda $d000          // INFO: I/O register access
    sta $d001          // INFO: I/O register access

    // Stack area warnings
    lda $0150          // INFO: Reading from stack area
    sta $0160          // WARNING: Writing to stack area - may corrupt stack

// ===========================================
// SECTION 8: Style Guide Tests
// ===========================================

// Good style
.const MAX_SPRITES = 8        // OK: UPPER_CASE constant
.const SCREEN_WIDTH = 320     // OK: UPPER_CASE constant

// Bad style (should generate hints)
.const bad_constant = 42      // HINT: Consider UPPER_CASE for constant

good_label:                   // OK: Descriptive name
    nop

a:                           // HINT: Consider more descriptive name
    nop

// ===========================================
// SECTION 9: Complex Control Flow
// ===========================================

complex_flow:
    ldx #$10

loop_start:
    dex
    beq loop_end
    bne loop_start     // OK: Normal loop

    // Unreachable
    nop                // WARNING: Unreachable after unconditional branch

loop_end:
    rts

// Infinite loop detection
infinite_test:
    jmp infinite_test  // WARNING: Potential infinite loop detected

// ===========================================
// END OF TEST FILE
// ===========================================

// Test file summary:
// - Branch distance validation (short OK, long ERROR)
// - Illegal opcode warnings (SLO, LAX, SAX, DCP, ISC)
// - Hardware bug detection (JMP $xxFF page boundary)
// - Zero page optimization hints ($0080 -> $80)
// - Magic number detection (C64 addresses)
// - Dead code detection after JMP/RTS
// - Memory layout awareness (I/O, stack)
// - Style guide enforcement (constants, labels)
// - Complex control flow analysis
