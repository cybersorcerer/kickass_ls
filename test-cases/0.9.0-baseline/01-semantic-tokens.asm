// ============================================================================
// 01-semantic-tokens.asm - Semantic Token Highlighting Test
// ============================================================================
// Test: All token types should be correctly highlighted
// Critical: Tests the fixed enum highlighting bug (v0.9.8)
// ============================================================================

// Expected Token Colors:
// - Comments: gray
// - Mnemonics: magenta
// - Directives: purple
// - Preprocessor: light blue
// - Numbers: orange
// - Strings: green
// - Operators: white
// - Labels: blue
// - Variables: cyan
// - Functions/Macros: yellow

// ----------------------------------------------------------------------------
// 1. PREPROCESSOR DIRECTIVES (should be light blue)
// ----------------------------------------------------------------------------
#import "lib/macros.asm"
#define DEBUG
#undef RELEASE

// ----------------------------------------------------------------------------
// 2. DIRECTIVES (should be purple)
// ----------------------------------------------------------------------------
.encoding "petscii_upper"

.const SCREEN = $0400        // SCREEN: cyan, $0400: orange
.var counter = 0             // counter: cyan, 0: orange

// ----------------------------------------------------------------------------
// 3. CRITICAL TEST: ENUM HIGHLIGHTING (the bug we fixed!)
// ----------------------------------------------------------------------------
// This was broken - numbers showed character-by-character wrong colors
// First digit white, second digit cyan, etc.
.enum {
    BLACK = 0,
    WHITE = 1,
    RED = 2,
    CYAN = 3,
    PURPLE = 4,
    GREEN = 5,
    BLUE = 6,
    YELLOW = 7,
    ORANGE = 8,
    BROWN = 9,
    LIGHT_RED = 10,
    DARK_GRAY = 11,
    GRAY = 12,
    LIGHT_GREEN = 13,
    LIGHT_BLUE = 14,
    LIGHT_GRAY = 15
}

// Second enum with hex numbers
.enum {
    VIC_BORDER = $d020,
    VIC_BACKGROUND = $d021,
    SID_VOLUME = $d418
}

// ----------------------------------------------------------------------------
// 4. FUNCTIONS (name should be yellow)
// ----------------------------------------------------------------------------
.function add(a, b) {        // add: yellow, a: cyan, b: cyan
    .return a + b            // .return: purple, a: cyan, +: white, b: cyan
}

.function multiply(x, y) {   // multiply: yellow, x: cyan, y: cyan
    .return x * y            // x: cyan, *: white, y: cyan
}

// ----------------------------------------------------------------------------
// 5. MACROS (name should be yellow)
// ----------------------------------------------------------------------------
.macro clearScreen() {       // clearScreen: yellow
    lda #$20                 // lda: magenta, #: white, $20: $ white, 20 orange
    ldx #$00                 // ldx: magenta, #: white, $00: $ white, 00 orange
loop:                        // loop: blue (label)
    sta SCREEN,x             // sta: magenta, SCREEN: cyan, ,: white, x: cyan
    sta SCREEN+$100,x        // +: white, $100: $ white, 100 orange
    inx                      // inx: magenta
    bne loop                 // bne: magenta, loop: blue
}

.macro setColor(color) {     // setColor: yellow, color: cyan
    lda #color               // lda: magenta, #: white, color: cyan
    sta $d020                // sta: magenta, $d020: $ white, d020 orange
}

// ----------------------------------------------------------------------------
// 6. PSEUDOCOMMANDS (name should be yellow)
// ----------------------------------------------------------------------------
.pseudocommand mov src : dst {   // mov: yellow, src: cyan, :: white, dst: cyan
    lda src                      // lda: magenta, src: cyan
    sta dst                      // sta: magenta, dst: cyan
}

.pseudocommand inc16 addr {      // inc16: yellow, addr: cyan
    inc addr                     // inc: magenta, addr: cyan
    bne skip                     // bne: magenta, skip: blue
    inc addr+1                   // inc: magenta, addr: cyan, +: white, 1: orange
skip:                            // skip: blue (label)
}

