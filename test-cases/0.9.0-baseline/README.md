# Kick Assembler LSP v0.9.0 Baseline Test Suite

**Version:** 0.9.0
**Date:** 2025-10-13
**Focus:** Core LSP Features & Semantic Token Highlighting
**Purpose:** Establish baseline for current implementation

---

## 📋 Test Suite Overview

This baseline test suite validates the **current state** of the Kick Assembler LSP Server v0.9.0. It focuses on:
- ✅ **Semantic Token Highlighting** (recently fixed enum highlighting bug)
- ✅ **Basic Parsing** (Context-Aware Lexer/Parser)
- ✅ **Diagnostics** (Errors and Warnings)
- ✅ **Completion** (Context-aware suggestions)
- ✅ **Hover** (Symbol information)
- ✅ **Definition/References** (Go to definition, Find references)
- ✅ **Document Symbols** (Outline view)

### Test Coverage

| Test File | Focus Area | Features Tested | Status |
|-----------|-----------|-----------------|--------|
| `01-semantic-tokens.asm` | Semantic highlighting | All token types (mnemonic, directive, number, etc.) | ✅ TODO |
| `02-basic-syntax.asm` | Basic parsing | Labels, mnemonics, directives, comments | ✅ TODO |
| `03-diagnostics.asm` | Error detection | Unknown mnemonics, invalid syntax | ✅ TODO |
| `04-completion.asm` | Code completion | Mnemonics, directives, symbols | ✅ TODO |
| `05-hover.asm` | Hover information | Mnemonics, labels, constants | ✅ TODO |
| `06-goto-definition.asm` | Navigation | Label definitions, constant references | ✅ TODO |
| `07-document-symbols.asm` | Symbols list | Labels, constants, macros, functions | ✅ TODO |

**Total:** 7 test files
**Expected Pass:** 7/7 (v0.9.0 is current stable baseline)

---

## 🎯 Test Objectives

### 1. Semantic Token Highlighting
**File:** `01-semantic-tokens.asm`

**What We Test:**
- ✅ Mnemonics highlighted correctly (magenta)
- ✅ Directives highlighted correctly (purple)
- ✅ Numbers highlighted correctly (orange)
- ✅ Labels highlighted correctly (blue)
- ✅ Comments highlighted correctly (gray)
- ✅ Strings highlighted correctly (green)
- ✅ Operators highlighted correctly (white)
- ✅ Variables highlighted correctly (cyan)
- ✅ Functions/Macros highlighted correctly (yellow)
- ✅ Preprocessor directives highlighted correctly (light blue)

**Critical Test Case:**
```asm
.enum {
    RED = 1,      // RED should be cyan, 1 should be orange
    BLUE = 6,     // BLUE should be cyan, 6 should be orange
    GREEN = 10    // GREEN should be cyan, 10 should be orange (both digits!)
}
```

**Why This Matters:**
This was a critical bug fixed in v0.9.8 where `tryTokenizeMnemonic()` wasn't restoring the column position, causing character-by-character highlighting errors.

---

### 2. Basic Syntax Parsing
**File:** `02-basic-syntax.asm`

