// Regression Test 07: C64 Memory Map Recognition
// Purpose: Ensure all C64 memory addresses are recognized
// Status: Should PASS with 0 errors

* = $0801

start:
    // VIC-II Registers ($D000-$D3FF)
    lda #$00
    sta $d000      // Sprite 0 X
    sta $d001      // Sprite 0 Y
    sta $d020      // Border Color
    sta $d021      // Background Color
    sta $d011      // VIC Control Register
    sta $d016      // VIC Control Register 2
    sta $d018      // Memory Control

    // SID Registers ($D400-$D7FF)
    lda #$0f
    sta $d400      // Voice 1 Frequency Low
    sta $d401      // Voice 1 Frequency High
    sta $d404      // Voice 1 Control
    sta $d418      // Volume and Filter

    // CIA #1 Registers ($DC00-$DCFF)
    lda $dc00      // Data Port A (Keyboard)
    sta $dc01      // Data Port B
    lda $dc04      // Timer A Low
    sta $dc05      // Timer A High

    // CIA #2 Registers ($DD00-$DDFF)
    lda $dd00      // Data Port A (Serial/VIC Bank)
    sta $dd01      // Data Port B
    lda $dd04      // Timer A Low
    sta $dd05      // Timer A High

    // Color RAM ($D800-$DBFF)
    lda #$01
    sta $d800      // Color RAM start
    sta $d8ff      // Color RAM

    // Zero Page ($00-$FF)
    lda #$00
    sta $00        // Zero page
    sta $fb        // Common ZP location
    sta $fd        // Common ZP location

    // Kernal vectors
    lda $0314      // IRQ vector low
    sta $0315      // IRQ vector high

    rts
