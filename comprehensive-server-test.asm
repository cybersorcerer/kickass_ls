// ============================================================================
// COMPREHENSIVE KICK ASSEMBLER LSP SERVER TEST
// ============================================================================
// This file tests ALL features of the Kick Assembler LSP Server v0.9.7
// Use this in Neovim to verify: parsing, diagnostics, completion, hover, etc.
// ============================================================================

// ----------------------------------------------------------------------------
// 1. ENCODING DIRECTIVE
// ----------------------------------------------------------------------------
.encoding "petscii_upper"           // ✅ Valid encoding
.encoding "screencode_mixed"        // ✅ Valid encoding
.encoding "invalid_name"         // ⚠️ Should warn: Unknown encoding

// ----------------------------------------------------------------------------
// 2. DEFINE/UNDEF DIRECTIVES
// ----------------------------------------------------------------------------
.define DEBUG                       // ✅ Symbol-only define
.define RELEASE                     // ✅ Another define
.define DEBUG                    // ⚠️ Should warn: Redefinition

.undef RELEASE                      // ✅ Undefine symbol
.undef UNKNOWN                   // ⚠️ Could warn: Symbol not defined

// ----------------------------------------------------------------------------
// 3. IMPORT DIRECTIVE
// ----------------------------------------------------------------------------
.import source "lib/macros.asm"     // ✅ Import source file
.import binary "data/charset.bin"   // ✅ Import binary data
.import c64 "music/tune.sid"        // ✅ Import C64 file

// ----------------------------------------------------------------------------
// 4. CONSTANTS AND VARIABLES
// ----------------------------------------------------------------------------
.const SCREEN = $0400               // ✅ Constant definition
.const CHARSET = $2000              // ✅ Another constant
.const BORDER = $d020               // ✅ C64 memory address
.const BACKGROUND = $d021           // ✅ C64 memory address
.const new_background = $d021

.var counter = 0                    // ✅ Variable definition
.var temp = $ff                     // ✅ Variable with hex value

// ----------------------------------------------------------------------------
// 5. NAMESPACE DIRECTIVE
// ----------------------------------------------------------------------------
.namespace graphics {
    .const SPRITE_PTR = $07f8

    clear:
        lda #$20
        ldx #$00
    loop:
        sta SCREEN,x
        sta SCREEN+$100,x
        sta SCREEN+$200,x
        sta SCREEN+$300,x
        inx
        bne loop
        rts
}

.namespace sound {
    .const SID_BASE = $d400
    .const SID_VOL = SID_BASE + 24

    init:
        lda #$0f
        sta SID_VOL
        rts
}

