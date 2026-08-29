// Test Case 02: Preprocessor symbol definitions
// Purpose: Test symbol-only definitions (no value) and conditional assembly
// Issue #3: Directive Parameter Parsing
//
// Kick Assembler defines preprocessor symbols with #define, not .define.
// The .define directive is something else entirely: it runs a block of
// directives in function mode (manual table A.7). Conditional assembly uses
// #if / #else / #endif; there is no .ifdef, .ifndef, .endif or .undef.

* = $0801

// Basic definition (symbol only, no value)
#define DEBUG
#define RELEASE_MODE
#define ENABLE_SOUND

// Conditional compilation based on definitions
#if DEBUG
    nop
    nop
#endif

#if ENABLE_SOUND
    lda #$ff
    sta $d020
#endif

// Redefinition without an intervening #undef (should warn)
#define ENABLE_SOUND

// Removing a definition
#undef DEBUG

start:
    lda #$00
    rts
