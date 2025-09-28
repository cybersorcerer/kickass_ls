; ============================================================================
; 6510 Features Test File - Kick Assembler for C64
; Tests all newly implemented and fixed features for 6510 CPU
; ============================================================================

BasicFileType "C64 program"
.pc = $0801

; ============================================================================
; CRITICAL BUG FIXES TEST
; ============================================================================

; Test 1: TAX Mnemonic (was broken: "ta x" instead of "tax")
test_tax_fix:
    lda #$42        ; Load accumulator
    tax             ; Transfer A to X (this was broken before!)
    txa             ; Transfer X to A
    tay             ; Transfer A to Y
    tya             ; Transfer Y to A
    tsx             ; Transfer stack pointer to X
    txs             ; Transfer X to stack pointer

; Test 2: HEX Number Validation (was accepting invalid G-Z characters)
test_hex_validation:
    lda #$A0        ; Valid hex
    ldx #$FF        ; Valid hex
    ldy #$DEAD      ; 4-digit hex (maximum allowed)
    sta $D020       ; Valid hex address
    lda #$GGGG      ; This should now be flagged as invalid

; ============================================================================
; ANONYMOUS LABELS TEST
; ============================================================================

test_anonymous_labels:
    ldx #$00
!:
    lda data,x      ; Anonymous label
    beq !+          ; Forward reference to next !
    sta $0400,x
    inx
    bne !-          ; Backward reference to previous !
!:
    rts

; More complex anonymous labels
complex_loop:
    ldy #$00
!:
    ldx #$00
!:
    lda screen_data,x
    sta $0400,y
    iny
    inx
    cpx #$28
    bne !-          ; Inner loop back
    cpy #$E8
    bne !--         ; Outer loop back (double minus)
    rts

; ============================================================================
; BUILT-IN FUNCTIONS TEST
; ============================================================================

; String Functions
test_string_functions:
    .var hexString = toHexString(255)
    .var binString = toBinaryString(170)
    .var octString = toOctalString(64)
    .var intString = toIntString(1337)

    .print "Hex: " + hexString
    .print "Bin: " + binString
    .print "Oct: " + octString
    .print "Int: " + intString

; Math Functions
test_math_functions:
    .var pi_value = PI
    .var e_value = E
    .var sine_val = sin(pi_value/2)
    .var cosine_val = cos(0)
    .var sqrt_val = sqrt(16)
    .var max_val = max(10, 20)
    .var min_val = min(10, 20)
    .var abs_val = abs(-42)
    .var pow_val = pow(2, 8)
    .var random_val = random()

    .print "PI: " + toHexString(floor(pi_value * 1000))
    .print "E: " + toHexString(floor(e_value * 1000))
    .print "Sin(PI/2): " + toHexString(floor(sine_val * 1000))

; File Functions
test_file_functions:
    .var sprite_data = LoadBinary("sprite.dat")
    .var sid_data = LoadSid("music.sid")
    .var picture_data = LoadPicture("image.koa", BF_KOALA)

; 3D Functions
test_3d_functions:
    .var identity_matrix = Matrix(4,4)
    .var rotation_matrix = RotationMatrix(PI/4, 0, 0)
    .var scale_matrix = ScaleMatrix(2, 2, 1)
    .var translation = MoveMatrix(10, 20, 0)
    .var vector3d = Vector(1, 2, 3)

; ============================================================================
; COLOR CONSTANTS TEST (from TextMate grammar)
; ============================================================================

test_color_constants:
    lda #BLACK
    sta $D020
    lda #WHITE
    sta $D021
    lda #RED
    sta $D022
    lda #CYAN
    sta $D023
    lda #PURPLE
    sta $D024
    lda #GREEN
    ldx #BLUE
    ldy #YELLOW

    ; Additional colors
    lda #ORANGE
    ldx #BROWN
    ldy #LIGHT_RED
    lda #DARK_GREY
    ldx #GREY
    ldy #LIGHT_GREEN
    lda #LIGHT_BLUE
    ldx #LIGHT_GREY

; ============================================================================
; 6510 INSTRUCTION SET TEST
; ============================================================================

