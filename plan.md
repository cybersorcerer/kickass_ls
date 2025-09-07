# Plan: Erweiterung für Kick Assembler-Syntax

  Phase 1: Fundamentale Symbol- und Scope-Erkennung

  Ziel: Dem Language Server beibringen, die grundlegenden und wichtigsten
  Syntax-Erweiterungen von Kick Assembler zu verstehen, insbesondere wie Symbole
  und Geltungsbereiche (Scopes) definiert werden.

   1. Parsing von Direktiven: Die Kern-Parsing-Logik wird erweitert, um die
      wichtigsten Kick Assembler-Direktiven zur Symboldefinition zu erkennen:
       * .const für Konstanten.
       * .var für Variablen.
       * .label und die Kurzform *= für Adress-Labels.
       * .namespace und { ... }-Blöcke für Namensräume und Scopes.
       * .function und .macro für wiederverwendbare Codeblöcke.

   2. Erweitertes Symbol-Modell: Intern wird die Art, wie Symbole gespeichert
      werden, erweitert. Statt nur "Label" zu kennen, wird ein Symbol nun einen
      "Typ" haben (z.B. Konstante, Variable, Funktion, Namespace). Dies ist die
      Grundlage für alle weiteren Verbesserungen.

   3. Hierarchisches Scope-Management: Die Logik wird so angepasst, dass sie mit
      verschachtelten Scopes (z.B. einem Namespace in einer Datei) umgehen kann.
      Wenn der Code ein Symbol wie gfx.border_color verwendet, muss der Server
      verstehen, dass border_color im gfx-Namespace gesucht werden muss.

  Phase 2: Verbesserung der bestehenden LSP-Funktionen

  Ziel: Die in Phase 1 gewonnenen Informationen nutzen, um die bereits
  implementierten IDE-Funktionen (Completion, Hover, etc.) "intelligenter" zu
  machen.

   1. Kontextsensitive Code-Vervollständigung (`completion`):
       * Die Auto-Vervollständigung schlägt nun nicht mehr nur Labels, sondern auch
         Variablen, Konstanten, Funktionen und Namespaces vor.
       * Anhand des Symbol-Typs aus Phase 1 werden passende Icons angezeigt (z.B.
         ein anderes Symbol für eine Funktion als für eine Variable).
       * Die Vorschläge sind Scope-sensitiv. Innerhalb eines gfx-Namespaces werden
         gfx-Symbole priorisiert.

   2. Detailreichere Hover-Informationen (`hover`):
       * Beim Überfahren einer Konstante (.const) wird deren Wert angezeigt.
       * Beim Überfahren einer Variable (.var) wird deren Typ angezeigt.
       * Beim Überfahren einer Funktion (.function) oder eines Makros (.macro) wird
         deren Signatur (erwartete Parameter) angezeigt.

   3. Präzises "Go to Definition":
       * Die Funktion wird auf alle neuen Symbol-Typen erweitert. Man kann dann
         direkt zur Definition einer Konstante, Variable oder Funktion springen.
       * Die Navigation zu Symbolen innerhalb von Namespaces (z.B. zu border_color
         in JMP gfx.border_color) wird korrekt aufgelöst.

  Phase 3: Assembler-spezifische Fehlerdiagnose

  Ziel: Den Server in die Lage versetzen, typische Fehler zu erkennen, die
  spezifisch für die Kick Assembler-Syntax sind.

   1. Neue Diagnose-Regeln (`diagnostics`):
       * Warnung bei der Verwendung einer Variable, die im aktuellen Scope nicht
         sichtbar ist.
       * Fehler bei falscher Anzahl von Argumenten beim Aufruf einer Funktion oder
         eines Makros.
       * Fehler bei Typ-Inkonsistenzen (z.B. wenn eine Zahl an eine Funktion
         übergeben wird, die einen String erwartet).

   2. Unterstützung für Multi-Datei-Projekte (`#import`):
       * Als letzten Schritt wird eine grundlegende Unterstützung für die
         #import-Direktive implementiert. Der Server liest importierte Dateien ein,
         um deren Symbole für die Code-Vervollständigung und Definitionssuche im
         gesamten Projekt verfügbar zu machen.
