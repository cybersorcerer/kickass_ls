# 🧠 Semantische Analyse - Comprehensive Improvement Plan

## 📊 **Aktuelle Schwächen der bestehenden Analyse**

### **Limitierte Feature-Set:**
- ✅ **Vorhanden**: Symbol-Nutzung, Unused-Warnings, Basic Macro-Validierung
- ❌ **Fehlt**: Address-Calculation, Branch-Distance, Memory-Layout-Awareness
- ❌ **Fehlt**: 6502-spezifische Optimierungen, Dead-Code-Detection
- ❌ **Fehlt**: Cross-Reference-Analysis, Dependency-Tracking

### **Oberflächliche 6502-Expertise:**
- ❌ **Keine PC-Tracking** - Program Counter wird nicht verfolgt
- ❌ **Keine Address-Resolution** - Labels haben keine Adressen
- ❌ **Keine Memory-Map-Awareness** - ROM/RAM/I/O-Bereiche unbekannt
- ❌ **Keine Branch-Distance-Validation** - Critical für 6502!

## 🚀 **Mein Verbesserungsvorschlag - Multi-Pass Enhanced Analyzer**

### **🎯 Kernkonzept: 6-Pass Semantic Analysis**

```
Pass 1: Symbol Collection & Address Calculation
├── PC-Tracking durch alle Statements
├── Label-Adressen berechnen  
├── Directive-Processing (.pc, .byte, .word, etc.)
└── Forward-Reference-Collection

Pass 2: Forward Reference Resolution
├── Alle Symbol-Referenzen auflösen
├── Address-Dependencies berechnen
└── Circular-Dependency-Detection

Pass 3: Usage Analysis (Enhanced)
├── Symbol-Usage-Tracking (existing)  
├── Scope-aware Cross-References
└── Comment-Filtering (existing)

Pass 4: 6502/C64 Specialized Analysis
├── Branch-Distance-Validation (-128 to +127)
├── Zero-Page-Access-Optimization
├── Illegal-Opcode-Warnings
├── Memory-Access-Pattern-Analysis
└── CPU-Flag-Dependency-Tracking

Pass 5: Optimization Hints
├── Dead-Code-Detection
├── Unreachable-Code-Detection  
├── Redundant-Instruction-Analysis
└── Performance-Optimization-Suggestions

Pass 6: Traditional Checks (Enhanced)
├── Unused-Symbol-Detection (existing)
├── Style-Guide-Violations
└── Best-Practice-Recommendations
```

## 💡 **Konkrete Verbesserungen**

### **1. Address-Aware Analysis** ⭐⭐⭐⭐⭐

#### **Problem:**
```assembly
.pc = $1000
start:
    jmp loop    ; Wo ist 'loop'? Welche Distanz?
    nop
loop:
    bne start   ; Branch-Distance ok? 
```

#### **Lösung:**
```go
type AnalysisContext struct {
    CurrentPC        int64              // Track program counter
    DefinedLabels    map[string]*Symbol // Labels mit Adressen
    ForwardRefs      []ForwardReference // Unresolved references
}

func (a *Analyzer) processInstruction(node *InstructionStatement) {
    mnemonic := node.Token.Literal
    length := a.getInstructionLength(mnemonic, node.Operand)
    a.context.CurrentPC += int64(length)
    
    // Branch-distance validation
    if isBranchInstruction(mnemonic) {
        a.validateBranchDistance(node.Operand, node.Token)
    }
}
```

### **2. 6502-Specific Optimizations** ⭐⭐⭐⭐⭐

#### **Zero-Page Optimization:**
```assembly
lda $0080    ; Could be: lda $80 (saves 1 byte, 1 cycle!)
sta $00FF    ; Zero-page - good!
```

#### **Implementation:**
```go
func (a *Analyzer) analyzeZeroPageAccess(mnemonic string, operand Expression) {
    if addr, err := a.evaluateExpression(operand); err == nil {
        if addr >= 0x00 && addr <= 0xFF && a.supportsZeroPageMode(mnemonic) {
            a.addHint("Consider zero-page addressing for $%02X", addr)
        }
    }
}
```

### **3. Branch Distance Validation** ⭐⭐⭐⭐⭐

#### **Critical für 6502:**
```assembly
start:
    ; ... 200 bytes of code ...
    bne start   ; ERROR: Distance > 127 bytes!
```

#### **Implementation:**
```go
func (a *Analyzer) validateBranchDistance(operand Expression, token Token) {
    if label, ok := operand.(*Identifier); ok {
        if symbol, found := a.context.DefinedLabels[label.Value]; found {
            distance := symbol.Address - (a.context.CurrentPC + 2)
            if distance < -128 || distance > 127 {
                a.addError("Branch distance %d out of range", distance)
            }
        }
    }
}
```

