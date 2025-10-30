# Kick Assembler Language Server - Geplante Erweiterungen

Dieses Dokument beschreibt geplante Features und Verbesserungen für den Kick Assembler Language Server.

## Status Übersicht

### ✅ Bereits implementiert

- **Goto Definition** - Springe zur Definition von Labels, Macros, Funktionen (inkl. Multi-Labels)
- **Find References** - Finde alle Referenzen zu einem Symbol (inkl. Multi-Labels)
- **Document Symbols** - Symbol-Übersicht für aktuelle Datei
- **Semantic Tokens** - Syntax-Highlighting basierend auf semantischer Analyse
- **Diagnostics** - Fehler, Warnungen und Hinweise
  - Branch Distance Validation (inkl. Multi-Labels)
  - Memory Map Analysis
  - Dead Code Detection
  - Zero-Page Optimization Hints
- **Hover** - Zeigt Symbol-Informationen beim Überfahren mit der Maus
- **Completion** - Auto-Completion für Mnemonics, Direktiven, Labels

---

## 🎯 Phase 1 - Essentials (Must-Have)

### 1. Document Formatting (`textDocument/formatting`)

**Priorität**: ⭐⭐⭐⭐⭐

**Beschreibung**: Automatisches Formatieren von Kick Assembler Code mit konsistenter Einrückung und Stil.

**Features**:
- Konsistente Einrückung für Code-Blöcke (`.macro`, `.function`, `.namespace`)
- Alignment von Kommentaren
- Spacing zwischen Mnemonics und Operanden
- Konfigurierbare Formatierungsregeln

**Beispiel**:
```kickasm
// Vorher:
lda #$01
sta $d020
.macro test(param){
lda param
}

// Nachher:
    lda #$01
    sta $d020

.macro test(param) {
    lda param
}
```

**Implementierung**:
- `textDocument/formatting` Handler
- `textDocument/rangeFormatting` für Selection-Formatting
- Konfiguration über LSP Settings

---

### 2. Rename Symbol (`textDocument/rename`)

**Priorität**: ⭐⭐⭐⭐⭐

**Beschreibung**: Umbenennen von Symbolen (Labels, Macros, Funktionen, Variablen) über alle Dateien hinweg.

**Features**:
- Rename von Labels (inkl. Multi-Labels)
- Rename von Macro/Function/pseudocommand Namen
- Rename von Namespace-Mitgliedern
- Rename von Variablen und Konstannten
- Preview der Änderungen vor dem Ausführen

**Beispiel**:
```kickasm
// Umbenennen von "oldLabel" zu "newLabel"
oldLabel:
    lda #$01
    jmp oldLabel

// →

newLabel:
    lda #$01
    jmp newLabel
```

**Implementierung**:
- `textDocument/rename` Handler
- `textDocument/prepareRename` für Validierung
- Multi-File Support über Workspace

---

### 3. Signature Help mit Kommentaren (`textDocument/signatureHelp`)

**Priorität**: ⭐⭐⭐⭐⭐

**Beschreibung**: Zeige Parameter-Beschreibungen und Dokumentation beim Aufruf von Macros, Functions und Pseudocommands.

**Features**:
- Parse Kommentare über `.macro`, `.function`, `.pseudocommand` Definitionen
- Zeige Parameter-Beschreibungen
- Zeige Return-Type (bei Functions)
- Unterstütze JSDoc-ähnliche Syntax (`@param`, `@return`)

**Beispiel**:
```kickasm
// Multiply two 16-bit numbers
// @param num1 - First 16-bit number
// @param num2 - Second 16-bit number
// @return Product of num1 * num2
.function multiply16(num1, num2) {
    .return num1 * num2
}

// Bei Aufruf: multiply16(
// → Zeigt: multiply16(num1: First 16-bit number, num2: Second 16-bit number)
```

**Implementierung**:
- Erweitere Parser um Kommentar-Extraktion
- `textDocument/signatureHelp` Handler
- Speichere Kommentare im Symbol Store

---

### 4. Code Actions (`textDocument/codeAction`)

**Priorität**: ⭐⭐⭐⭐⭐

**Beschreibung**: Quick Fixes und Refactoring-Aktionen für häufige Probleme.

**Features**:

