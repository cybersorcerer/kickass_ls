## Phase 1: Fundament - Die semantische Analyse-Engine

  Das Kernstück der neuen Funktionen wird eine zentrale Analyse-Komponente sein. Diese Komponente wird nach dem Parsen des Codes ausgeführt und durchläuft
  den erzeugten Syntaxbaum (AST). Ihre Aufgabe ist es, den Code auf logische und kontextbezogene Korrektheit zu prüfen. Anstatt Fehler an verschiedenen
  Stellen im Code zu suchen, schaffen wir damit eine einzige, maßgebliche Instanz für die Fehleranalyse.

### Aktivitäten:

   1. Erstellen einer neuen analyze.go-Datei, die die Logik der Analyse-Engine enthalten wird.
   2. Anpassen des Haupt-Handlers (textDocument/didChange und didOpen), sodass die Analyse-Engine nach dem Parsen aufgerufen wird.
   3. Die Engine wird eine Liste von semantischen Fehlern zurückgeben, die dann von publishDiagnostics an den Editor gesendet werden.

  ---

## Phase 2: Symbol-Validierung (Doppelte & undefinierte Symbole)

  In dieser Phase konzentrieren wir uns auf die korrekte Definition und Verwendung von Symbolen (Labels, Konstanten, Variablen).

### Aktivitäten:

   1. Doppelte Symbole: Die Information über doppelte Symbole ist bereits im Symbol-Manager vorhanden, wird aber bisher nur als Warnung im Log ausgegeben. Ich werde dies so ändern, dass stattdessen ein klarer Fehler im Editor angezeigt wird.
   2. Undefinierte Symbole: Die Analyse-Engine wird jede Anweisung im Code durchgehen. Bei jeder Verwendung eines Symbols wird sie prüfen, ob dieses im aktuellen oder einem übergeordneten Gültigkeitsbereich (Scope) definiert ist. Falls nicht, wird ein "Undefined Symbol"-Fehler generiert.

  ---

## Phase 3: Validierung der Adressierungsmodi

  Diese Phase bringt ein tieferes Verständnis für die 6502-Befehle mit sich.

### Aktivitäten:

   1. Operand-Analyse: Die Analyse-Engine lernt, die Struktur von Operanden zu erkennen und sie einem spezifischen Adressierungsmodus zuzuordnen (z.B. Immediate, Absolute,X, Indirect,Y).
   2. Abgleich mit Mnemonic-Daten: Für jeden Befehl wird der erkannte Adressierungsmodus mit den gültigen Modi aus der mnemonic.json-Datei abgeglichen. Ungültige Kombinationen (wie z.B. lda #$12,x) werden als Fehler gemeldet.

  ---

## Phase 4: Warnungen für unbenutzte Symbole

  Um die Codequalität zu verbessern, führen wir eine Prüfung auf "toten Code" ein.

### Aktivitäten:

   1. Nutzungsverfolgung: Während die Analyse-Engine den Code prüft, markiert sie alle verwendeten Symbole.
   2. Abschließender Vergleich: Nach der Analyse des gesamten Dokuments vergleicht die Engine die Liste aller definierten Symbole mit der Liste der genutzten Symbole.
   3. Warnung generieren: Für jedes definierte, aber nie verwendete Symbol wird eine Warnung im Editor angezeigt. Diese Funktion wird über die bereits vorhandene Einstellung warnUnusedLabelsEnabled konfigurierbar sein.
