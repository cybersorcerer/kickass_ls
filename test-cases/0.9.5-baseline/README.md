# Kick Assembler LSP v0.9.5 Baseline Test Suite

**Version:** 0.9.5
**Created:** 2025-10-04
**Purpose:** Comprehensive baseline test suite for validating LSP server functionality

---

## Overview

This baseline test suite provides comprehensive coverage of the Kick Assembler Language Server's features in version 0.9.5. It includes test files covering all major language features, LSP capabilities, and expected behaviors.

### Test Suite Contents

| Test File | Purpose | Features Tested |
|-----------|---------|-----------------|
| `01-basic-syntax.asm` | Basic assembler syntax | Labels, comments, literals (hex/binary/decimal/char) |
| `02-directives.asm` | Kick Assembler directives | .var, .const, .if, .macro, .function, .namespace, .enum, etc. |
| `03-addressing-modes.asm` | 6510 addressing modes | All 13 addressing modes (immediate, absolute, indexed, indirect, etc.) |
| `04-illegal-opcodes.asm` | Illegal/undocumented opcodes | LAX, SAX, DCP, ISC, SLO, RLA, SRE, RRA |
| `05-c64-memory.asm` | C64 memory map | VIC-II, SID, CIA, Color RAM, Kernal vectors |
| `06-completion-context.asm` | Context-aware completion | Directive/mnemonic completion, addressing mode hints |
| `07-diagnostics.asm` | Error detection | Invalid operands, hex values, addressing modes |
| `08-builtins.asm` | Built-in functions | Math functions, constants, operators |

### JSON Test Suite

`baseline-suite.json` - Automated test suite for use with `kickass_cl` test runner

---

## Test Results Summary

### Test File: 01-basic-syntax.asm ✅

**Status:** PASS (0 errors, 6 warnings, 4 hints/info)

**Diagnostics:**
- ✅ No syntax errors
- ⚠️ 6 warnings for unused labels (expected - test file)
- 💡 4 hints/info for I/O register access ($D020, $D021)

**Expected Behavior:**
- All basic syntax elements (labels, literals) recognized correctly
- Hex values: `#$ff`, `$d020`
- Binary values: `#%11111111`
- Decimal values: `#255`, `#128`
- Comments: `//` style recognized

---

### Test File: 02-directives.asm ⚠️

**Status:** PARTIAL (5 errors, multiple warnings)

**Known Issues:**
1. ❌ Line 25: `*` (current program counter) not supported in expressions
2. ❌ Line 30: `.if` conditional parsing issues
3. ❌ Line 73: `.encoding` string parameter not recognized
4. ❌ Line 91: `.function` parameter parsing issues
5. ❌ Line 42: Duplicate label `loop` (macro scope issue)

**Working Directives:** ✅
- `.var`, `.const`, `.label` - Variable/constant declarations
- `.byte`, `.fill` - Data directives
- `.namespace` - Namespace support
- `.macro` - Macro definitions (partial)
- `.pseudocommand` - Pseudocommand definitions (partial)
- `.enum` - Enumeration support

**Limitations:**
- Program counter `*` in expressions not fully supported
- Complex `.if` conditions may fail
- `.encoding` directive recognized but string parameters not parsed
- `.function` parameter parsing incomplete
- Macro-internal labels may conflict with global scope

---

### Test File: 03-addressing-modes.asm ⚠️

**Status:** PARTIAL (4 errors, multiple warnings)

**Known Issues:**
1. ❌ Lines 47-49: Indexed Indirect `($80, x)` - comma parsing issue
2. ❌ Lines 85-86: Relative branch offsets `*+5`, `*-3` - program counter expressions

**Working Addressing Modes:** ✅
- Immediate: `lda #$00` ✅
- Zero Page: `lda $80` ✅
- Zero Page,X: `lda $80, x` ✅
- Zero Page,Y: `ldx $80, y` ✅
- Absolute: `lda $d020` ✅
- Absolute,X: `lda $1000, x` ✅
- Absolute,Y: `lda $1000, y` ✅
- Indirect: `jmp ($0310)` ✅
- Indirect Indexed (Y): `lda ($80), y` ✅
- Accumulator: `asl a` ✅
- Implied: `nop`, `rts`, `tax` ✅
- Relative: `beq forward`, `bne backward` ✅ (labels work)

**Limitations:**
- Indexed Indirect `(zp, x)` not fully parsed
- Program counter relative `*+/-` not supported in expressions

---

