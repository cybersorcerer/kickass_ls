.const base = $1000
.const unused_const = 42

start:
    lda #$01
    sta base
    
    ; This should generate undefined symbol error
    lda undefined_symbol
    
    ; This should work fine
    lda #base
    
    ; Jump to undefined label should error
    jmp unknown_label
    
    rts

; Unreferenced label
unreferenced_label:
    nop
    rts