; Standard 6502/6510 Instructions
test_standard_opcodes:
    adc #$01        ; Add with carry
    and #$FF        ; Logical AND
    asl             ; Arithmetic shift left
    bit $00         ; Bit test
    clc             ; Clear carry
    cld             ; Clear decimal
    cli             ; Clear interrupt
    clv             ; Clear overflow
    cmp #$42        ; Compare accumulator
    cpx #$00        ; Compare X
    cpy #$00        ; Compare Y
    dec $00         ; Decrement memory
    dex             ; Decrement X
    dey             ; Decrement Y
    eor #$AA        ; Exclusive OR
    inc $00         ; Increment memory
    inx             ; Increment X
    iny             ; Increment Y
    lda #$00        ; Load accumulator
    ldx #$00        ; Load X
    ldy #$00        ; Load Y
    lsr             ; Logical shift right
    nop             ; No operation
    ora #$55        ; Logical OR
    pha             ; Push accumulator
    php             ; Push processor status
    pla             ; Pull accumulator
    plp             ; Pull processor status
    rol             ; Rotate left
    ror             ; Rotate right
    sbc #$01        ; Subtract with carry
    sec             ; Set carry
    sed             ; Set decimal
    sei             ; Set interrupt
    sta $00         ; Store accumulator
    stx $00         ; Store X
    sty $00         ; Store Y

; Control Flow Instructions
test_control_flow:
    bcc branch1     ; Branch if carry clear
    bcs branch2     ; Branch if carry set
    beq branch3     ; Branch if equal
    bmi branch4     ; Branch if minus
    bne branch5     ; Branch if not equal
    bpl branch6     ; Branch if plus
    bvc branch7     ; Branch if overflow clear
    bvs branch8     ; Branch if overflow set
    jmp main_loop   ; Jump
    jsr subroutine  ; Jump to subroutine
    rti             ; Return from interrupt
    rts             ; Return from subroutine

branch1:
branch2:
branch3:
branch4:
branch5:
branch6:
branch7:
branch8:
    nop

; Illegal Opcodes (6510-specific undocumented opcodes)
test_illegal_opcodes:
    slo $00         ; Shift left and OR
    rla $00         ; Rotate left and AND
    sre $00         ; Shift right and EOR
    rra $00         ; Rotate right and add
    sax $00         ; Store A AND X
    lax $00         ; Load A and X
    dcp $00         ; Decrement and compare
    isc $00         ; Increment and subtract
    anc #$00        ; AND with carry
    asr #$00        ; AND and shift right
    arr #$00        ; AND and rotate right
    sbx #$00        ; Subtract from X
    dop $00         ; Double NOP
    top $0000       ; Triple NOP
    jam             ; Halt processor

; ============================================================================
; KICK ASSEMBLER DIRECTIVES TEST
; ============================================================================

; Data Directives
test_data_directives:
    .byte $01, $02, $03
    .word $1234, $5678
    .dword $12345678
    .text "Hello World!"
    .fill 10, $00

; Variable and Constant Declarations
test_variables:
    .const SCREEN = $0400
    .var counter = 0
    .var max_loops = 256
    .label loop_start

; Preprocessor Directives
test_preprocessor:
    #define DEBUG 1
    #if DEBUG
        .print "Debug mode enabled"
    #else
        .print "Release mode"
    #endif

; Control Flow Directives
test_control_directives:
    .for (var i = 0; i < 10; i++) {
        .byte i
    }

    .if (counter > 0) {
        inc counter
    }

; ============================================================================
; DATA SECTION
; ============================================================================

data:
    .byte 1, 2, 3, 4, 5, 0

screen_data:
    .text "6510 FEATURES TEST COMPLETE!"
    .byte 0

subroutine:
    nop
    rts

main_loop:
    jmp main_loop

; ============================================================================
; TEST SUMMARY
; ============================================================================
; This file tests:
; ✅ Fixed TAX mnemonic (was "ta x")
; ✅ Fixed HEX validation (no more G-Z)
; ✅ Anonymous labels (!, !+, !-)
; ✅ Built-in string functions (toHexString, etc.)
; ✅ Built-in math functions (sin, cos, sqrt, etc.)
; ✅ Built-in file functions (LoadBinary, etc.)
; ✅ Built-in 3D functions (Matrix, Vector, etc.)
; ✅ Color constants (BLACK, WHITE, RED, etc.)
; ✅ All standard 6502/6510 opcodes
; ✅ All control flow opcodes
; ✅ Illegal opcodes (6510-specific undocumented)
; ✅ Kick Assembler directives
; ============================================================================