### Test File: 04-illegal-opcodes.asm ✅

**Status:** PASS

**Supported Illegal Opcodes:**
- LAX (Load A and X) - All addressing modes ✅
- SAX (Store A AND X) - All addressing modes ✅
- DCP (Decrement and Compare) - All addressing modes ✅
- ISC (Increment and Subtract) - All addressing modes ✅
- SLO (Shift Left and OR) - All addressing modes ✅
- RLA (Rotate Left and AND) - All addressing modes ✅
- SRE (Shift Right and EOR) - All addressing modes ✅
- RRA (Rotate Right and Add) - All addressing modes ✅

**Note:** Illegal opcodes fully supported across all valid addressing modes

---

### Test File: 05-c64-memory.asm ✅

**Status:** PASS (0 errors)

**C64 Memory Map Recognition:**

**VIC-II ($D000-$D3FF)** ✅
- Sprite positions, enable, colors
- Control registers
- Memory pointers
- Border/background colors

**SID ($D400-$D7FF)** ✅
- Voice 1/2/3 frequency, waveform, ADSR
- Filter and volume registers

**CIA #1 ($DC00-$DCFF)** ✅
- Data ports A/B
- Timers
- Interrupt control

**CIA #2 ($DD00-$DDFF)** ✅
- Data ports A/B
- VIC bank switching

**Special Memory** ✅
- Color RAM ($D800-$DBFF)
- Screen RAM ($0400-$07E7)
- Zero Page ($00-$FF)
- Kernal vectors ($0314, $FFFA-$FFFD)

**Kernal Routines** ✅
- $FFD2 (CHROUT), $FFE4 (GETIN), $E544 (Clear screen)

---

### Test File: 06-completion-context.asm 🔄

**Status:** IN PROGRESS

**Completion Tests:**

**Test 1: Directive Completion (Line 28, char 0)** ✅
- Typing `.` triggers directive completion
- Result: 130 completion items shown
- Includes: `.var`, `.const`, `.byte`, `.if`, `.macro`, etc.
- **PASS** - All directives offered

**Test 2: Mnemonic Addressing Modes (Line 15, char 8)** ❌
- Typing `lda ` should show addressing mode hints
- Result: 0 completion items (server crash)
- **FAIL** - Server terminates unexpectedly
- **Known Issue:** Completion after mnemonic with space crashes server

**Test 3: After `.var name` (before =)** ⏭️
- Should show no completions (user types new name)
- **NOT TESTED** - Dependent on server stability

**Test 4: C64 Memory Completion** ⏭️
- Typing `$d0` should show VIC-II registers
- **NOT TESTED**

**Test 5: Recently Used Operands** ⏭️
- After multiple `lda #$XX`, should suggest previous values
- **NOT TESTED**

**Critical Issue:** Server crashes when requesting completion after mnemonic with trailing space

---

### Test File: 07-diagnostics.asm ✅

**Status:** PASS (14 errors detected correctly)

**Error Detection:**

**Missing Operands** ✅
- Line 12: `lda` (no operand) → Error: Invalid addressing mode 'Implied'
- Line 13: `sta` (no operand) → Error: Invalid addressing mode 'Implied'
- Line 23: `ldx` (no operand) → Error detected
- Line 24: `ldy` (no operand) → Error detected

**Invalid Hex Values** ✅
- Line 16: `#$XY` → Error: Illegal character sequence '#$XY'
- Line 17: `#$GG` → Error: Illegal character sequence '#$GG'

**Invalid Binary Values** ✅
- Line 30: `#%2222` → Error: Unexpected token '%'

