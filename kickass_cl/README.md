# Kick Assembler LSP Test Client

A comprehensive LSP test client for the Kick Assembler Language Server (`kickass_ls`). This client provides three modes of operation: quick file testing, test suite execution, and interactive debugging.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage Modes](#usage-modes)
  - [Quick Test Mode](#quick-test-mode)
  - [Test Suite Mode](#test-suite-mode)
  - [Interactive Mode](#interactive-mode)
- [Test Suite Format](#test-suite-format)
- [LSP Features Supported](#lsp-features-supported)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

---

## Features

✅ **Three Operation Modes:**
- Quick test mode for single files
- Test suite execution for automated testing
- Interactive mode for manual debugging

✅ **Full LSP Protocol Support:**
- Document lifecycle (open, change, close)
- Code completion
- Hover information
- Go-to-definition
- Find references
- Document symbols
- Diagnostics (errors, warnings, hints)

✅ **Real Integration Testing:**
- Tests the actual LSP server binary
- No code duplication or mocking
- Same behavior as real editors (Neovim, VSCode, etc.)

✅ **Developer-Friendly:**
- Colorful, emoji-based output
- Verbose mode for debugging
- JSON output for CI/CD integration

---

## Installation

### Build the Test Client

```bash
cd test-client
go build -o kickass_cl .
```

The binary `kickass_cl` will be created in the `test-client` directory.

### Prerequisites

- Go 1.19 or later
- Built `kickass_ls` server binary in the parent directory

---

## Usage Modes

### Quick Test Mode

Test a single assembly file and get instant diagnostic feedback.

#### Basic Usage

```bash
./kickass_cl -server ../kickass_ls <file.asm>
```

#### Example

```bash
# Test a file
./kickass_cl -server ../kickass_ls ../test-cases/test_simple.asm

# Output:
Testing file: ../test-cases/test_simple.asm
=====================================
✅ No diagnostics - file is clean!
```

#### With Diagnostics

```bash
./kickass_cl -server ../kickass_ls ../test-cases/test_errors.asm

# Output:
Testing file: ../test-cases/test_errors.asm
=====================================

📋 Diagnostics (3):
-------------------------------------
❌ Line 5:15 [Error] Invalid hex value '$XY' - hex values must only contain digits 0-9 and letters A-F
⚠️ Line 10:5 [Warning] Unused label 'start'
💡 Line 15:8 [Hint] Consider UPPER_CASE naming for constant 'my_const'
-------------------------------------
Summary: 1 errors, 1 warnings, 1 info/hints
```

#### Verbose Mode

```bash
./kickass_cl -server ../kickass_ls -verbose file.asm
```

Shows additional information like diagnostic sources.

#### Exit Codes

- `0` - Success (no errors)
- `1` - Errors found or test failed

---

### Test Suite Mode

Run comprehensive test suites defined in JSON files.

#### Basic Usage

```bash
./kickass_cl -suite <test-suite.json>
```

#### Example

```bash
# Run a test suite
./kickass_cl -suite ../test-cases/basic-completion.json

# With verbose output
./kickass_cl -suite ../test-cases/diagnostics.json -verbose

# Save results to JSON
./kickass_cl -suite ../test-cases/full-suite.json -output results.json
```

#### Options

- `-suite <file>` - Path to test suite JSON file
- `-server <path>` - Path to LSP server binary (default: `../kickass_ls`)
- `-output <file>` - Save test results as JSON
- `-verbose` - Show detailed output

---

### Interactive Mode

Manual testing and debugging mode with a command-line interface.

#### Start Interactive Mode

```bash
./kickass_cl -interactive -server ../kickass_ls
```

#### Available Commands

```text
open <file>              - Open a document
completion <line> <char> - Get completion at position (0-indexed)
hover <line> <char>      - Get hover info at position (0-indexed)
diagnostics              - Show current diagnostics for open file
symbols                  - Show document symbols
quit (or q, exit)        - Exit interactive mode
```

#### Example Session

```bash
$ ./kickass_cl -interactive -server ../kickass_ls
LSP Test Client - Interactive Mode
==================================
Initializing LSP server...

Server initialized. Available commands:
> open ../test-cases/example.asm
Opened: ../test-cases/example.asm

> diagnostics
📋 Diagnostics (2):
  ❌ example.asm:10:5 [Error] Undefined symbol 'loop'
  ⚠️ example.asm:15:8 [Warning] Unused variable 'temp'

> hover 5 8
Hover content:
**LDA** - Load Accumulator with Memory

Addressing Modes:
- Immediate: LDA #nn
- Zero Page: LDA nn
- Absolute: LDA nnnn

> completion 10 8
Completion items (45):
  LDA - Load Accumulator
  STA - Store Accumulator
  LDX - Load X Register
  ... and 42 more items

> quit
```

#### Notes

- Line and character positions are **0-indexed** (first line = 0, first character = 0)
- Commands are case-sensitive
- Use absolute or relative paths for files

---

## Test Suite Format

Test suites are defined in JSON format with the following structure:

### Basic Structure

```json
{
  "name": "Basic Completion Tests",
  "description": "Test completion functionality",
  "setup": {
    "serverPath": "../kickass_ls",
    "serverArgs": [],
    "rootPath": "../test-cases",
    "files": {
      "test.asm": "start:\n    lda #$02\n    sta $D020"
    }
  },
  "testCases": [
    {
      "name": "Mnemonic Completion",
      "description": "Test that mnemonics appear in completion",
      "type": "completion",
      "input": {
        "file": "test.asm",
        "line": 1,
        "character": 8
      },
      "expected": {
        "minItems": 5,
        "completionItems": [
          {
            "label": "LDA",
            "kind": 3,
            "detail": "Load Accumulator"
          }
        ]
      }
    }
  ]
}
```

### Test Types

#### 1. Completion Tests

Test code completion functionality.

```json
{
  "type": "completion",
  "input": {
    "file": "test.asm",
    "line": 5,
    "character": 10
  },
  "expected": {
    "minItems": 10,
    "maxItems": 100,
    "completionItems": [
      {
        "label": "$D020",
        "kind": 12,
        "detail": "Border Color Register",
        "documentation": "VIC-II border color"
      }
    ]
  }
}
```

**CompletionItemKind Values:**
- `1` - Text
- `3` - Function
- `6` - Variable
- `12` - Value
- `13` - Enum
- `14` - Keyword

#### 2. Hover Tests

Test hover information display.

```json
{
  "type": "hover",
  "input": {
    "file": "test.asm",
    "line": 3,
    "character": 5
  },
  "expected": {
    "hoverContent": "Load Accumulator",
    "hoverRange": {
      "start": {"line": 3, "character": 4},
      "end": {"line": 3, "character": 7}
    }
  }
}
```

#### 3. Diagnostics Tests

Test error and warning detection.

```json
{
  "type": "diagnostics",
  "input": {
    "file": "test_errors.asm"
  },
  "expected": {
    "diagnostics": [
      {
        "line": 5,
        "column": 10,
        "severity": 1,
        "message": "Invalid hex value",
        "source": "context-aware-parser"
      }
    ]
  }
}
```

**Severity Levels:**
- `1` - Error
- `2` - Warning
- `3` - Information
- `4` - Hint

#### 4. Definition Tests

Test go-to-definition functionality.

```json
{
  "type": "definition",
  "input": {
    "file": "test.asm",
    "line": 10,
    "character": 8
  },
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

#### 5. References Tests

Test find-all-references functionality.

```json
{
  "type": "references",
  "input": {
    "file": "test.asm",
    "line": 5,
    "character": 10,
    "includeDeclaration": true
  },
  "expected": {
    "locations": [
      {"file": "test.asm", "line": 0, "character": 0},
      {"file": "test.asm", "line": 5, "character": 8},
      {"file": "test.asm", "line": 15, "character": 12}
    ]
  }
}
```

#### 6. Symbols Tests

Test document symbol extraction.

```json
{
  "type": "symbols",
  "input": {
    "file": "test.asm"
  },
  "expected": {
    "symbols": [
      {
        "name": "start",
        "kind": 13,
        "line": 0,
        "detail": "Label"
      },
      {
        "name": "loop",
        "kind": 13,
        "line": 5,
        "detail": "Label"
      }
    ]
  }
}
```

**SymbolKind Values:**

- `12` - Number
- `13` - Function (used for labels)
- `14` - Variable
- `15` - Constant

---

## LSP Features Supported

### Document Lifecycle

- **Open Document** - `textDocument/didOpen`
- **Change Document** - `textDocument/didChange`
- **Close Document** - `textDocument/didClose`

### Language Features

- **Completion** - `textDocument/completion`
- **Hover** - `textDocument/hover`
- **Definition** - `textDocument/definition`
- **References** - `textDocument/references`
- **Document Symbols** - `textDocument/documentSymbol`

### Diagnostics

- **Publish Diagnostics** - `textDocument/publishDiagnostics` (notification)
- Automatic updates when document changes
- Support for errors, warnings, information, and hints

---

## Examples

### Example 1: Quick Test for Syntax Errors

```bash
# Test a file with potential syntax errors
./kickass_cl -server ../kickass_ls ../test-cases/syntax_test.asm
```

### Example 2: Test Completion in Verbose Mode

Create a test suite `completion-test.json`:

```json
{
  "name": "Completion Test",
  "setup": {
    "serverPath": "../kickass_ls",
    "rootPath": "../test-cases",
    "files": {
      "test.asm": ".const START = $0801\n* = START\n    l"
    }
  },
  "testCases": [
    {
      "name": "Mnemonic starts with L",
      "type": "completion",
      "input": {
        "file": "test.asm",
        "line": 2,
        "character": 5
      },
      "expected": {
        "minItems": 3,
        "completionItems": [
          {"label": "LDA"},
          {"label": "LDX"},
          {"label": "LDY"}
        ]
      }
    }
  ]
}
```

Run it:

```bash
./kickass_cl -suite completion-test.json -verbose
```

### Example 3: Interactive Debugging

```bash
./kickass_cl -interactive -server ../kickass_ls

> open ../test-cases/complex.asm
> diagnostics
> hover 10 8
> completion 15 12
> symbols
> quit
```

### Example 4: CI/CD Integration

```bash
# Run tests and save results for CI
./kickass_cl -suite ../test-cases/full-regression.json -output test-results.json

# Check exit code
if [ $? -eq 0 ]; then
  echo "All tests passed!"
else
  echo "Tests failed!"
  exit 1
fi
```

---

## Troubleshooting

### Server Not Found

```
Error: failed to start server: no such file or directory
```

**Solution:** Specify the correct server path:

```bash
./kickass_cl -server /path/to/kickass_ls file.asm
```

### No Diagnostics Received

If diagnostics aren't showing up:

1. Check that the file has actual issues
2. Try increasing the wait timeout in the code (default: 2 seconds)
3. Use verbose mode: `-verbose`
4. Check server logs: `~/.local/share/kickass_ls/kickass_ls.log`

### Connection Timeout

```
Error: request timeout
```

**Possible causes:**

- Server is taking too long to respond
- Server crashed
- Deadlock in server

**Solution:**

- Check server logs
- Try with a simpler test file
- Restart the test client

### Incorrect Line/Column

Remember: **Line and column numbers are 0-indexed!**

- First line = 0
- First character = 0

Example:

```text
Visual line 5, column 10  →  LSP line 4, character 9
```

### Test Suite Fails to Load

```text
Error: failed to load test suite
```

**Solution:**

- Check JSON syntax is valid
- Verify file paths in test suite are correct
- Ensure `setup.files` contains valid assembly code

---

## Advanced Usage

### Custom Server Arguments

```bash
./kickass_cl -server ../kickass_ls -args "--debug --log-level=trace" -interactive
```

### Testing Multiple Files

Create a test suite with multiple files:

```json
{
  "setup": {
    "files": {
      "main.asm": "* = $0801\n.import source \"lib.asm\"\njsr init",
      "lib.asm": "init:\n    lda #$00\n    rts"
    }
  }
}
```

### Custom Root Path

```bash
./kickass_cl -root /path/to/project -suite test.json
```

---

## Contributing

When adding new test cases:

1. Place test suite files in `../test-cases/`
2. Use descriptive test names
3. Document expected behavior
4. Test both success and failure cases

---

## Architecture

The test client uses the Language Server Protocol (LSP) over stdio:

```
Test Client (kickass_cl)  ←→  LSP Server (kickass_ls)
     JSON-RPC 2.0 over stdio
```

**Key Components:**

- `client.go` - LSP client implementation
- `protocol.go` - LSP message types
- `runner.go` - Test suite runner
- `main.go` - CLI interface

---

## License

Part of the Kick Assembler LSP project.

---

## See Also

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- [Kick Assembler Documentation](http://www.theweb.dk/KickAssembler/)
- [Main project README] (../README.md)
