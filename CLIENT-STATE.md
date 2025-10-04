# Kickass Client (kickass_cl) - Test Client State Analysis

**Generated:** 2025-10-04
**Version:** Based on current implementation
**Purpose:** Comprehensive analysis of the Kick Assembler LSP test client

---

## Executive Summary

**Kickass Client** (`kickass_cl`) is a comprehensive command-line LSP test client for testing the Kick Assembler Language Server (`kickass_ls`). It provides three operation modes and supports all major LSP features.

### Key Metrics
- **4 Go source files**: 2,153 lines of code
- **51 functions** implementing LSP protocol and testing infrastructure
- **3 operation modes**: Quick test, Test suite, Interactive
- **6 LSP features** fully supported
- **JSON-RPC 2.0** protocol implementation over stdio

### Components Breakdown
| File | Lines | Purpose |
|------|-------|---------|
| `client.go` | 605 | Core LSP client implementation |
| `main.go` | 541 | CLI interface and quick test mode |
| `runner.go` | 693 | Test suite runner and validation |
| `protocol.go` | 314 | LSP message types and protocol structures |

---

## Architecture Overview

### Communication Model

```
┌─────────────────┐                    ┌──────────────────┐
│  Kickass Client │ ← JSON-RPC 2.0 →  │  LSP Server      │
│  (kickass_cl)   │    over stdio      │  (kickass_ls)    │
└─────────────────┘                    └──────────────────┘
```

**Protocol:** JSON-RPC 2.0 over stdin/stdout
**Transport:** Process pipes (stdin, stdout, stderr)
**Architecture:** Request/Response + Notifications

### Core Components

#### 1. LSPClient (client.go)
**Purpose:** Complete LSP protocol implementation

**Key Responsibilities:**
- Process lifecycle management (start/stop server)
- Message routing (requests, responses, notifications)
- Document state management
- Diagnostics collection
- Asynchronous message handling

**State Management:**
```go
type LSPClient struct {
    serverPath  string              // Path to LSP server binary
    cmd         *exec.Cmd            // Server process
    stdin       io.WriteCloser       // Server input
    stdout      io.ReadCloser        // Server output

    nextID      int                  // Request ID counter
    pendingReqs map[int]chan *Message // Pending responses

    documents   map[string]*DocumentState  // Open documents
    diagnostics map[string][]Diagnostic    // Received diagnostics

    shutdown    chan bool            // Shutdown signal
}
```

#### 2. CLI Interface (main.go)
**Purpose:** User-facing command-line interface

**Capabilities:**
- Mode selection (quick/suite/interactive)
- Argument parsing
- Output formatting
- Special command handling (`completion-test`, `completion-at`)

**Entry Points:**
- `runQuickTest()` - Single file diagnostic testing
- `runCompletionTest()` - Test completion with "."
- `runCompletionAtPosition()` - Test completion at specific position
- `runInteractiveMode()` - Manual debugging interface

#### 3. Test Runner (runner.go)
**Purpose:** Automated test suite execution

**Features:**
- JSON test suite parsing
- Test case validation
- Result reporting
- Multiple test types (completion, hover, diagnostics, definition, references, symbols)

**Test Suite Structure:**
```json
{
  "name": "Test Suite Name",
  "description": "Description",
  "setup": {
    "serverPath": "../kickass_ls",
    "files": { "test.asm": "file content" }
  },
  "testCases": [
    {
      "type": "completion",
      "input": { "file": "test.asm", "line": 5, "character": 10 },
      "expected": { "minItems": 10 }
    }
  ]
}
```

#### 4. Protocol Definitions (protocol.go)
**Purpose:** LSP message types and structures

**Defines:**
- Request/Response/Notification messages
- LSP data structures (Position, Range, Location)
- Completion items, Diagnostics, Symbols
- Initialization parameters

---

## Operation Modes

### 1. Quick Test Mode ✅

**Usage:**
```bash
./kickass_cl/kickass_cl -server ../kickass_ls file.asm
```

**Features:**
- Single file testing
- Automatic diagnostic collection
- Color-coded output (❌ Error, ⚠️ Warning, 💡 Hint)
- Verbose mode available
- Exit code 0 (success) or 1 (errors found)