#### Quick Fixes:
- **Fix Branch Distance**: Konvertiere zu JMP wenn Branch zu weit
  ```kickasm
  bne far_label  // ERROR: Branch distance out of range
  // Quick Fix → jmp far_label
  ```

- **Convert to Zero-Page**: Optimiere Absolute zu Zero-Page Addressing
  ```kickasm
  lda $0080  // HINT: Could use zero-page addressing
  // Quick Fix → lda $80
  ```

- **Add Missing Import**: Füge fehlende `.import` hinzu
  ```kickasm
  // ERROR: Symbol 'external_label' not found
  // Quick Fix → #import "file.asm"
  ```

#### Refactorings:
- **Extract to Macro**: Selektierten Code in Macro extrahieren
- **Inline Macro**: Macro-Aufruf durch Inhalt ersetzen
- **Convert to Namespace**: Code in Namespace verschieben

**Implementierung**:
- `textDocument/codeAction` Handler
- Code Action Provider für verschiedene Diagnostic Types
- Text Edits für Transformationen

---

### 5. Inlay Hints (`textDocument/inlayHint`)

**Priorität**: ⭐⭐⭐⭐⭐

**Beschreibung**: Zeige inline Informationen ohne Hover zu benötigen.

**Features**:

#### Hardware Register Namen:
```kickasm
lda $d020  // → lda $d020 /* VIC-II Border Color */
sta $0400  // → sta $0400 /* Screen RAM */
```

#### Branch Distanzen:
```kickasm
bne !loop-  // → bne !loop- /* -12 bytes */
beq !skip+  // → beq !skip+ /* +5 bytes */
```

#### Macro Parameter Namen:
```kickasm
BasicUpstart(start)  // → BasicUpstart(address: start)
```

#### Berechnete Werte:
```kickasm
lda #<$1000  // → lda #<$1000 /* $00 */
lda #>$1000  // → lda #>$1000 /* $10 */
```

**Implementierung**:
- `textDocument/inlayHint` Handler
- Konfigurierbare Hint-Types (enable/disable per Type)
- C64memory.json Integration für Register-Namen

---

## 🚀 Phase 2 - Productivity (Should-Have)

### 6. Workspace Symbol Search (`workspace/symbol`)

**Priorität**: ⭐⭐⭐⭐

**Beschreibung**: Suche nach Symbolen über alle Dateien im Workspace.

**Features**:
- Fuzzy-Search über alle Symbole
- Filter nach Symbol-Type (Label, Macro, Function, etc.)
- Integration mit Telescope/FZF in Neovim

**Beispiel**:
```
Suche: "init"
→ init_screen (label in main.asm)
→ init_music (macro in music.asm)
→ initialize (function in utils.asm)
```

---

### 7. Hover für Imports mit Definitionen

**Priorität**: ⭐⭐⭐⭐

**Beschreibung**: Zeige beim Hover über `.import`/`#import` die Definitionen aus der importierten Datei.

**Features**:
- Zeige exportierte Symbole der importierten Datei
- Zeige Datei-Pfad (absolut und relativ)
- Zeige Kommentare/Dokumentation der Symbole
- **NICHT** den gesamten Dateiinhalt (zu viel Information)

**Beispiel**:
```kickasm
#import "macros.asm"
       ^^^^^^^^^^^^
       // Hover zeigt:
       // File: /path/to/macros.asm
       // Exports:
       //   - BasicUpstart(address) - Generate C64 BASIC upstart
       //   - ClearScreen() - Clear screen with space characters
       //   - WaitFrame() - Wait for vertical blank
```

**Implementierung**:
- Erweitere Hover Handler
- Parse importierte Dateien für exportierte Symbole
- Cache für Performance

---

### 8. Document Links (`textDocument/documentLink`)

**Priorität**: ⭐⭐⭐

**Beschreibung**: Klickbare Links für `.import` und `#import` Statements.

**Features**:
- Ctrl+Click auf Import → öffnet Datei
- Zeige Import-Pfad als anklickbaren Link
- Unterstütze relative und absolute Pfade

**Beispiel**:
```kickasm
#import "macros.asm"  // ← Ctrl+Click öffnet macros.asm
```

**Implementierung**:
- `textDocument/documentLink` Handler
- Pfad-Resolution für Imports

---

### 9. Folding Range (`textDocument/foldingRange`)

**Priorität**: ⭐⭐⭐

