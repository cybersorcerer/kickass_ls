Zusammenfassung: Plan vs. Implementierung

  Phase 1: Fundamentale Symbol- und Scope-Erkennung
   * 1. Parsing von Direktiven:
       * 🟡 .const / .var: Teilweise implementiert. Werden für Hover und Definition erkannt.
       * 🟡 .function / .macro: Teilweise implementiert. Werden für Definition erkannt.
       * ❌ .label / *: Nicht implementiert. Nur Labels mit : am Ende werden erkannt.
       * ❌ .namespace / {...}: Nicht implementiert. Das Konzept von Namespaces fehlt noch.
   * 2. Erweitertes Symbol-Modell:
       * ✅ Erledigt. Der Server unterscheidet intern bereits zwischen verschiedenen Symbol-Typen (variable, function, macro, label).
   * 3. Hierarchisches Scope-Management:
       * 🟡 Teilweise implementiert. Lokale Labels (z.B. .loop) werden bereits korrekt einem globalen Label zugeordnet. Das übergeordnete Namespace-Konzept fehlt aber noch.

  Phase 2: Verbesserung der bestehenden LSP-Funktionen
   * 1. Code-Vervollständigung:
       * 🟡 Teilweise implementiert. Schlägt Opcodes und Labels vor, aber noch keine Konstanten, Variablen, Funktionen oder Namespaces.
   * 2. Hover-Informationen:
       * 🟡 Teilweise implementiert. Zeigt Werte für Konstanten/Variablen, aber noch keine Signaturen für Funktionen/Makros.
   * 3. "Go to Definition":
       * 🟡 Teilweise implementiert. Funktioniert für die meisten erkannten Symbole, aber noch nicht für Symbole innerhalb von Namespaces.

  Phase 3: Assembler-spezifische Fehlerdiagnose
   * 1. Neue Diagnose-Regeln:
       * ❌ Offen. Die existierenden Diagnosen sind allgemein (z.B. "Label unbenutzt"). Es gibt keine spezifischen Prüfungen für Scopes, Argumente oder Typen.
   * 2. Multi-Datei-Projekte (#import):
       * ❌ Offen. #import wird noch nicht verarbeitet.