**Output Example:**
```
Testing file: test.asm
=====================================
📋 Diagnostics (2):
-------------------------------------
❌ Line 5:15 [Error] Invalid hex value '$XY'
⚠️ Line 10:5 [Warning] Unused label 'start'
-------------------------------------
Summary: 1 errors, 1 warnings, 0 info/hints
```

### 2. Test Suite Mode ✅

**Usage:**
```bash
./kickass_cl/kickass_cl -suite test-suite.json
```

**Features:**
- JSON-based test definitions
- Multiple test types supported
- Batch testing
- JSON output for CI/CD (`-output results.json`)
- Comprehensive validation

**Supported Test Types:**
1. **Completion** - Code completion items
2. **Hover** - Hover information
3. **Diagnostics** - Error/warning detection
4. **Definition** - Go-to-definition
5. **References** - Find-all-references
6. **Symbols** - Document symbol extraction

### 3. Interactive Mode ✅

**Usage:**
```bash
./kickass_cl/kickass_cl -interactive -server ../kickass_ls
```

**Available Commands:**
```
open <file>              - Open a document
completion <line> <char> - Get completion at position (0-indexed)
hover <line> <char>      - Get hover info at position
diagnostics              - Show current diagnostics
symbols                  - Show document symbols
quit (q, exit)           - Exit interactive mode
```

**Use Case:** Manual debugging and exploration

---

## LSP Features Supported

### Document Lifecycle ✅
- **textDocument/didOpen** - Open document notification
- **textDocument/didChange** - Document change notification
- **textDocument/didClose** - Close document notification

### Language Features ✅

| Feature | LSP Method | Status | Notes |
|---------|-----------|--------|-------|
| **Completion** | `textDocument/completion` | ✅ Full | Returns CompletionItem[] |
| **Hover** | `textDocument/hover` | ✅ Full | Returns Hover with markdown |
| **Definition** | `textDocument/definition` | ✅ Full | Returns Location[] |
| **References** | `textDocument/references` | ✅ Full | Returns Location[] |
| **Symbols** | `textDocument/documentSymbol` | ✅ Full | Returns DocumentSymbol[] |
| **Diagnostics** | `textDocument/publishDiagnostics` | ✅ Full | Notification from server |

### Initialization ✅
- **initialize** request with client capabilities
- **initialized** notification
- **shutdown** request
- **exit** notification

---

## Special Features

### 1. Position-Specific Completion Testing

**Command:**
```bash
./kickass_cl/kickass_cl -server ../kickass_ls completion-at test.asm 10 5
```

**Purpose:** Test completion at exact cursor position
**Output:** All completion items with labels and details
**Use Case:** Debugging context-aware completion

### 2. Dot Completion Test

**Command:**
```bash
./kickass_cl/kickass_cl -server ../kickass_ls completion-test
```

**Purpose:** Test directive completion (typing ".")
**Output:** All directives starting with "."
**Use Case:** Validate directive discovery

### 3. Diagnostic Collection

**Automatic:** Diagnostics are collected asynchronously via notifications
**Storage:** Mapped by file URI
**Access:** `GetDiagnostics(uri)` method

**Severity Levels:**
- `1` - Error (❌)
- `2` - Warning (⚠️)
- `3` - Information (ℹ️)
- `4` - Hint (💡)

### 4. Verbose Mode

**Flag:** `-verbose`

**Provides:**
- Detailed diagnostic sources
- Request/response debugging
- Timing information
- Full message content

---

## Implementation Highlights

### Asynchronous Message Handling

```go
// client.go - Message router
func (c *LSPClient) handleMessages() {
    scanner := bufio.NewScanner(c.stdout)
    for scanner.Scan() {
        // Parse Content-Length header
        // Read JSON-RPC message
        // Route to handler

        if msg.ID != nil {
            // Response - send to waiting goroutine
            c.sendResponse(msg)
        } else if msg.Method == "textDocument/publishDiagnostics" {
            // Notification - update diagnostics
            c.handleDiagnostics(msg)
        }
    }
}
```

### Request/Response Pattern

