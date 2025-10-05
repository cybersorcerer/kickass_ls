# Kick Assembler LSP Regression Test Suite

**Version:** 1.0.0
**Purpose:** Prevent feature regressions during development
**Status:** All tests should **PASS** ✅

---

## 📋 Overview

This regression test suite ensures that **all working features from v0.9.6** continue to function correctly during ongoing development. Every test in this suite must pass before releasing a new version.

### Why Regression Testing?

- 🛡️ **Prevents Breaking Changes** - Catches bugs before they reach production
- ✅ **Validates Existing Features** - Ensures proven functionality stays intact
- 🔄 **Safe Refactoring** - Allows confident code improvements
- 📊 **Quality Assurance** - Maintains high code quality standards

---

## 🎯 Test Coverage

### Test Files

| # | Test File | Feature Tested | Lines | Expected Result |
|---|-----------|---------------|-------|-----------------|
| 01 | `01-basic-mnemonics.asm` | All 6510 standard opcodes | 80 | ✅ 0 errors |
| 02 | `02-addressing-modes.asm` | All 13 addressing modes | 60 | ✅ 0 errors |
| 03 | `03-number-literals.asm` | Hex/Dec/Bin/Char literals | 40 | ✅ 0 errors |
| 04 | `04-data-directives.asm` | `.byte`, `.word`, `.text`, `.fill` | 50 | ✅ 0 errors |
| 05 | `05-constants-variables.asm` | `.const`, `.var` | 45 | ✅ 0 errors |
| 06 | `06-labels-symbols.asm` | Label definition & usage | 55 | ✅ 0 errors |
| 07 | `07-c64-memory-map.asm` | C64 memory recognition | 60 | ✅ 0 errors |
| 08 | `08-expressions.asm` | Operators & expressions | 50 | ✅ 0 errors |
| 09 | `09-comments.asm` | Line & block comments | 45 | ✅ 0 errors |
| 10 | `10-illegal-opcodes.asm` | Illegal opcode recognition | 40 | ✅ 18+ warnings |

**Total:** 10 test files, 525 lines of test code

---

## 🧪 Test Categories

### 1. Core Language Features ✅

#### 01: Basic Mnemonics
**Coverage:** All 6510 standard opcodes

- Load/Store: `LDA`, `LDX`, `LDY`, `STA`, `STX`, `STY`
- Transfer: `TAX`, `TAY`, `TXA`, `TYA`, `TXS`, `TSX`
- Stack: `PHA`, `PLA`, `PHP`, `PLP`
- Arithmetic: `ADC`, `SBC`, `INC`, `DEC`, `INX`, `INY`, `DEX`, `DEY`
- Logic: `AND`, `ORA`, `EOR`, `BIT`
- Shifts: `ASL`, `LSR`, `ROL`, `ROR`
- Branches: `BEQ`, `BNE`, `BCC`, `BCS`, `BPL`, `BMI`, `BVC`, `BVS`
- Jump/Call: `JMP`, `JSR`, `RTS`, `RTI`
- Flags: `CLC`, `SEC`, `CLI`, `SEI`, `CLD`, `SED`, `CLV`
- Compare: `CMP`, `CPX`, `CPY`
- Misc: `NOP`, `BRK`

**Expected:** 0 errors, 0 warnings

---

#### 02: Addressing Modes
**Coverage:** All 13 6510 addressing modes

1. ✅ **Immediate:** `lda #$00`
2. ✅ **Zero Page:** `lda $80`
3. ✅ **Zero Page,X:** `lda $80, x`
4. ✅ **Zero Page,Y:** `ldx $80, y`
5. ✅ **Absolute:** `lda $d020`
6. ✅ **Absolute,X:** `lda $1000, x`
7. ✅ **Absolute,Y:** `lda $1000, y`
8. ✅ **Indexed Indirect:** `lda ($80, x)` - **Critical Fix!**
9. ✅ **Indirect Indexed:** `lda ($80), y`
10. ✅ **Indirect:** `jmp ($0310)`
11. ✅ **Accumulator:** `asl a`
12. ✅ **Implied:** `nop`, `tax`
13. ✅ **Relative:** `beq label`

**Expected:** 0 errors, 0 warnings

**Note:** Indexed Indirect was fixed in v0.9.6 - this test ensures it stays fixed!

---

#### 03: Number Literals
**Coverage:** All number formats

- **Hexadecimal:** `$00`, `$FF`, `$D020`, `$FFFF`
- **Decimal:** `0`, `10`, `255`, `65535`
- **Binary:** `%00000000`, `%11111111`, `%10101010`
- **Character:** `'A'`, `'Z'`, `'0'`, `' '`
- **Mixed Expressions:** `$10 + 5`, `%11110000 | $0f`

**Expected:** 0 errors, 0 warnings

---

### 2. Directives & Data ✅

