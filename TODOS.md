# 6510 LSP Server - TODO List

## ✅ Completed Tasks

### Critical Issues Fixed
- **✅ CRITICAL: .for loop parsing breaks all subsequent diagnostics** - Fixed and confirmed by user in Neovim
- **✅ CRITICAL: Multiple analysis passes causing duplicate diagnostics** - Fixed and confirmed by user
- **✅ Fix missing Zero Page optimization hints** - Fixed and confirmed working by user
- **✅ Fix Range validation for .const directives** - Fixed and confirmed working by user
- **✅ Fix illegal opcode detection - dcp not generating warning** - DCP illegal opcode detection works

## 🚨 Critical Current Issues

### Parser Architecture Problems
- **🚨 CRITICAL: Parser fails to create DirectiveStatements for .byte/.word with comma-separated values**
  - **Issue**: Lines like `.byte $00, $FF, $100` and `.word $0000, $FFFF, $10000` are tokenized correctly but don't create AST nodes
  - **Evidence**: Parser processes `DIRECTIVE_KICK_DATA '.byte'` and `NUMBER_HEX` tokens but never calls `processDirective`
  - **Impact**: Range validation for multi-value data directives completely broken
  - **Priority**: CRITICAL - blocks range validation feature

### Blocked by Parser Issues
- **⏸️ Fix Range validation for .byte and .word directives** - BLOCKED by parser issue above
  - **Workaround**: Implement token-level range validation bypass

## 📋 Pending Tasks

### Parser & Language Features
- **Fix missing Dead Code detection for .if(0) blocks**
  - Test case exists but detection not implemented
- **Fix .for loop brace matching - parser expects '{' at end of file**
  - Parser has issues with .for loop termination
- **Fix .for loop number highlighting**
  - Syntax highlighting issues within .for loops

### Code Quality & Maintenance
- **Remove debug logging added during hex token fix**
  - Clean up temporary debug statements
- **Revert broken regression tests that use @comprehensive-test.asm**
  - Restore working regression test suite

### Architecture Improvements
- **ARCHITECTURE: Implement proper context-aware lexer/parser for Kick Assembler**
  - Current parser is too simplistic for complex Kick Assembler directives
  - Need proper grammar for nested expressions, macro calls, complex .for loops
  - Consider implementing proper AST for data directives with multiple values

## 🔧 Implementation Notes

### Range Validation Status
- **Working**: `.const` directives (single values)
- **Broken**: `.byte` and `.word` directives (multiple values)
- **Root Cause**: Parser doesn't create DirectiveStatements for comma-separated data directives

### Test Coverage
- ✅ Zero Page optimization hints - test passes
- ✅ Range validation for .const - test passes
- ❌ Range validation for .byte/.word - test passes but validation doesn't work in practice
- ❌ Dead code detection - no test implementation yet

### User Testing Required
All completed features need user confirmation in Neovim before being marked as truly fixed:
- User rule: "NICHTS als behoben markieren bevor ich nicht in Neovim getestet habe"

## 🚀 Next Steps

1. **Immediate**: Implement workaround for .byte/.word range validation using token-level processing
2. **Short-term**: Fix parser to properly handle comma-separated data directives
3. **Medium-term**: Implement dead code detection for .if(0) blocks
4. **Long-term**: Complete parser/lexer architecture redesign for full Kick Assembler support