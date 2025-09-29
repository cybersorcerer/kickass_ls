// Test file for dynamic directive loading
.const MAX_VALUE = 255
.var counter = 0

.macro SetBorder(color) {
    lda #color
    sta $d020
}

.function Add(a, b) {
    .return a + b
}

.namespace Graphics {
    .const BORDER = $d020
}

#import "library.asm"

.if (counter < 10) {
    .print "Counter is small"
}

.byte 1, 2, 3, 4
.text "Hello World"
.fill 10, 0

start:
    :SetBorder(1)
    lda #Add(5, 3)
    rts