#### 04: Data Directives
**Coverage:** `.byte`, `.word`, `.text`, `.fill`, `.align`

- `.byte` - single and multiple values
- `.word` - 16-bit values
- `.text` - string literals
- `.fill` - memory fill
- `.align` - alignment

**Expected:** 0 errors, 0 warnings

---

#### 05: Constants & Variables
**Coverage:** `.const`, `.var`

- **Constants:** Immutable values
- **Variables:** Mutable in Kick Assembler
- **Expressions:** `SPRITE_BASE = SCREEN + $3f8`

**Expected:** 0 errors, 0 warnings

---

#### 06: Labels & Symbols
**Coverage:** Label definition and usage

- Basic labels
- Label assignment: `.label data_ptr = $fb`
- Subroutine labels
- Forward references
- Expression usage: `INIT_SIZE = clear_screen - init`

**Expected:** 0 errors, 0 warnings

---

### 3. C64 Specific Features ✅

#### 07: C64 Memory Map
**Coverage:** Complete C64 memory address recognition

**VIC-II ($D000-$D3FF):**
- Sprite registers, control, colors

**SID ($D400-$D7FF):**
- Voice, filter, volume registers

**CIA #1 & #2 ($DC00-$DDFF):**
- Data ports, timers, interrupts

**Special:**
- Color RAM ($D800-$DBFF)
- Zero Page ($00-$FF)
- Kernal vectors ($0314-$0315)

**Expected:** 0 errors, 0 warnings

---

### 4. Advanced Features ✅

#### 08: Expressions & Operators
**Coverage:** Expression evaluation

**Operators:**
- Arithmetic: `+`, `-`, `*`, `/`
- Bitwise: `&`, `|`, `^`, `~`
- Shift: `<<`, `>>`
- Byte extraction: `<`, `>`

**Expected:** 0 errors, 0 warnings

---

#### 09: Comments
**Coverage:** All comment styles

- Line comments: `//`, `;`
- Block comments: `/* */`
- Multi-line block comments
- End-of-line comments
- Commented code

**Expected:** 0 errors, 0 warnings

---

#### 10: Illegal Opcodes
**Coverage:** Undocumented 6510 opcodes

**Illegal Opcodes:**
- `LAX` - Load A and X
- `SAX` - Store A AND X
- `DCP` - Decrement and Compare
- `ISC` - Increment and Subtract
- `SLO` - Shift Left and OR
- `RLA` - Rotate Left and AND
- `SRE` - Shift Right and EOR
- `RRA` - Rotate Right and Add

**Expected:** 0 errors, 18+ warnings (intended)

**Note:** Illegal opcodes should be recognized but warned about.

---

## 🧪 LSP Feature Tests

### Completion Tests (3 tests)

1. **Mnemonic Suggestions**
   - Should suggest 50+ mnemonics
   - Must include: `lda`, `ldx`, `ldy`, `sta`

2. **Memory Address Suggestions**
   - Should suggest C64 addresses
   - Must include: `$d020`, `$d021`

### Hover Tests (2 tests)

1. **Mnemonic Documentation**
   - Hovering over `LDA` shows documentation

2. **Memory Address Info**
   - Hovering over `$d020` shows "Border Color"

### Symbol Tests (2 tests)

1. **Label Extraction**
   - Must find: `start`, `init`, `clear_screen`, `main`, `end`

2. **Constants Extraction**
   - Must find: `SCREEN`, `BORDER`

### Performance & Memory Tests (2 tests)

1. **Large File Performance**
   - Completion should respond quickly

2. **Memory Leak Detection**
   - 100 operations (50 completion + 50 hover)
   - No memory leaks

---

## 🚀 Running the Tests

### Quick Test (Single File)

```bash
cd kickass_cl
./kickass_cl -server ../kickass_ls ../test-cases/regression-test/01-basic-mnemonics.asm
```

### Full Regression Suite

```bash
cd kickass_cl
./kickass_cl -suite ../test-cases/regression-test/regression-suite.json -verbose
```

### CI/CD Integration

```bash
cd kickass_cl
./kickass_cl -suite ../test-cases/regression-test/regression-suite.json -output regression-results.json

if [ $? -eq 0 ]; then
  echo "✅ All regression tests passed!"
else
  echo "❌ Regression tests failed!"
  exit 1
fi
```

---

## 📊 Expected Results

### All Tests Must Pass ✅

