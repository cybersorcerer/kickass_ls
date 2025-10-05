# Kick Assembler LSP v0.9.7 Baseline Test Suite

**Version:** 0.9.7
**Date:** 2025-10-04
**Focus:** Directive Parameter Parsing & Program Counter Expressions
**Reference:** [directive-parsing-status.md](../../directive-parsing-status.md)

---

## 📋 Test Suite Overview

This baseline test suite focuses on **Issue #3** (Directive Parameter Parsing) and **Issue #4** (Program Counter Expressions) identified in version 0.9.6.

### Test Coverage

| Test File | Focus Area | Directive(s) Tested | Status |
|-----------|-----------|---------------------|--------|
| `01-directive-encoding.asm` | String parameters | `.encoding "string"` | 🔴 TODO |
| `02-directive-define.asm` | Symbol-only | `.define`, `.ifdef`, `.ifndef`, `.undef` | 🔴 TODO |
| `03-directive-function.asm` | Parameter lists | `.function name(params) { .return }` | 🔴 TODO |
| `04-directive-macro.asm` | Parameter lists | `.macro name(params) { }` | 🔴 TODO |
| `05-directive-namespace.asm` | Scope management | `.namespace name { }` | 🔴 TODO |
| `06-directive-pseudocommand.asm` | Colon syntax | `.pseudocommand name arg : arg { }` | 🔴 TODO |
| `07-directive-enum.asm` | Member values | `.enum name { MEMBER = value }` | 🔴 TODO |
| `08-directive-import.asm` | Keyword + string | `.import source "file.asm"` | 🔴 TODO |
| `09-pc-expressions.asm` | PC expressions | `* = value`, `beq *+5`, `.label = *` | 🔴 TODO |
| `10-mixed-directives.asm` | Integration | All features combined | 🔴 TODO |

**Total:** 10 test files
**Expected Pass:** 0/10 (baseline for implementation)

---

## 🎯 Test Objectives

### Phase 1: Directive Parameter Parsing (Issue #3)

#### 1. `.encoding "string"` - String Parameter
**File:** `01-directive-encoding.asm`

**Test Cases:**
- ✅ Basic encoding directive with string parameter
- ✅ Multiple encoding types (petscii, screencode, ascii)
- ✅ Invalid encoding name (should warn)

**Expected Behavior:**
- Parser successfully extracts string parameter
- Semantic analyzer validates known encoding types
- Warning for unknown encoding types

---

#### 2. `.define symbol` - Symbol-Only Directive
**File:** `02-directive-define.asm`

**Test Cases:**
- ✅ Basic `.define` without value
- ✅ Conditional compilation (`.ifdef`, `.ifndef`)
- ✅ `.undef` directive
- ✅ Redefinition warning

**Expected Behavior:**
- Parser handles directive without value
- Symbol table tracking for defined symbols
- Warning on redefinition

---

#### 3. `.function name(params)` - Parameter Lists
**File:** `03-directive-function.asm`

**Test Cases:**
- ✅ Function with single parameter
- ✅ Function with multiple parameters
- ✅ Function with no parameters
- ✅ Complex expressions in function body
- ✅ Missing `.return` statement (should warn)
- ✅ Unused parameter (should hint)

**Expected Behavior:**
- Parser extracts parameter list correctly
- Parameter count validation on function calls
- Return statement validation
- Unused parameter detection

---

#### 4. `.macro name(params)` - Parameter Lists
**File:** `04-directive-macro.asm`

**Test Cases:**
- ✅ Macro with no parameters
- ✅ Macro with single/multiple parameters
- ✅ Macro invocation
- ✅ Parameter count mismatch (should error)
- ✅ Unused parameter (should hint)

**Expected Behavior:**
- Parser extracts parameter list
- Macro expansion with parameter substitution
- Argument count validation

---

#### 5. `.namespace name { }` - Scope Management
**File:** `05-directive-namespace.asm`

