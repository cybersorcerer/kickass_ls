// Comprehensive Test File for 6510-LSP
// Use this file to test all implemented features.

// --- SECTION 1: Valid Symbol Definitions ---
// This section should parse without any errors.
.macro one_arg_macro(arg) {
  lda arg
}
.const MAX_SPRITES = 8
.const SCREEN_MEM = $0400

.CONST test = $100
.CONST huhu = $100

.var sprite_x_pos = 0

+one_arg_macro(1,2)

// Namespace for graphics routines
.namespace gfx {
    .const BORDER_COLOR = $d020

    .label loop:
        inc BORDER_COLOR
        jmp loop
}

// Global and local labels
.label global_label:
.label .local_label:
    lda #<.local_label
    jmp global_label

start:
    lda #MAX_SPRITES
    sta sprite_x_pos
    jmp gfx.loop


// --- SECTION 2: LSP Feature Tests ---
// Instructions:
// 1. Hover over symbols like `MAX_SPRITES`, `sprite_x_pos`, `gfx.loop`, `.local_label`.
//    -> A tooltip with symbol information should appear.
// 2. Use "Go to Definition" on the same symbols.
//    -> The cursor should jump to their definition.
// 3. Test code completion:
//    - After `lda #`, type `MAX` -> `MAX_SPRITES` should be suggested.
//    - After `jmp `, type `gfx.` -> `loop` should be suggested.
//    - After `jmp `, type `.lo` -> `.local_label` should be suggested.


// --- SECTION 3: Parser Diagnostics (Syntax Errors) ---
// The following lines should be marked as errors by the parser.

    lda #$ffg          // ERROR: Invalid hex number
    .const BAD_CONST = // ERROR: Unexpected token (or missing value)


// --- SECTION 4: Semantic Diagnostics (Logical Errors) ---
// The following lines should be marked as errors by the new semantic analysis.

.const MAX_SPRITES = 99 // ERROR: Duplicate identifier 'MAX_SPRITES'

start:
    nop                // ERROR: Duplicate identifier 'start'

.namespace gfx {
    .label loop:
        rts            // ERROR: Duplicate identifier 'loop' in scope 'gfx'
}
