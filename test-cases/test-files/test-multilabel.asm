// Test file for Multi-Label feature
// Multi labels can be declared multiple times with !label:
// Referenced with !label+ (forward) or !label- (backward)

* = $0801
BasicUpstart(start)

* = $1000
start:
    ldx #0

// Test 1: Simple loop with multi-labels
!loop:
    lda data,x
    sta $0400,x
    inx
    cpx #10
    bne !loop-      // Branch backward to previous !loop:

// Test 2: Nested loops with same multi-label name
    ldx #0
!loop:
    ldy #0
!loop:
    lda #$20
    sta $0400,x
    iny
    cpy #5
    bne !loop-      // Inner loop backward
    inx
    cpx #10
    bne !loop-      // Outer loop backward

// Test 3: Forward references
    ldx #10
    cpx #0
    beq !skip+      // Skip forward to next !skip:
    dex
    jmp *-3
!skip:
    nop

// Test 4: Multiple forward skips
    lda #0
    beq !skip+
    lda #1
!skip:
    sta $d020
    lda #0
    beq !skip+
    lda #2
!skip:
    sta $d021

    rts

// Test 5: Branch distance errors - too far for normal label
far_label:
    .fill 200, 0  // 200 bytes of data
    nop
    // This should cause a branch distance error (>127 bytes)
    beq far_label  // ERROR: Branch distance out of range

// Test 6: Multi-label branch distance error - backward too far
!far:
    .fill 150, 0  // 150 bytes after first !far:
    nop
    nop
    beq !far-  // ERROR: Branch distance out of range (>150 bytes back to first !far:)

// Test 7: Forward multi-label that's too far
    nop
    beq !far_forward+  // ERROR: Branch distance out of range
    .fill 200, 0
!far_forward:
    nop

data:
    .byte 1,2,3,4,5,6,7,8,9,10

.macro BasicUpstart(address) {
    * = $0801 "Basic"
    .byte $0b, $08, $0a, $00, $9e
    .text toIntString(address)
    .byte $00, $00, $00
}
