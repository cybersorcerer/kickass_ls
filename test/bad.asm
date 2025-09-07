LDA             ; Fehler: Operand fehlt (Implied nicht erlaubt für LDA)
LDA ($10)       ; Fehler: LDA unterstützt kein Indirect (nur ($10,X) und ($10),Y)
JMP #$10        ; Fehler: JMP unterstützt kein Immediate
STA #$20        ; Fehler: STA unterstützt kein Immediate
JSR $10         ; Fehler: JSR unterstützt kein Zeropage (nur Absolute)
