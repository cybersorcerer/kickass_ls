# 6510 LSP Test Client

Ein vollwertiger LSP Test Client für echte Integration Tests des 6510 Language Servers.

## Features

- **Echter LSP Client** - Kommuniziert über JSON-RPC mit dem Live Server
- **Vollständige LSP Unterstützung** - Completion, Hover, Diagnostics, Definition, References, Symbols
- **Test Suite System** - JSON-basierte Test-Case Definitionen
- **Interactive Mode** - Manuelle Tests und Debugging
- **Robuste Integration** - Nutzt den Server "as-is" ohne Code-Duplikation

## Build

```bash
cd test-client
go build -o test-client .
```

## Usage

### Test Suite ausführen
```bash
./test-client -suite ../test-cases/basic-completion.json
./test-client -suite ../test-cases/diagnostics.json -verbose
```

### Interactive Mode für manuelle Tests
```bash
./test-client -interactive -server ../6510lsp_server
```

### Optionen
- `-suite <file>` - Test Suite JSON Datei
- `-server <path>` - Pfad zum LSP Server (Standard: ../6510lsp_server)
- `-root <path>` - Root Verzeichnis für Test-Dateien
- `-output <file>` - Test-Ergebnisse als JSON speichern
- `-verbose` - Detaillierte Ausgabe
- `-interactive` - Interactive Mode

## Test Suite Format

```json
{
  "name": "Basic Completion Tests",
  "description": "Test completion functionality",
  "setup": {
    "serverPath": "../6510lsp_server",
    "serverArgs": [],
    "rootPath": "../test-cases",
    "files": {
      "test.asm": "start:\n    lda #$02\n    sta $D0"
    }
  },
  "testCases": [
    {
      "name": "Memory Register Completion",
      "description": "Test $ completion shows memory registers",
      "type": "completion",
      "input": {
        "file": "test.asm",
        "line": 2,
        "character": 9
      },
      "expected": {
        "minItems": 10,
        "completionItems": [
          {
            "label": "$D000",
            "kind": 12,
            "detail": "VIC-II - Sprite 0 X Position"
          }
        ]
      }
    }
  ]
}
```

## Test Types

### completion
Testet Code-Vervollständigung:
```json
{
  "type": "completion",
  "expected": {
    "minItems": 5,
    "maxItems": 50,
    "completionItems": [
      {
        "label": "LDA",
        "kind": 3,
        "detail": "Load Accumulator",
        "documentation": "Load Accumulator with Memory"
      }
    ]
  }
}
```

### hover
Testet Hover-Informationen:
```json
{
  "type": "hover",
  "expected": {
    "hoverContent": "Load Accumulator",
    "hoverRange": {
      "start": {"line": 1, "character": 4},
      "end": {"line": 1, "character": 7}
    }
  }
}
```

### diagnostics
Testet Error/Warning Detection:
```json
{
  "type": "diagnostics",
  "expected": {
    "diagnostics": [
      {
        "line": 1,
        "column": 10,
        "severity": 1,
        "message": "Undefined symbol"
      }
    ]
  }
}
```

### definition
Testet Go-to-Definition:
```json
{
  "type": "definition",
  "expected": {
    "locations": [
      {
        "file": "test.asm",
        "line": 0,
        "character": 0
      }
    ]
  }
}
```

### references
Testet Find References:
```json
{
  "type": "references",
  "expected": {
    "locations": [
      {"file": "test.asm", "line": 0, "character": 0},
      {"file": "test.asm", "line": 5, "character": 8}
    ]
  }
}
```

### symbols
Testet Document Symbols:
```json
{
  "type": "symbols",
  "expected": {
    "symbols": [
      {
        "name": "start",
        "kind": 13,
        "line": 0,
        "detail": "Label"
      }
    ]
  }
}
```

## Vorteile gegenüber alten Test-Modi

✅ **Echte Integration** - Testet den kompletten LSP-Stack
✅ **Konsistenz** - Identisches Verhalten wie echte Editoren
✅ **Wartungsfreundlich** - Server-Änderungen brechen Tests nicht
✅ **Vollständige Abdeckung** - Document Lifecycle, Async Events, etc.
✅ **Flexibilität** - JSON-basierte Test-Definitionen
✅ **Debugging** - Interactive Mode für manuelle Tests

## Interactive Mode Kommandos

```
open <file>              - Öffne Dokument
completion <line> <char> - Completion an Position
hover <line> <char>      - Hover an Position
diagnostics              - Zeige aktuelle Diagnostics
symbols                  - Zeige Document Symbols
quit                     - Beenden
```

## Beispiel Session

```bash
$ ./test-client -interactive
LSP Test Client - Interactive Mode
==================================
Initializing LSP server...
Server initialized. Available commands:
> open test.asm
Opened: test.asm
> completion 2 10
Completion items (23):
  $D000 - VIC-II - Sprite 0 X Position
  $D001 - VIC-II - Sprite 0 Y Position
  ...
> hover 1 4
Hover content:
**LDA** - Load Accumulator with Memory
...
> quit
```