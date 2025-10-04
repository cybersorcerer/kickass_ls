// Test Case 05: C64 Memory Map References
// Purpose: Test recognition of C64 memory-mapped registers

* = $0810

// VIC-II Registers ($D000-$D3FF)
    lda $d000       // Sprite 0 X
    sta $d001       // Sprite 0 Y
    lda $d010       // Sprites 0-7 MSB of X coordinate
    sta $d015       // Sprite enable
    lda $d016       // Control register 2
    sta $d018       // Memory pointers
    lda $d020       // Border color
    sta $d021       // Background color 0
    lda $d027       // Sprite 0 color
    sta $d028       // Sprite 1 color

// SID Registers ($D400-$D7FF)
    lda $d400       // Voice 1 frequency low
    sta $d401       // Voice 1 frequency high
    lda $d402       // Voice 1 pulse waveform width low
    sta $d403       // Voice 1 pulse waveform width high
    lda $d404       // Voice 1 control register
    sta $d405       // Voice 1 attack/decay
    lda $d406       // Voice 1 sustain/release
    lda $d418       // Volume and filter modes

// CIA #1 Registers ($DC00-$DCFF)
    lda $dc00       // Data port A
    sta $dc01       // Data port B
    lda $dc02       // Data direction port A
    sta $dc03       // Data direction port B
    lda $dc04       // Timer A low byte
    sta $dc05       // Timer A high byte
    lda $dc0d       // Interrupt control register
    sta $dc0e       // Control register A

// CIA #2 Registers ($DD00-$DDFF)
    lda $dd00       // Data port A
    sta $dd01       // Data port B
    lda $dd02       // Data direction port A
    sta $dd03       // Data direction port B

// Color RAM ($D800-$DBFF)
    lda $d800       // Color RAM start
    sta $d801
    lda $dbe7       // Color RAM

// Screen RAM (default location)
    lda $0400       // Screen RAM start
    sta $0401
    lda $07e7       // Screen RAM end

// Zero Page
    lda $00
    sta $01         // 6510 I/O port
    lda $02
    sta $fb
    lda $fc
    sta $fd
    lda $fe

// Kernal vectors
    lda $0314       // IRQ vector low
    sta $0315       // IRQ vector high
    lda $fffa       // NMI vector low
    sta $fffb       // NMI vector high
    lda $fffc       // RESET vector low
    sta $fffd       // RESET vector high

// ROM areas
    jsr $ffd2       // CHROUT - output character to screen
    jsr $ffe4       // GETIN - get character from keyboard
    jsr $e544       // Clear screen
