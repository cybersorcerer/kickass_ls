START:  LDA #$01
        STA $0400
.LOOP:  INC $0400
        BNE .LOOP   ; Korrekt: Lokales Label im Scope
        JMP START   ; Korrekt: Globales Label

BAD:    JMP .NOTDEF ; Fehler: .NOTDEF nicht definiert
