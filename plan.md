# Neuer Plan für die Weiterentwicklung

  Basierend auf dieser Analyse schlage ich folgenden, aktualisierten Plan vor, der auf dem bisher Erreichten aufbaut:

  Phase 1: Vervollständigung der Kern-IDE-Funktionen

   1. Signaturen für Hover & Completion:
       * Ziel: Beim Hovern über eine Funktion oder ein Makro soll dessen Signatur (Parameter, Rückgabewert) angezeigt werden.
       * Schritte:
           1. Erweitere die KickassDirective-Struktur (oder eine neue Struktur für Funktionen/Makros) um Signatur-Informationen.
           2. Passe den Parser an, um diese Signaturen zu extrahieren.
           3. Aktualisiere den textDocument/hover-Handler, um diese Informationen anzuzeigen.

   2. Signature Help (Parameter-Info):
       * Ziel: Während der Eingabe von Argumenten für eine Funktion/Makro soll ein Hilfsfenster mit den erwarteten Parametern erscheinen.
       * Schritte: Implementiere den textDocument/signatureHelp-Request des Language Server Protocols.

  Phase 2: Erweiterte Diagnose-Funktionen

   1. Scope-basierte Diagnose:
       * Ziel: Der Server soll warnen, wenn auf ein Symbol zugegriffen wird, das im aktuellen Geltungsbereich (Scope) nicht sichtbar ist.
       * Schritte: Erweitere die publishDiagnostics-Funktion, um die Sichtbarkeit von Symbolen bei jeder Verwendung zu prüfen.

   2. Argumenten-Prüfung für Funktionen/Makros:
       * Ziel: Fehler melden, wenn eine Funktion oder ein Makro mit einer falschen Anzahl von Argumenten aufgerufen wird.
       * Schritte: Nutze die in Phase 1 gesammelten Signatur-Informationen, um die Aufrufe in der publishDiagnostics-Funktion zu validieren.

  Phase 3: Projektweite Intelligenz

   1. Multi-Datei-Unterstützung via `#import`:
       * Ziel: Der Server soll #import-Anweisungen folgen, die importierten Dateien ebenfalls parsen und deren Symbole für das gesamte Projekt verfügbar machen.
       * Schritte:
           1. Erweitere den documentStore und symbolStore, um mehrere Dokumente gleichzeitig zu verwalten.
           2. Implementiere eine Logik, die bei didOpen oder didChange rekursiv Imports auflöst.
           3. Passe alle Symbol-Suchfunktionen (FindSymbol, FindAllVisibleSymbols etc.) an, damit sie projektweit suchen können.
