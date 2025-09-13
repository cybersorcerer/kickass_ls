// Kick Assembler Test-Datei für den 6510-LSP

// --- PHASE 1: Symbol- und Scope-Erkennung ---

; .const für Konstanten
.const MAX_SPRITES = 8
.const SCREEN_MEM = $0400
.const TEST = 1

; .var für Variablen
.var sprite_x_pos = 0

; Namespace für Grafik-Routinen
.namespace gfx {
    .const BORDER_COLOR_REG = $d020
    .const BG_COLOR_REG = $d021

    .label loop:
        inc BORDER_COLOR_REG
        jmp loop
}

; Lokale Labels (.label)
.label global_label:
.label .local_label:
    lda #.local_label
    lda %011101
    jmp global_label


; --- PHASE 2: LSP-Funktionstests ---

; A. Go-to-Definition & Hover
;    - Fahre mit der Maus über MAX_SPRITES, SCREEN_MEM, sprite_x_pos.
;    - Führe "Go to Definition" auf ihnen aus.
;    - Führe "Go to Definition" auf gfx.loop und .local_label aus.

start:
    lda #MAX_SPRITES
    ldx #<SCREEN_MEM
    ldy #>SCREEN_MEM
    sta sprite_x_pos
    jmp gfx.loop
    jmp .local_label


; B. Code-Vervollständigung
;    - Tippe nach "lda #" die Buchstaben "MAX" -> MAX_SPRITES sollte erscheinen.
;    - Tippe nach "jmp" die Buchstaben "gfx." -> gfx.loop sollte erscheinen.
;    - Tippe nach "jmp" die Buchstaben ".lo" -> .local_label sollte erscheinen.

    lda #M
    jmp gfx.


; --- PHASE 3: Fehlerdiagnose ---

; Fehler: Ungültiges Symbol (sollte unterstrichen werden)
    lda UNBEKANNTES_SYMBOL

; Fehler: Doppeltes Label (sollte unterstrichen werden)
end:
    nop
