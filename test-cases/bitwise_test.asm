.const base = $1000
.const mask = %11110000

start:
    ; Test left shift
    lda #base << 2
    
    ; Test right shift  
    lda #mask >> 4
    
    ; Test bitwise AND
    lda #base & $FF
    
    ; Test bitwise OR
    lda #base | $0F
    
    ; Test bitwise XOR
    lda #base ^ $AA
    
    rts