### **4. Memory-Layout Awareness** ⭐⭐⭐⭐⭐

#### **C64-Specific Memory Analysis:**
```go
type MemoryMap struct {
    ZeroPage    Range64 // $0000-$00FF - Fast access
    Stack       Range64 // $0100-$01FF - Stack area  
    BasicArea   Range64 // $0800-$9FFF - BASIC ROM
    IO          Range64 // $D000-$DFFF - I/O registers
    Kernal      Range64 // $E000-$FFFF - KERNAL ROM
}

func (a *Analyzer) analyzeMemoryAccess(addr int64, isWrite bool) {
    if isWrite && a.isROMArea(addr) {
        a.addWarning("Writing to ROM area $%04X", addr)
    }
    if a.isIOArea(addr) {
        a.addInfo("I/O register access: $%04X", addr)
    }
}
```

### **5. Illegal Opcode Detection** ⭐⭐⭐⭐☆

```go
func (a *Analyzer) analyzeIllegalOpcodes(mnemonic string) {
    illegalOpcodes := []string{"SLO", "RLA", "SAX", "LAX", "DCP", "ISC", "JAM"}
    if contains(illegalOpcodes, mnemonic) {
        a.addWarning("'%s' is illegal opcode - may not work on all systems", mnemonic)
    }
}
```

### **6. Dead Code Detection** ⭐⭐⭐⭐☆

```assembly
start:
    jmp end
    nop         ; DEAD CODE: Never reached
    lda #$00    ; DEAD CODE: Never reached
end:
    rts
```

```go
func (a *Analyzer) detectDeadCode(statements []Statement) {
    for i, stmt := range statements {
        if a.isUnconditionalJump(stmt) {
            // Check next statements until label
            for j := i + 1; j < len(statements); j++ {
                if !a.isLabel(statements[j]) {
                    a.addWarning("Unreachable code after unconditional jump")
                }
            }
        }
    }
}
```

### **7. 6502 Jump Indirect Bug Detection** ⭐⭐⭐⭐⭐

```assembly
jmp ($20FF)  ; BUG: Reads from $20FF and $2000 instead of $20FF/$2100!
```

**Das ist ein berühmter 6502-Hardware-Bug!**

```go
func (a *Analyzer) validate6502JumpIndirect(addr int64, token Token) {
    if (addr & 0xFF) == 0xFF {
        a.addWarning("JMP ($%04X) triggers 6502 page-boundary bug - " +
                    "will read from $%04X and $%04X instead of $%04X/$%04X",
                    addr, addr, addr&0xFF00, addr, addr+1)
    }
}
```

### **8. CPU Flag Dependency Analysis** ⭐⭐⭐⭐☆

```assembly
clc           ; Clear carry
lda #$00      ; Doesn't affect carry
adc #$01      ; Uses carry from CLC - good!

cmp #$FF      ; Sets flags
lda #$00      ; OVERWRITES FLAGS!  
beq label     ; BUG: Tests A, not previous CMP!
```

```go
type CPUFlags struct {
    N, Z, C, V, I, D bool  // 6502 flags
    LastModified   Token   // Where were they last set?
}

func (a *Analyzer) analyzeCPUFlagDependency(mnemonic string, token Token) {
    flagReaders := map[string][]string{
        "BCC": {"C"}, "BCS": {"C"},
        "BEQ": {"Z"}, "BNE": {"Z"},
        "BMI": {"N"}, "BPL": {"N"},
        "BVC": {"V"}, "BVS": {"V"},
    }
    
    if flags, reads := flagReaders[mnemonic]; reads {
        for _, flag := range flags {
            if !a.context.Flags.isValid(flag) {
                a.addWarning("Branch instruction '%s' uses potentially stale %s flag", 
                           mnemonic, flag)
            }
        }
    }
}
```

### **9. Macro Analysis Enhancement** ⭐⭐⭐⭐☆

```assembly
.macro PRINT_STRING(text, color)
    lda #color
    sta $D021
    ; ... print text ...
.endmacro

; Usage analysis:
PRINT_STRING("Hello", 1)     ; OK
PRINT_STRING("Test")         ; ERROR: Missing color parameter
PRINT_STRING("A", 2, 3)      ; ERROR: Too many parameters
```

```go
type MacroDefinition struct {
    Name           string
    Parameters     []string
    LocalLabels    []string  // Labels local to macro
    Body           []Statement
    UsageCount     int
}

func (a *Analyzer) validateMacroCall(name string, args []Expression, token Token) {
    if macro, found := a.context.MacroDefinitions[name]; found {
        if len(args) != len(macro.Parameters) {
            a.addError("Macro '%s' expects %d parameters, got %d", 
                      name, len(macro.Parameters), len(args))
        }
        macro.UsageCount++
    }
}
```

### **10. Style Guide Enforcement** ⭐⭐⭐☆☆

