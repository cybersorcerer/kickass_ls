# Semantische Analyse

## Phase 1: Fundament - Die semantische Analyse-Engine

  Das Kernstück der neuen Funktionen wird eine zentrale Analyse-Komponente sein. Diese Komponente wird nach dem Parsen des Codes ausgeführt und durchläuft
  den erzeugten Syntaxbaum (AST). Ihre Aufgabe ist es, den Code auf logische und kontextbezogene Korrektheit zu prüfen. Anstatt Fehler an verschiedenen
  Stellen im Code zu suchen, schaffen wir damit eine einzige, maßgebliche Instanz für die Fehleranalyse.

**Aktivitäten:**

   1. Erstellen einer neuen analyze.go-Datei, die die Logik der Analyse-Engine enthalten wird.
   2. Anpassen des Haupt-Handlers (textDocument/didChange und didOpen), sodass die Analyse-Engine nach dem Parsen aufgerufen wird.
   3. Die Engine wird eine Liste von semantischen Fehlern zurückgeben, die dann von publishDiagnostics an den Editor gesendet werden.

  ---

## Phase 2: Symbol-Validierung (Doppelte & undefinierte Symbole)

  In dieser Phase konzentrieren wir uns auf die korrekte Definition und Verwendung von Symbolen (Labels, Konstanten, Variablen).

**Aktivitäten:**

   1. Doppelte Symbole: Die Information über doppelte Symbole ist bereits im Symbol-Manager vorhanden, wird aber bisher nur als Warnung im Log ausgegeben. Ich werde dies so ändern, dass stattdessen ein klarer Fehler im Editor angezeigt wird.
   2. Undefinierte Symbole: Die Analyse-Engine wird jede Anweisung im Code durchgehen. Bei jeder Verwendung eines Symbols wird sie prüfen, ob dieses im aktuellen oder einem übergeordneten Gültigkeitsbereich (Scope) definiert ist. Falls nicht, wird ein "Undefined Symbol"-Fehler generiert.

  ---

## Phase 3: Validierung der Adressierungsmodi

  Diese Phase bringt ein tieferes Verständnis für die 6502-Befehle mit sich.

**Aktivitäten:**

   1. Operand-Analyse: Die Analyse-Engine lernt, die Struktur von Operanden zu erkennen und sie einem spezifischen Adressierungsmodus zuzuordnen (z.B. Immediate, Absolute,X, Indirect,Y).
   2. Abgleich mit Mnemonic-Daten: Für jeden Befehl wird der erkannte Adressierungsmodus mit den gültigen Modi aus der mnemonic.json-Datei abgeglichen. Ungültige Kombinationen (wie z.B. lda #$12,x) werden als Fehler gemeldet.

  ---

## Phase 4: Warnungen für unbenutzte Symbole

  Um die Codequalität zu verbessern, führen wir eine Prüfung auf "toten Code" ein.

**Aktivitäten:**

   1. Nutzungsverfolgung: Während die Analyse-Engine den Code prüft, markiert sie alle verwendeten Symbole.
   2. Abschließender Vergleich: Nach der Analyse des gesamten Dokuments vergleicht die Engine die Liste aller definierten Symbole mit der Liste der genutzten Symbole.
   3. Warnung generieren: Für jedes definierte, aber nie verwendete Symbol wird eine Warnung im Editor angezeigt. Diese Funktion wird über die bereits vorhandene Einstellung warnUnusedLabelsEnabled konfigurierbar sein.

I have read the content of internal/lsp/analyze.go.

  Existing Analyses:
   * Unused Symbols: There is a checkForUnusedSymbols function that checks Label, Constant, and Variable for unused occurrences.
   * Argument Count for Functions/Macros/PseudoCommands: The walkExpression function checks the number of arguments for CallExpressions if the symbol type is Macro, Function, or
     PseudoCommand.
   * Invalid Addressing Modes: validateInstruction in parser.go (not analyze.go) checks this.
   * Missing Expressions after `=`: parser.go checks this.

  Suggestions for improving semantic analysis:

   1. More Comprehensive Addressing Mode Validation:
       * The current determineAddressingModeFromAST in parser.go is very rudimentary (Absolute, Immediate, Implied). It should be extended to correctly recognize all 6510 addressing
         modes (e.g., ZeroPage, Absolute,X, Indirect,Y, etc.).
       * The isAddressingModeValid function should then perform more detailed checks to ensure that the operand actually fits the recognized addressing mode (e.g., whether a value
         for ZeroPage is actually 0-$FF).

   2. Range Checks for Values:
       * 8-bit values: For instructions expecting 8-bit operands (e.g., LDA #$FF), it should be checked whether the value is within the range $00-$FF.
       * 16-bit values: For addresses or 16-bit operands, it should be checked whether the value is within the range $0000-$FFFF.
       * Relative Jumps: For instructions like BCC, BNE, etc., it should be checked whether the jump target is within the valid relative range (-128 to +127 bytes). This requires a
         basic code layout analysis to calculate the relative distance.

   3. Type Checking and Inference:
       * String Literals: Warnings if a string literal is used in a context expecting a numeric value, or vice versa.
       * Boolean Expressions: Warnings if a non-boolean value is used in a context expecting a boolean value (e.g., in .if statements).

   4. Dead Code / Unreachable Code Detection:
       * After an RTS, RTI, JMP, or BRK, subsequent code should be marked as unreachable unless there is another jump target to it. This is a common source of errors or unnecessary
         code.

   5. Warnings for Inefficient/Redundant Code:
       * LDA #$00 followed by STA $XX could trigger a warning if the value $00 is already in the accumulator or if $XX already contains $00.
       * INC $XX followed by DEC $XX (or vice versa) without intervening use.

   6. Cross-File Symbol Resolution:
       * If the LSP manages multiple files in the workspace, symbols defined in other files (e.g., via #import) should be resolvable. This requires an extension of Scope management.

   7. Context-Sensitive Completion:
       * The getCompletionContext function could be made even smarter. For example, if the cursor is after a JMP, labels and functions should be suggested, but not constants. After
         LDA #, constants and numeric expressions should be suggested.

   8. Directive-Specific Validation:
       * .fill: Check if the number of fills and the value are valid.
       * .byte, .word, .dword: Check if the specified values fit within the respective byte range.

  Implementation Priority:

  I would suggest starting with the following points, as they have the greatest impact on code quality and error detection:

   1. More Comprehensive Addressing Mode Validation and Range Checks for Values (8-bit/16-bit): These are fundamental errors that commonly occur in assembler code.
   2. Dead Code Detection: Helps identify unnecessary code and improve readability.
   3. Directive-Specific Validation: Especially for .byte, .word, .dword, and .fill to ensure data is correct.

  What do you think of these suggestions, Ronny? Would you like me to start with one of these points?