// ----------------------------------------------------------------------------
// 6. ENUM DIRECTIVE
// ----------------------------------------------------------------------------
.enum Colors {
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

.enum Registers {
    VIC_BORDER = $d020,
    VIC_BACKGROUND = $d021,
    SID_VOLUME = $d418
}

// ----------------------------------------------------------------------------
// 7. FUNCTION DIRECTIVE
// ----------------------------------------------------------------------------
.function add(a, b) {
    .return a + b
}

.function multiply(x, y) {
    .return x * y
}

.function square(n) {
    .return n * n
}

.function noReturn(x) {          // ⚠️ Should warn: No .return statement
     .var temp = x * 2
}

// ----------------------------------------------------------------------------
// 8. MACRO DIRECTIVE
// ----------------------------------------------------------------------------
.macro clearScreen() {
    lda #$20
    ldx #$00
clearLoop:
    sta SCREEN,x
    sta SCREEN+$100,x
    sta SCREEN+$200,x
    sta SCREEN+$300,x
    inx
    bne clearLoop
}

.macro setColor(color) {
    lda #color
    sta BORDER
}

.macro setBorderAndBackground(borderColor, bgColor) {
    lda #borderColor
    sta BORDER
    lda #bgColor
    sta BACKGROUND
}

.macro copyMemory(source, dest, count) {
    ldx #0
copyLoop:
    lda source,x
    sta dest,x
    inx
    cpx #count
    bne copyLoop
}

// ----------------------------------------------------------------------------
// 9. PSEUDOCOMMAND DIRECTIVE
// ----------------------------------------------------------------------------
.pseudocommand mov src : dst {
    lda src
    sta dst
}

.pseudocommand add16 src1 : src2 : dst {
    clc
    lda src1
    adc src2
    sta dst
    lda src1+1
    adc src2+1
    sta dst+1
}

.pseudocommand inc16 addr {
    inc addr
    bne skip
    inc addr+1
skip:
}

// ----------------------------------------------------------------------------
// 10. DATA DIRECTIVES
// ----------------------------------------------------------------------------
.pc = $1000 "Data Section"

text:
    .text "HELLO WORLD!"          // ✅ Text directive
    .byte 0                        // ✅ Single byte

numbers:
    .byte $00, $01, $02, $03       // ✅ Multiple bytes
    .word $1234, $5678             // ✅ Word values

fillData:
    .fill 256, $00                 // ✅ Fill with zeros
    .fill 16, i                    // ✅ Fill with counter

// ----------------------------------------------------------------------------
// 11. MAIN PROGRAM
// ----------------------------------------------------------------------------
.pc = $0801 "BASIC Upstart"

BasicUpstart(start)

.pc = $0810 "Main Program"

start:
    // Test macro calls
    clearScreen()                   // ✅ Correct: 0 arguments
    setColor(Colors.BLUE)          // ✅ Correct: 1 argument
    setBorderAndBackground(Colors.CYAN, Colors.BLACK)  // ✅ Correct: 2 arguments

    // Invalid macro calls (commented to avoid errors)
    // setColor(1, 2)               // ⚠️ Should warn: Too many arguments
    // setBorderAndBackground(1)    // ⚠️ Should warn: Too few arguments

    // Test function calls
    .var sum = add(5, 3)            // ✅ Function call: 5 + 3 = 8
    .var product = multiply(4, 7)   // ✅ Function call: 4 * 7 = 28
    .var squared = square(9)        // ✅ Function call: 9 * 9 = 81

    // Test pseudocommand calls
    mov #Colors.RED : BORDER        // ✅ Correct: 2 arguments
    add16 $fb : $fd : $c0          // ✅ Correct: 3 arguments
    inc16 counter                   // ✅ Correct: 1 argument

    // Invalid pseudocommand calls (commented)
    // mov #$05                     // ⚠️ Should warn: Too few arguments
    // add16 $fb : $fd             // ⚠️ Should warn: Too few arguments

    // Test namespace member access
    jsr graphics.clear              // ✅ Namespace member
    jsr sound.init                  // ✅ Namespace member

    // Test enum values
    lda #Colors.GREEN
    sta Registers.VIC_BORDER
    lda #Colors.YELLOW
    sta Registers.VIC_BACKGROUND

// ----------------------------------------------------------------------------
// 12. ALL STANDARD MNEMONICS (for completion/hover testing)
// ----------------------------------------------------------------------------
instructions:
    // Load/Store
    lda #$00                        // ✅ Load Accumulator
    ldx #$00                        // ✅ Load X
    ldy #$00                        // ✅ Load Y
    sta $1000                       // ✅ Store Accumulator
    stx $1001                       // ✅ Store X
    sty $1002                       // ✅ Store Y

    // Transfer
    tax                             // ✅ Transfer A to X
    tay                             // ✅ Transfer A to Y
    txa                             // ✅ Transfer X to A
    tya                             // ✅ Transfer Y to A
    tsx                             // ✅ Transfer Stack to X
    txs                             // ✅ Transfer X to Stack

    // Stack operations
    pha                             // ✅ Push Accumulator
    php                             // ✅ Push Processor Status
    pla                             // ✅ Pull Accumulator
    plp                             // ✅ Pull Processor Status

    // Arithmetic
    adc #$01                        // ✅ Add with Carry
    sbc #$01                        // ✅ Subtract with Carry
    inc $1000                       // ✅ Increment Memory
    inx                             // ✅ Increment X
    iny                             // ✅ Increment Y
    dec $1000                       // ✅ Decrement Memory
    dex                             // ✅ Decrement X
    dey                             // ✅ Decrement Y

    // Logical
    and #$0f                        // ✅ AND
    ora #$80                        // ✅ OR
    eor #$ff                        // ✅ Exclusive OR
    bit $1000                       // ✅ Bit Test

    // Shift/Rotate
    asl                             // ✅ Arithmetic Shift Left (Accumulator)
    asl $1000                       // ✅ Arithmetic Shift Left (Memory)
    lsr                             // ✅ Logical Shift Right (Accumulator)
    lsr $1000                       // ✅ Logical Shift Right (Memory)
    rol                             // ✅ Rotate Left (Accumulator)
    rol $1000                       // ✅ Rotate Left (Memory)
    ror                             // ✅ Rotate Right (Accumulator)
    ror $1000                       // ✅ Rotate Right (Memory)

    // Compare
    cmp #$00                        // ✅ Compare Accumulator
    cpx #$00                        // ✅ Compare X
    cpy #$00                        // ✅ Compare Y

    // Branch
    bcc *+2                         // ✅ Branch if Carry Clear
    bcs *+2                         // ✅ Branch if Carry Set
    beq *+2                         // ✅ Branch if Equal
    bne *+2                         // ✅ Branch if Not Equal
    bmi *+2                         // ✅ Branch if Minus
    bpl *+2                         // ✅ Branch if Plus
    bvc *+2                         // ✅ Branch if Overflow Clear
    bvs *+2                         // ✅ Branch if Overflow Set

    // Jump/Subroutine
    jmp loop                        // ✅ Jump
    jsr subroutine                  // ✅ Jump to Subroutine
    rts                             // ✅ Return from Subroutine

    // Flags
    clc                             // ✅ Clear Carry
    sec                             // ✅ Set Carry
    cli                             // ✅ Clear Interrupt Disable
    sei                             // ✅ Set Interrupt Disable
    clv                             // ✅ Clear Overflow
    cld                             // ✅ Clear Decimal Mode
    sed                             // ✅ Set Decimal Mode

    // System
    brk                             // ✅ Break
    nop                             // ✅ No Operation
    rti                             // ✅ Return from Interrupt

// ----------------------------------------------------------------------------
// 13. ALL ADDRESSING MODES (for parser testing)
// ----------------------------------------------------------------------------
addressingModes:
    lda #$42                        // ✅ Immediate
    lda $42                         // ✅ Zero Page
    lda $42,x                       // ✅ Zero Page, X
    ldy $42,x                       // ✅ Zero Page, X (with LDY)
    ldx $42,y                       // ✅ Zero Page, Y
    lda $1234                       // ✅ Absolute
    lda $1234,x                     // ✅ Absolute, X
    lda $1234,y                     // ✅ Absolute, Y
    lda ($42,x)                     // ✅ Indexed Indirect (X)
    lda ($42),y                     // ✅ Indirect Indexed (Y)
    jmp ($1234)                     // ✅ Indirect (JMP only)
    asl                             // ✅ Accumulator
    nop                             // ✅ Implied

// ----------------------------------------------------------------------------
// 14. NUMBER LITERALS (for lexer testing)
// ----------------------------------------------------------------------------
numberFormats:
    .byte $ff                       // ✅ Hexadecimal
    .byte 255                       // ✅ Decimal
    .byte %11111111                 // ✅ Binary
    .byte 'A'                       // ✅ Character literal
    .word $1234                     // ✅ Hex word
    .word 4660                      // ✅ Decimal word
    .word %0001001000110100         // ✅ Binary word

// ----------------------------------------------------------------------------
// 15. EXPRESSIONS (for expression evaluation)
// ----------------------------------------------------------------------------
expressions:
    .byte 5 + 3                     // ✅ Addition
    .byte 10 - 4                    // ✅ Subtraction
    .byte 6 * 7                     // ✅ Multiplication
    .byte 20 / 4                    // ✅ Division
    .byte 17 % 5                    // ✅ Modulo
    .byte $ff & $0f                 // ✅ Bitwise AND
    .byte $f0 | $0f                 // ✅ Bitwise OR
    .byte $ff ^ $aa                 // ✅ Bitwise XOR
    .byte $01 << 4                  // ✅ Left Shift
    .byte $80 >> 2                  // ✅ Right Shift
    .byte <$1234                    // ✅ Low Byte
    .byte >$1234                    // ✅ High Byte
    .byte (5 + 3) * 2               // ✅ Parentheses
    .byte -42                       // ✅ Negation

// ----------------------------------------------------------------------------
// 16. LABELS AND REFERENCES (for symbol resolution)
// ----------------------------------------------------------------------------
loop:
    ldx #10
innerLoop:
    dex
    bne innerLoop
    rts

subroutine:
    nop
    rts

.label localLabel = $2000
.label anotherLabel = localLabel + $100

// ----------------------------------------------------------------------------
// 17. PROGRAM COUNTER EXPRESSIONS (for PC handling)
// ----------------------------------------------------------------------------
pcTest:
    .pc = * + $100 "Skip 256 bytes"

dataAfterSkip:
    .byte $01, $02, $03

    .pc = $2000 "Fixed address"

fixedLocation:
    nop

    .pc = * + 50 "Skip 50 bytes"

// ----------------------------------------------------------------------------
// 18. ILLEGAL OPCODES (should warn)
// ----------------------------------------------------------------------------
illegalOpcodes:
    // slo $1000                    // ⚠️ Illegal opcode (should warn)
    // rla $1000                    // ⚠️ Illegal opcode (should warn)
    // sre $1000                    // ⚠️ Illegal opcode (should warn)
    // rra $1000                    // ⚠️ Illegal opcode (should warn)

// ----------------------------------------------------------------------------
// 19. COMMENTS (all styles)
// ----------------------------------------------------------------------------
// Single line comment
/* Block comment
   spanning multiple
   lines */

; Semicolon comment (assembly style)

nop  // Inline comment

// ----------------------------------------------------------------------------
// 20. C64 MEMORY-MAPPED REGISTERS (for hover info)
// ----------------------------------------------------------------------------
c64Registers:
    // VIC-II Chip
    lda $d000                       // Sprite 0 X
    lda $d001                       // Sprite 0 Y
    lda $d020                       // Border Color
    lda $d021                       // Background Color

    // SID Chip
    lda $d400                       // Voice 1 Frequency Low
    lda $d401                       // Voice 1 Frequency High
    lda $d418                       // Volume and Filter

    // CIA Chips
    lda $dc00                       // CIA1 Data Port A
    lda $dc01                       // CIA1 Data Port B
    lda $dd00                       // CIA2 Data Port A
    lda $dd01                       // CIA2 Data Port B

// ----------------------------------------------------------------------------
// TEST SUMMARY
// ----------------------------------------------------------------------------
// This file tests:
// ✅ All v0.9.7 directives (.encoding, .define, .import, .function, .macro,
//    .pseudocommand, .namespace, .enum)
// ✅ All directive validations (redefinition, missing .return, invalid encoding,
//    argument count validation)
// ✅ All standard 6502/6510 mnemonics
// ✅ All addressing modes
// ✅ Number formats (hex, decimal, binary, char)
// ✅ Expression evaluation
// ✅ Symbol resolution and references
// ✅ Program counter handling
// ✅ Comments (all styles)
// ✅ C64 memory-mapped registers
// ✅ Illegal opcode warnings
//
// HOW TO TEST IN NEOVIM:
// 1. Open this file: nvim comprehensive-server-test.asm
// 2. Check for diagnostics (no errors, only expected warnings)
// 3. Test completion: Type a partial mnemonic and trigger completion
// 4. Test hover: Hover over mnemonics, registers, symbols
// 5. Test goto definition: Jump to label/symbol definitions
// 6. Test find references: Find all references to a symbol
// 7. Uncomment lines marked with ⚠️ to see warnings
// ============================================================================