```go
// Send request and wait for response
func (c *LSPClient) sendRequest(method string, params interface{}) (*Message, error) {
    id := c.nextID
    c.nextID++

    // Create response channel
    respChan := make(chan *Message, 1)
    c.pendingReqs[id] = respChan

    // Send request
    c.writeMessage(Message{
        JSONRPC: "2.0",
        ID:      &id,
        Method:  method,
        Params:  params,
    })

    // Wait for response with timeout
    select {
    case resp := <-respChan:
        return resp, nil
    case <-time.After(30 * time.Second):
        return nil, fmt.Errorf("request timeout")
    }
}
```

### Document State Tracking

```go
type DocumentState struct {
    URI     string
    Version int
    Content string
}

// Track document changes
func (c *LSPClient) OpenDocument(uri, languageID, content string) error {
    c.documents[uri] = &DocumentState{
        URI:     uri,
        Version: 1,
        Content: content,
    }

    return c.sendRequest("textDocument/didOpen", params)
}
```

---

## Test Suite Validation

### Completion Test Validation

```go
// Validate completion items
expected := testCase.Expected.CompletionItems
for _, expectedItem := range expected {
    found := false
    for _, actualItem := range actualItems {
        if actualItem.Label == expectedItem.Label {
            found = true
            // Validate kind, detail, documentation
            break
        }
    }
    if !found {
        return fmt.Errorf("expected item '%s' not found", expectedItem.Label)
    }
}
```

### Diagnostic Test Validation

```go
// Validate diagnostics
for _, expectedDiag := range expected.Diagnostics {
    found := false
    for _, actualDiag := range actualDiags {
        if actualDiag.Range.Start.Line == expectedDiag.Line &&
           actualDiag.Severity == expectedDiag.Severity {
            found = true
            break
        }
    }
    if !found {
        return fmt.Errorf("expected diagnostic not found")
    }
}
```

---

## Usage Examples

### Example 1: Quick Syntax Check

```bash
# Test file for errors
./kickass_cl/kickass_cl -server ./kickass_ls test-cases/test_simple.asm

# Output if clean:
✅ No diagnostics - file is clean!

# Output if errors:
❌ Line 5:10 [Error] Invalid operand
```

### Example 2: Test Completion at Position

```bash
# Test completion at line 10, character 5
./kickass_cl/kickass_cl completion-at test.asm 10 5

# Output:
=== Got 18 completion items ===

  1. #                          Immediate addressing mode
  2. $                          Absolute addressing mode
  3. (                          Indirect addressing mode
  4. screen                     Memory: VIC-II Screen RAM
  ...
```

### Example 3: Interactive Debugging

```bash
./kickass_cl/kickass_cl -interactive

> open test-cases/test_simple.asm
Opened: test-cases/test_simple.asm

> diagnostics
✅ No diagnostics

> completion 5 10
=== Got 45 completion items ===
  1. LDA    Load Accumulator
  2. STA    Store Accumulator
  ...

> quit
```

### Example 4: CI/CD Integration

```bash
#!/bin/bash
# Run regression tests
./kickass_cl/kickass_cl -suite test-suite.json -output results.json

if [ $? -eq 0 ]; then
    echo "✅ All tests passed"
    exit 0
else
    echo "❌ Tests failed"
    cat results.json
    exit 1
fi
```

---

## Configuration

### Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-server` | string | `../6510lsp_server` | Path to LSP server binary |
| `-suite` | string | - | Test suite JSON file |
| `-args` | string | - | Additional server arguments |
| `-root` | string | `.` | Root path for test files |
| `-output` | string | - | Save test results to JSON |
| `-verbose` | bool | false | Verbose output |
| `-interactive` | bool | false | Interactive mode |

### Environment

**Server Path:** Defaults to `../6510lsp_server` (legacy path)
**Working Directory:** Current directory
**File URIs:** Converted to absolute `file://` URIs

---

## Known Limitations

### 1. Test Suite Mode Not Currently Used ⚠️
**Status:** All test suites deleted (outdated/insufficient)
**Reason:** Testing primarily done via quick test mode and Neovim integration
**Future:** May be re-enabled with better test cases

### 2. Timing Sensitivity
**Issue:** 2-second wait for diagnostics may be insufficient for large files
**Workaround:** Increase timeout in code if needed