**Test Cases:**
- ✅ Basic namespace declaration
- ✅ Multiple namespaces
- ✅ Nested namespaces
- ✅ Qualified symbol access (`namespace.symbol`)
- ✅ Duplicate namespace warning

**Expected Behavior:**
- Hierarchical scope creation
- Qualified name resolution
- Duplicate detection

---

#### 6. `.pseudocommand name arg : arg` - Colon Syntax
**File:** `06-directive-pseudocommand.asm`

**Test Cases:**
- ✅ Colon-separated parameter syntax
- ✅ Multiple parameters with colons
- ✅ Pseudocommand invocation
- ✅ Missing parameters (should error)

**Expected Behavior:**
- Parser handles colon-separated parameters
- Parameter validation on invocation

---

#### 7. `.enum name { MEMBER = value }` - Member Values
**File:** `07-directive-enum.asm`

**Test Cases:**
- ✅ Enum with explicit values
- ✅ Enum with auto-increment values
- ✅ Mixed explicit/auto values
- ✅ Duplicate value detection (should warn)
- ✅ Expression values

**Expected Behavior:**
- Parser extracts enum members and values
- Duplicate value detection
- Expression evaluation in enum values

---

#### 8. `.import source "file.asm"` - Keyword + String
**File:** `08-directive-import.asm`

**Test Cases:**
- ✅ Import with `source` keyword
- ✅ Import with `binary` keyword
- ✅ Import with `c64` keyword
- ✅ Conditional import
- ⚠️ File existence check (should warn if missing)
- ❌ Invalid syntax (no quotes - should error)

**Expected Behavior:**
- Parser extracts keyword and string parameter
- File existence validation (optional)

---

### Phase 2: Program Counter Expressions (Issue #4)

#### 9. `*` (Program Counter) Expressions
**File:** `09-pc-expressions.asm`

**Test Cases:**
- ✅ `.label = *` - PC as value
- ✅ `beq *+5` - Forward relative branch
- ✅ `bne *-5` - Backward relative branch
- ✅ `* - start` - PC in expression
- ✅ Alignment with PC
- ✅ PC in data directives (`.byte <*, >*`)

**Expected Behavior:**
- Parser recognizes `*` as PC reference in expression context
- Semantic analyzer resolves PC value
- Relative branch distance calculation

---

### Phase 3: Integration Testing

#### 10. Mixed Directive Features
**File:** `10-mixed-directives.asm`

**Test Cases:**
- ✅ Combination of encoding, namespaces, functions, enums
- ✅ Pseudocommands using namespace symbols
- ✅ Functions with enum returns
- ✅ PC expressions with all features

**Expected Behavior:**
- All features work together correctly
- No conflicts between features

---

## 🧪 Running the Tests

### Quick Test (Single File)

```bash
cd kickass_cl
./kickass_cl -server ../kickass_ls ../test-cases/0.9.7-baseline/01-directive-encoding.asm
```

### Full Test Suite

```bash
cd kickass_cl
./kickass_cl -suite ../test-cases/0.9.7-baseline/baseline-suite.json -verbose
```

### With Output

```bash
cd kickass_cl
./kickass_cl -suite ../test-cases/0.9.7-baseline/baseline-suite.json -output results-0.9.7.json
```

---

## 📊 Expected Results (Current Implementation)

### Baseline Expectations for v0.9.7 Development

| Category | Expected Errors | Expected Warnings | Expected Pass |
|----------|----------------|-------------------|---------------|
| `.encoding` | 0-1 | 1 | Partial |
| `.define` | 0 | 1 | Partial |
| `.function` | 0 | 2 | Partial |
| `.macro` | 1 | 0 | Partial |
| `.namespace` | 0 | 1 | Partial |
| `.pseudocommand` | 1 | 0 | Partial |
| `.enum` | 0 | 1 | Partial |
| `.import` | 0 | 0 | Partial |
| PC expressions | 0 | 0 | No |
| Mixed features | 0 | 0 | No |

**Initial Status:** Most tests will **FAIL** - this is expected!
**Goal:** Implement features until all tests **PASS**

---

## 🔄 Development Workflow