// ----------------------------------------------------------------------------
// 7. NAMESPACES
// ----------------------------------------------------------------------------
.namespace graphics {
    .const SPRITE_PTR = $07f8    // SPRITE_PTR: cyan, $07f8: $ white, 07f8 orange

    clear:                       // clear: blue (label)
        lda #$20                 // lda: magenta
        rts                      // rts: magenta
}

// ----------------------------------------------------------------------------
// 8. MAIN PROGRAM - ALL TOKEN TYPES TOGETHER
// ----------------------------------------------------------------------------
.pc = $0810 "Main"               // .pc: purple, =: white, $0810: $ white, 0810 orange, "Main": green

start:                           // start: blue (label)
    // Test mnemonics (should be magenta)
    lda #$00                     // lda: magenta, #: white, $00: $ white, 00 orange
    ldx #$ff                     // ldx: magenta, #: white, $ff: $ white, ff orange
    ldy #BLACK                   // ldy: magenta, #: white, BLACK: cyan

    // Test operators (should be white)
    sta SCREEN + 5               // +: white
    lda #<$1234                  // <: white (low byte)
    ldx #>$1234                  // >: white (high byte)

    // Test function call (should be yellow)
    .var sum = add(5, 3)         // add: yellow, sum: cyan, 5: orange, 3: orange
    .var product = multiply(4, 7) // multiply: yellow, product: cyan

    // Test macro call (should be yellow)
    clearScreen()                // clearScreen: yellow
    setColor(BLUE)               // setColor: yellow, BLUE: cyan

    // Test pseudocommand call (should be yellow)
    mov #RED : $d020             // mov: yellow, #: white, RED: cyan, :: white
    inc16 counter                // inc16: yellow, counter: cyan

    // Test all number formats (should be orange, with $ and % as white)
    .byte $ff                    // $: white, ff: orange
    .byte 255                    // 255: orange
    .byte %11111111              // %: white, 11111111: orange
    .byte 'A'                    // 'A': orange (character literal)

    // Test strings (should be green)
    .text "Hello World!"         // "Hello World!": green

    // Test expressions (operators white, numbers orange)
    .byte 5 + 3                  // 5: orange, +: white, 3: orange
    .byte 10 - 4                 // 10: orange, -: white, 4: orange
    .byte $ff & $0f              // $: white, ff: orange, &: white, $: white, 0f: orange

    // Test namespace member access
    jsr graphics.clear           // jsr: magenta, graphics.clear: blue

    rts                          // rts: magenta

// ----------------------------------------------------------------------------
// 9. ADDRESSING MODES - ALL MNEMONICS SHOULD BE MAGENTA
// ----------------------------------------------------------------------------
addressing_modes:                // addressing_modes: blue (label)
    lda #$42                     // Immediate
    lda $42                      // Zero Page
    lda $42,x                    // Zero Page, X
    ldx $42,y                    // Zero Page, Y
    lda $1234                    // Absolute
    lda $1234,x                  // Absolute, X
    lda $1234,y                  // Absolute, Y
    lda ($42,x)                  // Indexed Indirect
    lda ($42),y                  // Indirect Indexed
    jmp ($1234)                  // Indirect (JMP only)
    asl                          // Accumulator
    nop                          // Implied

// ----------------------------------------------------------------------------
// EXPECTED RESULTS:
// ============================================================================
// When visualizing with: ./kickass_cl semantic-tokens 01-semantic-tokens.asm
//
// ✅ All comments should be gray
// ✅ All mnemonics (lda, sta, jmp, etc.) should be magenta
// ✅ All directives (.const, .var, .enum, etc.) should be purple
// ✅ All preprocessor (#import, #define) should be light blue
// ✅ All numbers should be orange
// ✅ All $ and % prefixes should be white (operators)
// ✅ All strings should be green
// ✅ All labels should be blue
// ✅ All variables should be cyan
// ✅ All function/macro/pseudocommand names should be yellow
//
// CRITICAL: In enum blocks, numbers like "10", "11", "15" should be
//           COMPLETELY ORANGE - not character-by-character wrong colors!
// ============================================================================
