# Refactoring-Vorschlag für den Language Server

  Aktueller Zustand:
  Derzeit erledigt server.go fast alles: LSP-Nachrichten verarbeiten, Dokumente speichern, Symbole parsen, Hover-Informationen liefern, Definitionen finden, Referenzen suchen,
  Vervollständigungen anbieten und Diagnosen erstellen. Das führt zu einem "Monolithen" innerhalb einer Datei.

  Zielstruktur:
  Wir würden die Funktionalität in folgende, neue Pakete (oder separate Dateien innerhalb des lsp-Pakets, falls die Struktur nicht zu tief werden soll) aufteilen:

1 `lsp` (bestehend): Behält die Kernlogik für den Empfang und das Versenden von LSP-Nachrichten und die Weiterleitung an die spezialisierten Handler. Die Start()-Funktion bleibt hier.

2 `document` (neu): Verwaltet das Speichern und Abrufen der geöffneten Textdokumente.

3 `parser` (neu): Dies ist der wichtigste Teil. Dieses Paket würde die gesamte Logik für das Parsen des Kick Assembler-Codes enthalten. Es würde eine strukturierte Repräsentation des

Codes erstellen, einschliesslich einer Symboltabelle, die vollständig Namespace- und Scope-fähig ist.

       - Funktionen wie getWordAtPosition, isWordChar, normalizeLabel würden hierher verschoben.
       - Eine zentrale Parse()-Funktion würde den Text analysieren und die Symboltabelle aufbauen.

4 `symbols` (neu): Dieses Paket würde die Definitionen für die Symbolstrukturen (SymbolInfo, etc.) und Methoden zur Abfrage der Symboltabelle (z.B. GetDefinition, FindReferences,GetHoverInfo) enthalten. Die Logik, die derzeit in den einzelnen LSP-Handlern für die Symbolsuche dupliziert ist, würde hier zentralisiert.

5 `diagnostics` (neu): Enthält die gesamte Logik zur Analyse des geparsten Codes und zur Erstellung von Diagnosemeldungen (Fehler, Warnungen). Die publishDiagnostics()-Funktion würde hierher verschoben.
6 `mnemonics` (neu): Verwaltet das Laden und Bereitstellen der Mnemonik-Daten (Opcodes und deren Beschreibungen).

## Vorteile dieses Refactorings

- Übersichtlichkeit: server.go wird deutlich kleiner und fokussiert sich nur noch auf die LSP-Kommunikation.
- Modularität: Jedes Paket hat eine klare, abgegrenzte Aufgabe. Das macht den Code leichter verständlich und wartbar.
- Testbarkeit: Einzelne Komponenten (z.B. der Parser oder die Diagnoselogik) können unabhängig voneinander getestet werden.
- Wartbarkeit: Änderungen in einem Bereich (z.B. eine neue Syntaxregel) wirken sich weniger auf andere Bereiche aus.
- Erweiterbarkeit: Das Hinzufügen neuer LSP-Features oder Kick Assembler-Syntax-Elemente wird einfacher.
- Behebung der Kommunikationsprobleme: Eine klarere Struktur und weniger komplexe Funktionen könnten auch die Stabilität verbessern.

Wie das Namespace-Handling hineinpasst:
Die Implementierung des .namespace-Handlings, an der wir zuletzt gearbeitet haben, würde vollständig in das neue parser-Paket integriert.Der Parser wäre dann dafür verantwortlich, die
Namespaces korrekt zu erkennen und die Symbole entsprechend zu qualifizieren. Die LSP-Handler würden dann einfach die vom Parserbereitgestellte, bereits Namespace-bewusste Symboltabelle
abfragen.
