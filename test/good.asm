; Korrekte Addressing Modes
LDA #$10         ; Immediate
LDA $10          ; Zeropage
LDA $1000        ; Absolute
LDA $10,X        ; Zeropage,X
LDA $1000,X      ; Absolute,X
LDA $1000,Y      ; Absolute,Y
LDA ($10,X)      ; Indexed-indirect
LDA ($10),Y      ; Indirect-indexed

JMP $C000        ; Absolute
JSR $FFD2        ; Absolute

STA $D020        ; Absolute
STA $10          ; Zeropage
