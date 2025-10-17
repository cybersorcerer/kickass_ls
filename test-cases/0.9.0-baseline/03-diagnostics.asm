// ============================================================================
// 03-diagnostics.asm - Diagnostics Test
// ============================================================================
// Test: LSP server correctly reports errors and warnings
// Expected: Specific errors and warnings at known locations
// ============================================================================

// ----------------------------------------------------------------------------
// SECTION 1: VALID CODE (should have NO diagnostics)
// ----------------------------------------------------------------------------
.const SCREEN = $0400
.var counter = 0

valid_label:
    lda #$00
    sta SCREEN
    rts

// ----------------------------------------------------------------------------
// SECTION 2: UNKNOWN MNEMONICS (should generate ERRORS)
// ----------------------------------------------------------------------------
unknown_mnemonics:
    xyz #$00        // ERROR: Unknown mnemonic 'xyz'
    abc $1000       // ERROR: Unknown mnemonic 'abc'
    def             // ERROR: Unknown mnemonic 'def'

// ----------------------------------------------------------------------------
// SECTION 3: INVALID HEX NUMBERS (should generate ERRORS)
// ----------------------------------------------------------------------------
invalid_hex:
    lda #$gg        // ERROR: Invalid hexadecimal number '$gg'
    ldx #$zz        // ERROR: Invalid hexadecimal number '$zz'
    .byte $xy       // ERROR: Invalid hexadecimal number '$xy'

// ----------------------------------------------------------------------------
// SECTION 4: INVALID BINARY NUMBERS (should generate ERRORS)
// ----------------------------------------------------------------------------
invalid_binary:
    lda #%22222222  // ERROR: Invalid binary number '%22222222'
    ldx #%aaaaaaaa  // ERROR: Invalid binary number '%aaaaaaaa'

// ----------------------------------------------------------------------------
// SECTION 5: UNDEFINED LABELS (should generate WARNINGS/ERRORS)
// ----------------------------------------------------------------------------
undefined_labels:
    jmp unknown_label       // WARNING/ERROR: Undefined label 'unknown_label'
    jsr missing_subroutine  // WARNING/ERROR: Undefined label 'missing_subroutine'
    beq nonexistent         // WARNING/ERROR: Undefined label 'nonexistent'

// ----------------------------------------------------------------------------
// SECTION 6: UNDEFINED SYMBOLS (should generate WARNINGS)
// ----------------------------------------------------------------------------
undefined_symbols:
    lda #undefined_const    // WARNING: Undefined symbol 'undefined_const'
    sta missing_var         // WARNING: Undefined symbol 'missing_var'

// ----------------------------------------------------------------------------
// SECTION 7: DUPLICATE DEFINITIONS (should generate WARNINGS)
// ----------------------------------------------------------------------------
.const DUPLICATE = 10       // First definition (OK)
.const DUPLICATE = 20       // WARNING: Redefinition of 'DUPLICATE'

duplicate_label:            // First definition (OK)
    nop
duplicate_label:            // WARNING: Duplicate label 'duplicate_label'
    nop

// ----------------------------------------------------------------------------
// SECTION 8: UNUSED VARIABLES (should generate HINTS/WARNINGS)
// ----------------------------------------------------------------------------
.var unused_variable = 42   // HINT/WARNING: Unused variable 'unused_variable'
.const UNUSED_CONST = 100   // HINT/WARNING: Unused constant 'UNUSED_CONST'

// ----------------------------------------------------------------------------
// SECTION 9: MACRO ARGUMENT COUNT MISMATCH (should generate ERRORS)
// ----------------------------------------------------------------------------
.macro setColor(color) {
    lda #color
    sta $d020
}

macro_errors:
    setColor(1, 2)          // ERROR: Too many arguments (expected 1, got 2)
    setColor()              // ERROR: Too few arguments (expected 1, got 0)

// ----------------------------------------------------------------------------
// SECTION 10: FUNCTION WITHOUT RETURN (should generate WARNING)
// ----------------------------------------------------------------------------
.function noReturn(x) {     // WARNING: Function 'noReturn' has no return statement
    .var temp = x * 2
}

// ----------------------------------------------------------------------------
// SECTION 11: INVALID ADDRESSING MODES (should generate ERRORS)
// ----------------------------------------------------------------------------
invalid_addressing:
    lda ($42,y)             // ERROR: Invalid addressing mode for LDA (should be ,x)
    ldx $42,x               // ERROR: Invalid addressing mode for LDX (should be ,y)

// ----------------------------------------------------------------------------
// SECTION 12: SYNTAX ERRORS (should generate ERRORS)
// ----------------------------------------------------------------------------
syntax_errors:
    lda #                   // ERROR: Missing operand
    sta                     // ERROR: Missing operand
    .byte                   // ERROR: Missing value
    .const = 10             // ERROR: Missing identifier

// ============================================================================
// EXPECTED DIAGNOSTICS SUMMARY:
// ============================================================================
//
// ERRORS (should halt compilation):
// - Line ~24: Unknown mnemonic 'xyz'
// - Line ~25: Unknown mnemonic 'abc'
// - Line ~26: Unknown mnemonic 'def'
// - Line ~32: Invalid hex '$gg'
// - Line ~33: Invalid hex '$zz'
// - Line ~34: Invalid hex '$xy'
// - Line ~40: Invalid binary '%22222222'
// - Line ~41: Invalid binary '%aaaaaaaa'
// - Line ~88: Too many macro arguments
// - Line ~89: Too few macro arguments
// - Line ~101: Invalid addressing mode
// - Line ~102: Invalid addressing mode
// - Line ~108-111: Syntax errors
//
// WARNINGS (should display but allow compilation):
// - Line ~47: Undefined label 'unknown_label'
// - Line ~48: Undefined label 'missing_subroutine'
// - Line ~49: Undefined label 'nonexistent'
// - Line ~55: Undefined symbol 'undefined_const'
// - Line ~56: Undefined symbol 'missing_var'
// - Line ~62: Redefinition of 'DUPLICATE'
// - Line ~66: Duplicate label 'duplicate_label'
// - Line ~94: Function without return statement
//
// HINTS (suggestions, non-critical):
// - Line ~72: Unused variable 'unused_variable'
// - Line ~73: Unused constant 'UNUSED_CONST'
//
// VALID CODE (should have 0 diagnostics):
// - Lines 7-14: All valid, no errors/warnings
//
// ============================================================================
// HOW TO TEST:
// ============================================================================
// ./kickass_cl/kickass_cl --server ./kickass_ls test-cases/0.9.0-baseline/03-diagnostics.asm
//
// Expected output:
// - Display all errors with line numbers
// - Display all warnings with line numbers
// - Exit with code 1 (due to errors)
// ============================================================================