```go
func (a *Analyzer) analyzeCodeStyle(symbol *Symbol, token Token) {
    // Konstanten sollten UPPERCASE sein
    if symbol.Kind == Constant {
        if matched, _ := regexp.MatchString(`^[a-z]`, symbol.Name); matched {
            a.addHint("Consider UPPER_CASE for constant '%s'", symbol.Name)
        }
    }
    
    // Labels sollten descriptive sein
    if symbol.Kind == Label && len(symbol.Name) < 3 {
        a.addHint("Consider more descriptive name for label '%s'", symbol.Name)
    }
    
    // Magic numbers vermeiden
    if literal, ok := expr.(*IntegerLiteral); ok {
        if a.isMagicNumber(literal.Value) {
            a.addHint("Consider defining constant for magic number %d", literal.Value)
        }
    }
}
```

## 🔄 **Implementation Strategy**

### **Phase 1: Foundation (Week 1-2)**
```go
// 1. Erweiterte Context-Struktur
type AnalysisContext struct {
    CurrentPC        int64
    DefinedLabels    map[string]*Symbol  
    ForwardRefs      []ForwardReference
    MacroDefinitions map[string]*Macro
    MemoryMap        *MemoryMap
    CPUFlags         *CPUFlags
}

// 2. Enhanced Symbol mit Address
type Symbol struct {
    Name         string
    Kind         SymbolKind
    Address      int64        // NEW: Symbol address
    Size         int64        // NEW: Symbol size in bytes
    UsageCount   int
    Position     Position
    CrossRefs    []Position   // NEW: All usage positions
}
```

### **Phase 2: Address Calculation (Week 3)**
```go
// Multi-Pass Analysis Implementation
func (a *EnhancedAnalyzer) Analyze(program *Program) []Diagnostic {
    // Pass 1: Address calculation
    a.calculateAddresses(program.Statements)
    
    // Pass 2: Forward reference resolution  
    a.resolveForwardReferences()
    
    // Pass 3-6: Enhanced analysis...
}
```

### **Phase 3: 6502 Specialization (Week 4)**
```go
// 6502-specific analysis modules
func (a *EnhancedAnalyzer) perform6502Analysis() {
    a.validateBranchDistances()
    a.optimizeZeroPageAccess() 
    a.detectHardwareBugs()
    a.analyzeMemoryLayout()
}
```

## 📊 **Expected Benefits**

### **Immediate Impact:**
- ✅ **Branch-Distance-Errors** - Verhindert Runtime-Crashes
- ✅ **Zero-Page-Optimizations** - 10-15% Code-Size-Reduction möglich
- ✅ **Hardware-Bug-Detection** - Verhindert subtile 6502-Bugs
- ✅ **Dead-Code-Elimination** - Cleaner, smaller binaries

### **Long-term Benefits:**
- ✅ **Professional Development Experience** - IDE-quality tooling
- ✅ **Learning Tool** - Teaches 6502 best practices
- ✅ **Code Quality** - Enforces assembly best practices
- ✅ **Debugging Support** - Better error messages with context

## 🎯 **Competitive Advantage**

### **Comparison mit anderen Assembly-LSPs:**
```
Feature                     | c64.nvim | ACME-LSP | CA65-LSP
----------------------------|----------|----------|----------
Address Calculation         |    ✅    |    ❌    |    ❌    
Branch Distance Validation  |    ✅    |    ❌    |    ❌
6502 Hardware Bug Detection |    ✅    |    ❌    |    ❌  
Memory Layout Awareness     |    ✅    |    ❌    |    ❌
Zero-Page Optimization      |    ✅    |    ❌    |    ❌
Illegal Opcode Warnings     |    ✅    |    ❌    |    ✅
Cross-Reference Analysis    |    ✅    |    ❌    |    ❌
```

**Fazit: Diese Enhanced Semantic Analysis würde c64.nvim zum besten Assembly-LSP für 6502/C64 machen!** 🏆

## 🚀 **Implementation Roadmap**

### **Quick Wins (1-2 days):**
1. **Branch Distance Validation** - High impact, moderate effort
2. **Illegal Opcode Detection** - Low effort, good user value  
3. **Basic Style Hints** - Easy to implement, improves UX

### **Medium-term (1-2 weeks):**
1. **Address Calculation System** - Foundation for advanced features
2. **Zero-Page Optimization** - Significant performance benefit
3. **Memory Layout Awareness** - C64-specific value

### **Advanced Features (3-4 weeks):**
1. **Dead Code Detection** - Complex but valuable
2. **CPU Flag Dependency** - Advanced analysis
3. **Cross-Reference System** - Professional IDE feature

**Das Enhanced Semantic Analysis System würde c64.nvim von einem guten LSP zu einem außergewöhnlichen, domain-expert-level Tool transformieren!** 🌟