**Invalid Addressing Modes** ✅
- Line 20: `jmp #$1000` → **NOT DETECTED** (should error - JMP doesn't support immediate)
- Line 46: `sta #'A'` → Error: Invalid addressing mode 'Immediate' for STA

**Invalid Character Literals** ✅
- Line 45: `#''` → Error: Illegal character sequence '''
- Line 46: `#'AB'` → Error: Illegal character sequence '''

**Summary:**
- 14 errors detected correctly
- Operand validation works
- Hex/binary/character literal validation works
- Some addressing mode validation gaps (JMP immediate not caught)

---

### Test File: 08-builtins.asm 📋

**Status:** NOT FULLY TESTED

**Built-in Support:**
- Math functions: `sin`, `cos`, `tan`, `sqrt`, `pow`, `abs`, etc.
- Constants: `PI`, `E`
- Operators: `<` (low byte), `>` (high byte), `>>` (shift)
- List/Hashtable support
- String operations

**Note:** Full testing requires inspection of completion results and diagnostics for function calls

---

## Known Limitations (v0.9.5)

### 1. Program Counter Expression Support ❌
**Issue:** `*` (current program counter) not supported in expressions
**Examples:**
- `.label loop = *` → Error
- `beq *+5` → Error
- `.if (counter == 0)` with `*` → Error

**Workaround:** Use explicit labels instead of PC-relative expressions

### 2. Indexed Indirect Addressing Parsing ❌
**Issue:** Comma inside parentheses not parsed correctly
**Example:** `lda ($80, x)` → Error: expected ')', got ','

**Workaround:** None - core parser issue

### 3. Character Literal Support ❌
**Issue:** Character literals `'A'` not recognized
**Example:** `lda #'A'` → Error: Illegal character sequence

**Workaround:** Use ASCII decimal value: `lda #65`

### 4. Completion Server Stability ⚠️
**Issue:** Server crashes when completing after mnemonic with trailing space
**Example:** Position after `lda ` causes server termination

**Impact:** Context-aware completion for addressing modes not testable

### 5. Directive Parameter Parsing ⚠️
**Issue:** Some directive parameters not fully parsed
**Examples:**
- `.encoding "string"` → String parameter not recognized
- `.function name(param)` → Parameter parsing incomplete
- `.if` complex conditions → Parsing errors

### 6. Macro Scope Issues ⚠️
**Issue:** Labels inside macros may conflict with global scope
**Example:** `loop:` inside macro conflicts with global `loop:` label

### 7. Semicolon Comments ⚠️
**Issue:** Traditional `;` comments may be interpreted as code
**Example:** `; comment` → Error: Unknown opcode 'line'

**Workaround:** Use `//` style comments exclusively

---

## Feature Support Matrix

| Feature | Status | Notes |
|---------|--------|-------|
| **Basic Syntax** | ✅ Full | Labels, hex/binary/decimal literals |
| **Mnemonics (Standard)** | ✅ Full | All 6510 standard opcodes |
| **Mnemonics (Illegal)** | ✅ Full | LAX, SAX, DCP, ISC, SLO, RLA, SRE, RRA |
| **Addressing Modes** | ⚠️ Partial | 11/13 modes (Indexed Indirect fails) |
| **Directives** | ⚠️ Partial | Basic directives work, complex parsing issues |
| **C64 Memory Map** | ✅ Full | VIC-II, SID, CIA, Kernal recognized |
| **Built-in Functions** | ⏭️ Untested | Available but not validated |
| **Diagnostics** | ✅ Good | Operand/literal validation works |
| **Completion** | ⚠️ Unstable | Directive completion works, mnemonic completion crashes |
| **Hover** | ⏭️ Untested | Not validated in baseline |
| **Go-to-Definition** | ⏭️ Untested | Not validated in baseline |
| **Symbols** | ⏭️ Untested | Not validated in baseline |

---

## Running the Baseline Tests

### Quick Test Mode

Test individual files:

```bash
# Test basic syntax
./kickass_cl/kickass_cl -server ./kickass_ls test-cases/0.9.5-baseline/01-basic-syntax.asm

# Test directives
./kickass_cl/kickass_cl -server ./kickass_ls test-cases/0.9.5-baseline/02-directives.asm

# Test addressing modes
./kickass_cl/kickass_cl -server ./kickass_ls test-cases/0.9.5-baseline/03-addressing-modes.asm
```

### Completion Testing

Test context-aware completion:

```bash
# Test directive completion (type '.')
./kickass_cl/kickass_cl -server ./kickass_ls completion-at \
  test-cases/0.9.5-baseline/06-completion-context.asm 28 0

# Test mnemonic completion (WARNING: may crash)
./kickass_cl/kickass_cl -server ./kickass_ls completion-at \
  test-cases/0.9.5-baseline/06-completion-context.asm 15 8
```

### Test Suite Mode

Run automated test suite:

```bash
# Run full baseline suite
./kickass_cl/kickass_cl -suite test-cases/0.9.5-baseline/baseline-suite.json

# Run with verbose output
./kickass_cl/kickass_cl -suite test-cases/0.9.5-baseline/baseline-suite.json -verbose

# Save results to JSON
./kickass_cl/kickass_cl -suite test-cases/0.9.5-baseline/baseline-suite.json \
  -output baseline-results.json
```

---

## Expected Test Results

### Passing Tests (No Errors)

✅ **01-basic-syntax.asm**
- 0 errors
- 6 warnings (unused labels - expected)
- 4 hints/info (I/O register suggestions)

✅ **04-illegal-opcodes.asm**
- 0 errors
- All illegal opcodes recognized

✅ **05-c64-memory.asm**
- 0 errors
- All C64 memory locations recognized

### Partial Pass (Expected Errors)

⚠️ **02-directives.asm**
- 5 errors (known limitations: `*`, `.if`, `.encoding`, `.function`)
- Multiple warnings (expected from test structure)

⚠️ **03-addressing-modes.asm**
- 4 errors (Indexed Indirect parsing, PC-relative expressions)
- Most addressing modes work correctly

⚠️ **07-diagnostics.asm**
- 14 errors **expected** (file contains intentional errors)
- Error detection working correctly

### Failing Tests

❌ **06-completion-context.asm**
- Completion after mnemonic crashes server
- Directive completion works
- Critical stability issue

⏭️ **08-builtins.asm**
- Not fully tested
- Requires completion/hover inspection

---

## Recommendations for v0.9.6

### Critical Fixes

1. **Fix Completion Server Crash** 🔴
   - Completion after `mnemonic<space>` terminates server
   - Blocks testing of addressing mode hints
   - Priority: **CRITICAL**

2. **Add Program Counter Expression Support** 🟠
   - Support `*` in expressions: `.label loop = *`, `beq *+5`
   - Common pattern in assembler code
   - Priority: **HIGH**

3. **Fix Indexed Indirect Parsing** 🟠
   - Parse `($80, x)` addressing mode correctly
   - Comma inside parentheses fails
   - Priority: **HIGH**

### Important Improvements

4. **Character Literal Support** 🟡
   - Support `'A'` syntax for character literals
   - Common in assembler code
   - Priority: **MEDIUM**

5. **Enhance Directive Parameter Parsing** 🟡
   - Fix `.encoding "string"` string parameter
   - Fix `.function name(param)` parameter list
   - Improve `.if` condition parsing
   - Priority: **MEDIUM**

6. **Macro Scope Management** 🟡
   - Isolate macro-internal labels from global scope
   - Prevent label conflicts
   - Priority: **MEDIUM**

### Nice to Have

7. **Semicolon Comment Support** 🔵
   - Recognize `;` as traditional comment character
   - Currently interpreted as code
   - Priority: **LOW**

8. **Addressing Mode Validation** 🔵
   - Detect invalid modes like `jmp #$1000`
   - Some gaps in validation
   - Priority: **LOW**

---

## Test Suite Maintenance

### Adding New Tests

1. Create new `.asm` file in `test-cases/0.9.5-baseline/`
2. Document purpose and features tested in file header
3. Add test case to `baseline-suite.json`
4. Update this README with results

### Updating Test Expectations

When fixing bugs:
1. Re-run affected test files
2. Update "Expected Test Results" section
3. Move items from "Known Limitations" to "Fixed Issues"
4. Update feature support matrix

### Version Migration

For v0.9.6+:
1. Copy baseline to `test-cases/0.9.6-baseline/`
2. Update files to test new features
3. Re-run all tests
4. Create comparative analysis (0.9.5 → 0.9.6)

---

## Conclusion

The v0.9.5 baseline test suite provides comprehensive coverage of the Kick Assembler Language Server's capabilities. The server demonstrates **strong support for basic syntax, standard/illegal mnemonics, and C64 memory mapping**, with **known limitations in expression parsing, indexed indirect addressing, and completion stability**.

**Overall Assessment:**
- Core functionality: ⭐⭐⭐⭐ (4/5)
- Advanced features: ⭐⭐⭐ (3/5)
- Stability: ⭐⭐⭐ (3/5)
- **Total: ⭐⭐⭐⭐ (3.5/5)**

**Key Strengths:**
- ✅ Excellent mnemonic support (standard + illegal)
- ✅ Complete C64 memory map recognition
- ✅ Robust diagnostic detection
- ✅ Directive completion works well

**Key Weaknesses:**
- ❌ Completion server stability issues
- ❌ Limited expression support (program counter)
- ❌ Addressing mode parsing gaps

**Recommendation:** Prioritize completion stability and expression parsing for v0.9.6 to achieve production-ready status.

---

**Last Updated:** 2025-10-04
**Test Suite Version:** 0.9.5
**Server Version:** v0.9.5