### 1. Pick a Feature
Choose one directive from the test suite (e.g., `.encoding`)

### 2. Run Baseline Test
```bash
./kickass_cl -server ../kickass_ls ../test-cases/0.9.7-baseline/01-directive-encoding.asm -verbose
```

### 3. Implement Parser Changes
Edit [context_aware_parser.go](../../internal/lsp/context_aware_parser.go):
```go
case ".encoding":
    return p.parseEncodingDirective()
```

### 4. Implement Semantic Analysis
Edit [analyze.go](../../internal/lsp/analyze.go):
```go
func (a *SemanticAnalyzer) validateEncodingDirective(node *DirectiveStatement) {
    // Validate encoding type
}
```

### 5. Re-test
```bash
./kickass_cl -server ../kickass_ls ../test-cases/0.9.7-baseline/01-directive-encoding.asm
```

### 6. Iterate
Repeat until test passes

### 7. Run Full Suite
```bash
./kickass_cl -suite ../test-cases/0.9.7-baseline/baseline-suite.json
```

---

## 📈 Progress Tracking

### Implementation Checklist

- [ ] `.encoding "string"` - String parameter parsing
- [ ] `.define symbol` - Symbol-only directive
- [ ] `.function name(params)` - Parameter list parsing
- [ ] `.macro name(params)` - Parameter list parsing
- [ ] `.namespace name { }` - Scope management
- [ ] `.pseudocommand name arg : arg` - Colon-separated parameters
- [ ] `.enum name { MEMBER = value }` - Enum member parsing
- [ ] `.import source "file"` - Keyword + string parsing
- [ ] `*` in expressions - PC reference
- [ ] `*±offset` - PC relative expressions
- [ ] Integration testing - All features combined

### Diagnostics Implementation Checklist

- [ ] `.encoding` unknown encoding warning
- [ ] `.define` redefinition warning
- [ ] `.function` missing return warning
- [ ] `.function` unused parameter hint
- [ ] `.macro` parameter count error
- [ ] `.macro` unused parameter hint
- [ ] `.namespace` duplicate warning
- [ ] `.pseudocommand` parameter mismatch error
- [ ] `.enum` duplicate value warning
- [ ] `.import` file not found warning

---

## 🔗 Related Documentation

- [Directive Parsing Status](../../directive-parsing-status.md) - Detailed analysis and implementation plan
- [v0.9.6 Baseline](../0.9.6-baseline/README.md) - Previous baseline (all passing)
- [Test Client README](../../kickass_cl/README.md) - Test runner documentation
- [Context-Aware Parser](../../internal/lsp/context_aware_parser.go) - Parser implementation
- [Semantic Analyzer](../../internal/lsp/analyze.go) - Diagnostics implementation

---

## 🎯 Success Criteria

### v0.9.7 Release Requirements

✅ **All 10 test files pass with 0 errors**
✅ **All expected warnings are generated**
✅ **All expected hints are generated**
✅ **No regressions in v0.9.6 baseline tests**
✅ **Documentation updated**

### Definition of "PASS"

A test passes when:
1. Parser successfully creates AST nodes for all directives
2. Semantic analyzer validates according to spec
3. Expected diagnostics are generated (errors, warnings, hints)
4. No unexpected errors occur
5. LSP features (hover, completion, symbols) work correctly

---

## 📝 Notes

- Tests are designed to **fail initially** - this is the baseline
- Each test file focuses on **one feature** for clarity
- `10-mixed-directives.asm` tests **integration** of all features
- Tests follow **0.9.6 baseline** format for consistency
- All tests use **Context-Aware Parser** (no legacy parser)

---

## 🚀 Next Steps

1. **Run initial baseline** - establish current failure points
2. **Implement `.encoding`** - simplest feature (string parameter)
3. **Implement `.define`** - symbol-only directive
4. **Implement `.function`** - parameter lists
5. **Continue** through remaining directives
6. **Implement PC expressions** - `*` in expressions
7. **Run final suite** - verify all tests pass

Good luck! 🎉
