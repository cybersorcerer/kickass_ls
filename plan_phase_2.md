# Neuer Plan: Vervollständigung des 6510 Language Servers

Basierend auf der Analyse des aktuellen Zustands des Language Servers im Vergleich zum ursprünglichen Plan, wird hier ein aktualisierter Plan zur Implementierung der fehlenden Funktionen vorgestellt.

## Phase 1: Unterstützung für Funktionen und Makros

**Ziel:** Die grundlegende Unterstützung für die Direktiven `.function` und `.macro` von Kick Assembler implementieren.

1. **Parser-Erweiterung:**
    * Der Parser in `internal/lsp/parser.go` wird erweitert, um `.function`- und `.macro`-Definitionen zu erkennen.
    * Der Parser muss die Namen der Funktionen/Makros und deren Parameter (Namen und optional Typen) extrahieren.
    * Für jede Funktion/jedes Makro wird ein eigener `Scope` erstellt, um lokale Symbole (Parameter, lokale Labels) zu verwalten.

2. **Symbol-Erweiterung:**
    * Die `Symbol`-Struktur in `internal/lsp/symbol.go` wird möglicherweise erweitert, um Parameterinformationen für Funktionen und Makros zu speichern. Eine einfache Liste von Parameternamen pro `Symbol` sollte ausreichen.

## Phase 2: Integration von Funktionen und Makros in LSP-Features

**Ziel:** Die in Phase 1 hinzugefügten Informationen in den bestehenden LSP-Funktionen nutzbar machen.

1. **Code-Vervollständigung (`completion`):**
    * Funktions- und Makronamen werden zur Liste der globalen Vervollständigungsvorschläge hinzugefügt.
    * Beim Aufruf einer Funktion/eines Makros werden die Parameternamen als Snippets vorgeschlagen.

2. **Hover-Informationen (`hover`):**
    * Beim Überfahren eines Funktions- oder Makronamens wird dessen Signatur (z.B. `function myFunc(param1, param2)`) im Hover-Fenster angezeigt.

3. **"Go to Definition":**
    * Die "Go to Definition"-Funktionalität wird so erweitert, dass sie von einem Funktions- oder Makroaufruf zur entsprechenden Definition springen kann.

## Phase 3: Erweiterte Fehlerdiagnose

**Ziel:** Die diagnostischen Fähigkeiten des Servers verbessern, um komplexere Fehler zu finden.

1. **Scope-basierte Analyse:**
    * Implementierung einer Prüfung in `publishDiagnostics`, die warnt, wenn auf ein Symbol (Variable, Konstante, Label) zugegriffen wird, das im aktuellen Geltungsbereich nicht definiert ist. Die `FindSymbol`-Logik ist hierfür die Grundlage.

2. **Argumenten-Prüfung:**
    * Beim Aufruf von Funktionen und Makros wird die Anzahl der übergebenen Argumente mit der Anzahl der definierten Parameter verglichen. Bei Abweichungen wird ein Fehler gemeldet.

3. **(Optional) Typ-Prüfung:**
    * Wenn die Zeit es erlaubt, kann eine einfache Typ-Prüfung implementiert werden (z.B. sicherstellen, dass numerische Werte nicht an Parameter übergeben werden, die als String deklariert wurden).

## Phase 4: Multi-Datei-Unterstützung

**Ziel:** Dem Language Server beibringen, mit Projekten umzugehen, die aus mehreren Dateien bestehen.

1. **`#import`-Direktive:**
    * Der Parser wird erweitert, um `#import "dateiname.asm"` zu erkennen.
    * Wenn eine `#import`-Anweisung gefunden wird, wird die referenzierte Datei ebenfalls geparst.
    * Die Symbole aus der importierten Datei werden dem globalen Scope des importierenden Dokuments hinzugefügt, um sie für Vervollständigung, Hover und "Go to Definition" verfügbar zu machen.
    * Es muss ein Mechanismus implementiert werden, um zirkuläre Importe zu verhindern.
