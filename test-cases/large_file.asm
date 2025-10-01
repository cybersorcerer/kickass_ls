; Large file for performance testing
* = $1000
.const SCREEN = $0400
.const COLOR = $D800
.var counter = 0

main:
    ldx #$00
loop1:
    lda #$01
    sta SCREEN,x
    lda #$0E
    sta COLOR,x
    inx
    cpx #$FF
    bne loop1
    
    ldy #$00
loop2:
    lda #$02
    sta SCREEN+$100,y
    lda #$06
    sta COLOR+$100,y
    iny
    cpy #$FF
    bne loop2
    
    ; More complex code patterns
.for (var i=0; i<10; i++) {
    lda #i
    sta $1000+i
}

.macro fill_memory(start, end, value)
    ldx #$00
fill_loop:
    lda #value
    sta start,x
    inx
    cpx #(end-start)
    bne fill_loop
.endmacro

    fill_memory($2000, $2100, $FF)
    
    rts