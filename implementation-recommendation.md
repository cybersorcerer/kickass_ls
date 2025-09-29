# Implementation Recommendations

## Strategic Implementation Recommendations

### Critical Lessons Learned Today:
1. **Never break existing functionality** - every change must be backwards compatible
2. **Minimal, targeted changes** - avoid sweeping modifications
3. **Test immediately** - validate each small change before proceeding
4. **Systematic analysis** - understand the problem completely before coding

### Implementation Priority & Risk Assessment:

**PHASE 1: Low-Risk Extensions (Do First)**
- **kickass-directive-plan.md**: Add missing illegal opcodes and built-in functions
  - Risk: LOW - These are pure additions, won't break existing code
  - Approach: Add new token types and parser cases incrementally
  - Test: Each new directive individually

**PHASE 2: Medium-Risk Enhancements (Do Carefully)**
- **semantic-improvements-plan.md**: Address calculation and branch validation
  - Risk: MEDIUM - Could affect existing symbol resolution
  - Approach: Add new analysis passes without changing existing ones
  - Critical: Maintain current symbol table structure

**PHASE 3: High-Risk Changes (Do Last, If Ever)**
- **formatting-plan.md**: Complete autoformat system
  - Risk: HIGH - Complex token metadata changes
  - Concern: Could break lexer/parser integration
  - Recommendation: Prototype separately first

**PHASE 4: Architecture Changes (Extreme Caution)**
- **refactor-plan.md**: Layer separation redesign
  - Risk: EXTREME - Would require rewriting everything
  - Recommendation: Only consider after all features proven stable
  - Alternative: Gradual refactoring of individual components

### Updated Analysis: Illegal Opcodes Status

**✅ ILLEGAL OPCODES: Already Comprehensive**
After detailed analysis, the illegal opcodes implementation is more complete than expected:

- **mnemonic.json**: Contains 45 illegal opcodes with full addressing modes and descriptions
- **Lexer**: Correctly loads illegal opcodes and categorizes as TOKEN_MNEMONIC_ILL
- **Hover provider**: Shows warnings with "⚠️ **ILLEGAL OPCODE**" for type "Illegal" opcodes
- **analyze.go**: Has redundant hardcoded map with only 8 opcodes (outdated subset)

**Issues Found:**
1. **JAM opcode missing type field** in mnemonic.json (should be type: "Illegal")
2. **Redundant warning system** in analyze.go with outdated subset
3. **SBC listed incorrectly** in hardcoded map (SBC is legitimate, not illegal)

### Revised Next Steps:

1. **✅ Built-in functions** - Already implemented and working
2. **Fix illegal opcodes issues**:
   - Add missing "type": "Illegal" to JAM opcode in mnemonic.json
   - Remove/update redundant hardcoded illegal opcodes map in analyze.go
   - Remove incorrect SBC entry from illegal opcodes list
3. **Implement anonymous labels** - new token types for !, !+, !-
4. **Add semantic passes** - layer on top of existing symbol resolution
5. **Consider formatting** - only after everything else is rock solid

**Key Finding**: Illegal opcodes are essentially complete. The next logical step is anonymous labels or enhanced semantic analysis, not illegal opcodes.

The key insight from today: **incremental additions > wholesale changes**. Each plan document represents months of work - we must proceed with surgical precision, not broad strokes.