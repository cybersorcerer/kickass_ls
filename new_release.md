🔧 Was könnte noch verbessert werden:
1. Performance
Dokument wird mehrfach geparst (einmal für Symbol-Tree, einmal für Semantic Tokens)
Verbesserung: Semantic Tokens könnten die bereits geparste Symbol-Tree wiederverwenden
2. Macro/Pseudocommand Parameter
Werden die Parameter in Aufrufen validiert? (Anzahl, Typen)
Verbesserung: Parameter-Validierung, Autocomplete für Parameter
3. Go-to-Definition
Funktioniert das für Macros, Pseudocommands, Labels?
Test: Spring zu Macro-Definition, Label-Definition
4. Hover-Informationen
Werden Macro-Parameter, Function-Signaturen gezeigt?
Verbesserung: Dokumentation aus Kommentaren extrahieren
5. Diagnostics
Gibt es gute Error-Messages für typische Fehler?
Verbesserung: Bessere Fehlermeldungen, Warnings für unused symbols
6. Code Actions
Z.B. "Extract to Macro", "Rename Label"
Verbesserung: Quick-fixes für typische Probleme
7. Completion
Funktioniert Autocomplete für Labels, Macros, Pseudocommands?
Verbesserung: Context-aware completion (z.B. nur Labels bei JMP)

Was ist mit Code Formatierung?
## Code Formatierung fehlt komplett! LSP unterstützt:
textDocument/formatting - Ganzes Dokument formatieren
textDocument/rangeFormatting - Nur ausgewählten Bereich formatieren
textDocument/onTypeFormatting - Automatisch beim Tippen (z.B. nach } oder Newline)
Für Kick Assembler würde das bedeuten:
Formatierungs-Regeln:
Einrückung
Labels: Keine Einrückung (Col 0)
Instructions: 4 Spaces
Directives: 0 oder 4 Spaces (je nach Kontext)
Block-Inhalte (.macro, .function, .if): +4 Spaces pro Level
Alignment
lda #$00        ; Load zero
sta $d020       ; Border
jmp loop        ; Jump
vs. nicht aligned:
lda #$00 ; Load zero
sta $d020 ; Border
jmp loop ; Jump
Spacing
Nach Mnemonics: 1 Space
Um Operatoren: Optional Spaces (a+b vs. a + b)
Nach Kommas: 1 Space
Case-Konventionen
Mnemonics: lowercase oder UPPERCASE?
Labels: camelCase, snake_case, UPPER_CASE?
Konstanten: UPPER_CASE?
Implementation:
Ich könnte einen Formatter bauen der:
AST durchläuft
Jeden Node mit korrekter Einrückung und Spacing neu schreibt
Konfigurierbar über .kickass-format.json