**Beschreibung**: Code-Folding für bessere Übersicht bei großen Dateien.

**Features**:
- Fold `.macro { ... }`
- Fold `.function { ... }`
- Fold `.namespace { ... }`
- Fold Kommentar-Blöcke
- Fold `.if` Blöcke

**Implementierung**:
- `textDocument/foldingRange` Handler
- Brace-Matching für Block-Strukturen

---

### 10. Hover mit C64 Memory Map Informationen

**Priorität**: ⭐⭐⭐⭐

**Beschreibung**: Zeige detaillierte Hardware-Register Informationen beim Hover über Adressen.

**Features**:
- Nutze C64memory.json für Register-Beschreibungen
- Zeige Read/Write Eigenschaften
- Zeige Bit-Beschreibungen für Register
- Warnungen bei ROM/IO-Bereichen

**Beispiel**:
```kickasm
lda $d020
    ^^^^^
    // Hover zeigt:
    // $D020 - VIC-II Border Color Register
    // Type: I/O Register (Read/Write)
    // Bits: 0-3: Border color (0-15)
    // Common values:
    //   $00 - Black
    //   $06 - Blue
    //   $0E - Light Blue
```

**Implementierung**:
- Erweitere Hover Handler
- Parse C64memory.json
- Formatter für Memory Map Informationen

---

## 🎨 Phase 3 - Advanced (Nice-to-Have)

### 11. Call Hierarchy (`textDocument/prepareCallHierarchy`)

**Priorität**: ⭐⭐⭐

**Beschreibung**: Zeige Macro/Function Aufruf-Hierarchie.

**Features**:
- "Incoming Calls": Wo wird dieses Macro aufgerufen?
- "Outgoing Calls": Was ruft dieses Macro auf?
- Visualisierung als Baum-Struktur

---

### 12. Selection Range (`textDocument/selectionRange`)

**Priorität**: ⭐⭐⭐

**Beschreibung**: Smart text selection basierend auf AST-Struktur.

**Features**:
- Erste Selection: Operand
- Zweite Selection: Instruction
- Dritte Selection: Statement
- Vierte Selection: Block

---

### 13. Document Highlight (`textDocument/documentHighlight`)

**Priorität**: ⭐⭐⭐

**Beschreibung**: Highlight alle Vorkommen eines Symbols beim Cursor drauf.

**Features**:
- Highlight von Label-Referenzen
- Highlight von Variable-Usage
- Unterscheide Read/Write (verschiedene Farben)

---

## 📝 Konfiguration

Alle neuen Features sollten konfigurierbar sein über LSP Settings:

```json
{
  "kickass_ls": {
    "formatting": {
      "enabled": true,
      "indentSize": 4,
      "alignComments": true
    },
    "inlayHints": {
      "enabled": true,
      "showBranchDistances": true,
      "showRegisterNames": true,
      "showParameterNames": true,
      "showCalculatedValues": true
    },
    "codeActions": {
      "enabled": true,
      "showRefactorings": true
    },
    "hover": {
      "showMemoryMapInfo": true,
      "showImportedSymbols": true
    }
  }
}
```

---

## 🔄 Implementierungs-Reihenfolge

**Empfohlene Reihenfolge** (nach Impact/Effort Ratio):

1. **Inlay Hints** - Hoher Impact, mittlerer Aufwand
2. **Rename Symbol** - Hoher Impact, mittlerer Aufwand
3. **Document Formatting** - Hoher Impact, hoher Aufwand
4. **Signature Help mit Kommentaren** - Mittlerer Impact, niedriger Aufwand
5. **Code Actions** - Hoher Impact, hoher Aufwand
6. **Hover mit Memory Map** - Mittlerer Impact, niedriger Aufwand
7. **Document Links** - Niedriger Impact, niedriger Aufwand
8. **Workspace Symbols** - Mittlerer Impact, mittlerer Aufwand
9. **Folding Range** - Niedriger Impact, niedriger Aufwand
10. **Hover für Imports** - Niedriger Impact, mittlerer Aufwand

---

## 🤝 Beiträge

Wenn du zu einem dieser Features beitragen möchtest, erstelle bitte ein Issue im Repository oder kontaktiere die Maintainer.

---

**Letzte Aktualisierung**: 2025-10-26
**Version**: 1.0.0