**What We Test:**
- ✅ Labels (with colons)
- ✅ Mnemonics (all addressing modes)
- ✅ Directives (.byte, .word, .const, .var, etc.)
- ✅ Comments (single-line //, multi-line /* */)
- ✅ Numbers (hex $ff, decimal 255, binary %11111111)
- ✅ Strings ("hello", 'A')

**Expected Behavior:**
- No parsing errors
- All tokens correctly identified
- Proper AST generation

---

### 3. Diagnostics
**File:** `03-diagnostics.asm`

**What We Test:**
- ❌ Unknown mnemonics (e.g., `xyz`)
- ❌ Invalid syntax (e.g., `lda #$gg`)
- ❌ Undefined labels (e.g., `jmp unknown`)
- ⚠️ Unused variables
- ⚠️ Duplicate definitions

**Expected Behavior:**
- Errors for unknown mnemonics
- Errors for invalid hex numbers
- Warnings for undefined labels
- Warnings for unused symbols

---

### 4. Completion
**File:** `04-completion.asm`

**What We Test:**
- ✅ Mnemonic completion after space
- ✅ Directive completion after `.`
- ✅ Symbol completion in operands
- ✅ Context-aware completion (only labels after `jmp`)

**Expected Behavior:**
- After `.` → Only directives (.byte, .const, etc.)
- After mnemonic → Only valid operands
- After `jmp`/`jsr` → Only labels

---

### 5. Hover Information
**File:** `05-hover.asm`

**What We Test:**
- ✅ Hover over mnemonic → Description (e.g., "LDA - Load Accumulator")
- ✅ Hover over label → Definition location
- ✅ Hover over constant → Value
- ✅ Hover over directive → Documentation

**Expected Behavior:**
- Rich markdown hover content
- Correct information for each symbol type

---

### 6. Go to Definition / Find References
**File:** `06-goto-definition.asm`

**What We Test:**
- ✅ Go to definition for labels
- ✅ Go to definition for constants
- ✅ Find all references for labels
- ✅ Find all references for constants

**Expected Behavior:**
- Jump to correct line
- Show all usage locations

---

### 7. Document Symbols
**File:** `07-document-symbols.asm`

**What We Test:**
- ✅ Labels listed in outline
- ✅ Constants listed in outline
- ✅ Macros listed in outline
- ✅ Functions listed in outline
- ✅ Namespaces listed in outline

**Expected Behavior:**
- Complete symbol hierarchy
- Correct symbol kinds
- Accurate position ranges

---

## 🧪 Running the Tests

### Quick Test (Single File)

```bash
./kickass_cl/kickass_cl --server ./kickass_ls test-cases/0.9.0-baseline/01-semantic-tokens.asm
```

### Test Semantic Tokens Visualization

```bash
./kickass_cl/kickass_cl --server ./kickass_ls semantic-tokens test-cases/0.9.0-baseline/01-semantic-tokens.asm
```

### Full Test Suite (when baseline-suite.json is created)

```bash
./kickass_cl/kickass_cl --suite test-cases/0.9.0-baseline/baseline-suite.json --verbose
```

### With HTML Report

```bash
./kickass_cl/kickass_cl --suite test-cases/0.9.0-baseline/baseline-suite.json --html test-results-0.9.0.html
```

---

## 📊 Expected Results

### Baseline Expectations for v0.9.0

| Category | Expected Errors | Expected Warnings | Expected Pass |
|----------|----------------|-------------------|---------------|
| Semantic Tokens | 0 | 0 | ✅ Yes |
| Basic Syntax | 0 | 0 | ✅ Yes |
| Diagnostics | 3-5 | 2-3 | ✅ Yes |
| Completion | 0 | 0 | ✅ Yes |
| Hover | 0 | 0 | ✅ Yes |
| Go to Definition | 0 | 0 | ✅ Yes |
| Document Symbols | 0 | 0 | ✅ Yes |

**Status:** All tests should **PASS** - this is the stable baseline!

---

## 🔧 Test Client Features Used

This baseline uses the new `kickass_cl` test client features:

### 1. Semantic Token Visualization
```bash
./kickass_cl/kickass_cl --server ./kickass_ls semantic-tokens <file.asm> [line]
```
- Shows colored tokens in terminal
- Displays token counts by type
- Optionally shows details for specific line

### 2. Quick Tests
```bash
./kickass_cl/kickass_cl --server ./kickass_ls <file.asm>
```
- Fast diagnostics check
- Exit code 0 = pass, 1 = fail

### 3. Test Suites
```bash
./kickass_cl/kickass_cl --suite <suite.json> --html <report.html>
```
- Run multiple tests
- Generate HTML report
- Track pass/fail statistics

---

## 📈 Known Working Features

### ✅ Context-Aware Lexer
- All token types correctly identified
- Column position tracking fixed (enum bug resolved)
- Proper state management (StateBlock, StateDirective, etc.)

### ✅ Context-Aware Parser
- Full AST generation
- Program Counter expressions (`*`)
- All directives (.enum, .macro, .function, .pseudocommand, etc.)

### ✅ Semantic Analyzer
- Undefined symbol detection
- Duplicate definition warnings
- Type checking

### ✅ LSP Features
- textDocument/completion
- textDocument/hover
- textDocument/definition
- textDocument/references
- textDocument/documentSymbol
- textDocument/semanticTokens/full
- textDocument/publishDiagnostics

---

## 🐛 Known Issues (Fixed)

### ✅ Fixed in v0.9.8: Enum Highlighting Bug
**Problem:** Numbers in enum blocks showed character-by-character wrong colors.

**Root Cause:** `tryTokenizeMnemonic()` function consumed 3 characters, advanced column by 3, but only restored position (not column) when backtracking.

**Fix:** Added `l.column = startCol` in all three return paths in `tryTokenizeMnemonic()`.

**Test File:** `01-semantic-tokens.asm` validates this fix.

---

## 🔗 Related Files

- [comprehensive-server-test.asm](../../comprehensive-server-test.asm) - Full feature test
- [context_aware_lexer.go](../../internal/lsp/context_aware_lexer.go) - Lexer implementation
- [context_aware_parser.go](../../internal/lsp/context_aware_parser.go) - Parser implementation
- [semantic.go](../../internal/lsp/semantic.go) - Semantic token generation
- [kickass_cl](../../kickass_cl/) - Test client

---

## 🎯 Success Criteria

### v0.9.0 Baseline Requirements

✅ **All 7 test files pass with 0 unexpected errors**
✅ **Semantic token highlighting is correct everywhere**
✅ **All LSP features work as expected**
✅ **No critical bugs**
✅ **Test client successfully validates all features**

### Definition of "PASS"

A test passes when:
1. Parser successfully creates AST
2. Semantic analyzer validates correctly
3. Expected diagnostics are generated
4. No unexpected errors occur
5. LSP features (hover, completion, etc.) work correctly
6. Semantic tokens are correctly assigned

---

## 📝 Notes

- This is a **stable baseline** - all tests should pass
- Focus is on **current working features**, not new development
- Semantic token test is **critical** after recent bug fix
- Tests use the **new kickass_cl test client**
- HTML reports provide excellent visualization

---

## 🚀 Next Steps

1. **Create test files** - 7 .asm files covering all features
2. **Create baseline-suite.json** - JSON test suite definition
3. **Run initial tests** - Verify all pass
4. **Generate HTML report** - Baseline documentation
5. **Use as regression suite** - For future development

This baseline ensures we don't break working features when adding new ones! 🎉