| Test Category | Tests | Expected Errors | Expected Warnings | Status |
|--------------|-------|-----------------|-------------------|--------|
| Mnemonics | 1 | 0 | 0 | ✅ PASS |
| Addressing Modes | 1 | 0 | 0 | ✅ PASS |
| Number Literals | 1 | 0 | 0 | ✅ PASS |
| Data Directives | 1 | 0 | 0 | ✅ PASS |
| Constants/Variables | 1 | 0 | 0 | ✅ PASS |
| Labels/Symbols | 1 | 0 | 0 | ✅ PASS |
| C64 Memory Map | 1 | 0 | 0 | ✅ PASS |
| Expressions | 1 | 0 | 0 | ✅ PASS |
| Comments | 1 | 0 | 0 | ✅ PASS |
| Illegal Opcodes | 1 | 0 | 18+ | ✅ PASS |
| **LSP Features** | | | | |
| Completion | 2 | 0 | 0 | ✅ PASS |
| Hover | 2 | 0 | 0 | ✅ PASS |
| Symbols | 2 | 0 | 0 | ✅ PASS |
| Performance | 1 | 0 | 0 | ✅ PASS |
| Memory | 1 | 0 | 0 | ✅ PASS |
| **TOTAL** | **18** | **0** | **18+** | **✅ PASS** |

---

## 🔄 Development Workflow

### Before Making Changes

```bash
# Run baseline regression test
cd kickass_cl
./kickass_cl -suite ../test-cases/regression-test/regression-suite.json
```

**Expected:** All tests PASS ✅

### After Making Changes

```bash
# Run regression test again
./kickass_cl -suite ../test-cases/regression-test/regression-suite.json
```

**Required:** All tests must still PASS ✅

**If ANY test fails:**
1. ❌ **DO NOT COMMIT** - you've broken existing functionality
2. 🔍 Fix the regression
3. ✅ Re-run tests until all pass
4. ✅ Then commit

---

## 🛡️ Regression Prevention Rules

### Golden Rules

1. **Run regressions BEFORE every commit**
   - Prevents breaking changes from entering codebase

2. **All regression tests must pass**
   - No exceptions, no excuses

3. **New features must not break old features**
   - Add new tests, don't break existing ones

4. **Document any intentional changes**
   - If expected results change, update test suite

### Pre-Commit Checklist

- [ ] Run regression suite: `./kickass_cl -suite ../test-cases/regression-test/regression-suite.json`
- [ ] All tests pass: **18/18 PASS**
- [ ] No new errors introduced
- [ ] No unexpected warnings
- [ ] LSP features still work (completion, hover, symbols)
- [ ] Performance acceptable (< 1s for completion)

---

## 📈 Maintenance

### Adding New Regression Tests

When a bug is fixed or feature is added:

1. **Create test file** in `regression-test/`
2. **Add test case** to `regression-suite.json`
3. **Verify it passes** with current implementation
4. **Document** what it tests

### Updating Expected Results

If behavior intentionally changes:

1. **Update** `expected` values in `regression-suite.json`
2. **Document why** in commit message
3. **Verify** all other tests still pass

---

## 🔗 Related Documentation

- [v0.9.6 Baseline](../0.9.6-baseline/README.md) - Original test suite (all features working)
- [v0.9.7 Baseline](../0.9.7-baseline/README.md) - New directive features (in development)
- [Directive Parsing Status](../../directive-parsing-status.md) - Implementation plan
- [Test Client README](../../kickass_cl/README.md) - Test runner documentation

---

## 🎯 Success Criteria

### Definition of Success

✅ **All 18 tests PASS**
✅ **0 errors** (except illegal opcodes test)
✅ **18+ warnings** (illegal opcodes only)
✅ **LSP features functional** (completion, hover, symbols)
✅ **Performance acceptable** (< 1s response times)
✅ **No memory leaks** (100 operations pass)

### Failure Conditions

❌ **ANY test fails**
❌ **Unexpected errors**
❌ **Missing completions**
❌ **Broken hover/symbols**
❌ **Performance degradation**
❌ **Memory leaks detected**

**If regression suite fails → DO NOT RELEASE**

---

## 📝 Notes

- This suite tests **only proven, working features**
- Tests are based on **v0.9.6 baseline** (10/10 passing)
- **Context-Aware Parser** is used (no legacy parser)
- Tests run on **real LSP server** (no mocks)
- Suite is **comprehensive but fast** (~5-10 seconds)

---

## 🚨 Critical Tests

### Most Important Tests (Cannot Fail!)

1. ✅ **Indexed Indirect** (`02-addressing-modes.asm:47`)
   - This was a major bug fix in v0.9.6
   - `lda ($80, x)` must parse correctly

2. ✅ **C64 Memory Map** (`07-c64-memory-map.asm`)
   - Core feature - all addresses must be recognized

3. ✅ **Basic Mnemonics** (`01-basic-mnemonics.asm`)
   - Foundation - all opcodes must work

4. ✅ **Completion** (Test 11, 12)
   - Primary LSP feature - must be fast and accurate

---

## 🎉 Conclusion

This regression test suite is your **safety net** during development. Run it often, trust it completely, and **never commit** when it fails.

**Happy Coding!** 🚀