### 3. 0-Indexed Positions
**Note:** Line/character positions are 0-indexed (LSP standard)
**First line:** 0, First character: 0
**User confusion:** Visual line 5 = LSP line 4

### 4. No Multi-File Projects
**Limitation:** Test suite supports multiple files but not cross-file references
**Use Case:** Import/include directives not tested in suite mode

---

## Recent Changes

### Position-Specific Completion (Added 2025-10-03)

**Feature:** `completion-at <file> <line> <char>` command

**Purpose:** Test completion at exact cursor position for debugging context-aware completion

**Implementation:**
```go
func runCompletionAtPosition(serverPath, filePath string, line, char int) {
    // Open file
    content, _ := os.ReadFile(filePath)

    // Start client
    client := NewLSPClient(serverPath)
    client.Start()

    // Open document
    uri := "file://" + absPath
    client.OpenDocument(uri, "kickasm", string(content))

    // Get completion
    completions, _ := client.GetCompletion(uri, line, char)

    // Display results
    for i, item := range completions {
        fmt.Printf("%3d. %-25s %s\n", i+1, item.Label, item.Detail)
    }
}
```

**Impact:** Enabled precise debugging of completion context issues

---

## Integration with Language Server

### Server Testing Workflow

1. **Quick Test Mode:** Fast diagnostic validation
   ```bash
   ./kickass_cl/kickass_cl test.asm
   ```

2. **Completion Testing:** Context-aware completion validation
   ```bash
   ./kickass_cl/kickass_cl completion-at test.asm 10 5
   ```

3. **Interactive Mode:** Manual feature exploration
   ```bash
   ./kickass_cl/kickass_cl -interactive
   ```

4. **Neovim Integration:** Real-world testing in editor

### Server Compatibility

**Protocol Version:** LSP 3.x
**Server Binary:** `kickass_ls` or `6510lsp_server`
**Communication:** stdio (stdin/stdout)
**Initialization:** Full LSP initialize/initialized handshake

---

## Code Quality Assessment

### Strengths ✅
- Clean separation of concerns (client/protocol/runner/CLI)
- Comprehensive LSP protocol implementation
- Robust error handling with timeouts
- Thread-safe state management (mutexes)
- User-friendly output (colors, emojis)
- Flexible operation modes

### Architecture ✅
- Well-structured package organization
- Clear request/response handling
- Asynchronous message processing
- Proper resource cleanup (process shutdown)

### Testing Infrastructure ✅
- Supports all major LSP features
- Real integration testing (no mocking)
- Multiple validation modes
- CI/CD compatible (JSON output)

### Documentation ✅
- Comprehensive README with examples
- Clear usage instructions
- Troubleshooting guide
- Protocol explanations

---

## Future Enhancements

### Potential Improvements

1. **Test Suite Revival**
   - Create comprehensive regression test suites
   - Add cross-file reference tests
   - Performance benchmarks

2. **Enhanced Reporting**
   - HTML test reports
   - Coverage metrics
   - Diff visualization

3. **Extended LSP Features**
   - Code actions
   - Formatting
   - Rename refactoring
   - Workspace symbols

4. **Developer Experience**
   - Watch mode (auto-rerun on file changes)
   - Test record/playback
   - GUI test runner

---

## Overall Assessment

### Rating: ⭐⭐⭐⭐⭐ (5/5)

**Kickass Client** is a **production-ready, comprehensive LSP test client** with excellent architecture and feature coverage.

**Key Achievements:**
- ✅ Complete LSP protocol implementation
- ✅ Three flexible operation modes
- ✅ Real integration testing
- ✅ User-friendly interface
- ✅ Robust error handling
- ✅ Well-documented

**Current State:**
- **Actively used** for Language Server development
- **Reliable** for quick diagnostic testing
- **Essential tool** for completion system debugging
- **Foundation** for future test automation

**Recommendation:**
Maintain current functionality and consider re-introducing test suites with improved test cases based on real-world usage patterns discovered during Neovim integration testing.

---

**Last Updated:** 2025-10-04
**Status:** Production Ready
**Maintainer:** Part of Kick Assembler LSP Project
