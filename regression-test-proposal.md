# Comprehensive Regression Test Suite Proposal

**Purpose**: Ensure no functionality breaks during kickass-plan.md implementation phases.

## Test Categories Overview

### 1. Core LSP Functionality Tests 🔌
Test the fundamental Language Server Protocol operations that users depend on daily.

### 2. Parser & Lexer Stability Tests 📝
Validate that token recognition and AST generation remain intact.

### 3. Semantic Analysis Baseline Tests 🧠
Ensure existing directive processing and symbol resolution works.

### 4. Completion System Tests 💡
Verify that all completion scenarios continue to work correctly.

### 5. Memory & Performance Tests ⚡
Monitor resource usage and response times.

## Detailed Test Specifications

## 1. Core LSP Functionality Tests 🔌

### 1.1 Server Lifecycle Test
```json
{
  "name": "LSP Server Lifecycle Test",
  "description": "Test server startup, initialization, and shutdown",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "basic.asm": "start:\n    lda #$01\n    sta $d020\n    rts"
    }
  },
  "testCases": [
    {
      "name": "Server starts successfully",
      "type": "lifecycle",
      "action": "initialize",
      "expected": {
        "responseTime": "<2000ms",
        "status": "success"
      }
    },
    {
      "name": "Server processes basic file",
      "type": "textDocument/didOpen",
      "input": {
        "file": "basic.asm"
      },
      "expected": {
        "diagnostics": "<=3",
        "status": "success"
      }
    },
    {
      "name": "Server shuts down cleanly",
      "type": "lifecycle",
      "action": "shutdown",
      "expected": {
        "exitCode": 0,
        "timeout": false
      }
    }
  ]
}
```

### 1.2 Configuration Loading Test
```json
{
  "name": "Configuration Loading Test",
  "description": "Verify all JSON config files load from $HOME/.config/6510lsp",
  "setup": {
    "serverPath": "./6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "config_test.asm": ".const test = $1000\nlda #test\ninc $d020"
    }
  },
  "testCases": [
    {
      "name": "mnemonic.json loads successfully",
      "type": "completion",
      "input": {
        "file": "config_test.asm",
        "line": 1,
        "character": 0
      },
      "expected": {
        "containsItems": [{"label": "lda"}, {"label": "inc"}],
        "minItems": 50
      }
    },
    {
      "name": "kickass.json loads successfully",
      "type": "completion",
      "input": {
        "file": "config_test.asm",
        "line": 0,
        "character": 0
      },
      "expected": {
        "containsItems": [{"label": ".const"}],
        "minItems": 10
      }
    },
    {
      "name": "c64memory.json loads successfully",
      "type": "completion",
      "input": {
        "file": "config_test.asm",
        "line": 1,
        "character": 7
      },
      "expected": {
        "containsItems": [{"label": "$D020"}],
        "minItems": 15
      }
    }
  ]
}
```

## 2. Parser & Lexer Stability Tests 📝

### 2.1 Mnemonic Recognition Test
```json
{
  "name": "Mnemonic Recognition Regression Test",
  "description": "Ensure all 6502/illegal opcodes are correctly tokenized",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "mnemonics.asm": "start:\n    lda #$01    ; Standard\n    inc $d020   ; Standard  \n    nop         ; Standard\n    jam         ; Illegal\n    slo $1000   ; Illegal\n    dcp $2000,x ; Illegal"
    }
  },
  "testCases": [
    {
      "name": "Standard mnemonics recognized",
      "type": "diagnostics",
      "input": {
        "file": "mnemonics.asm"
      },
      "expected": {
        "maxDiagnostics": 0,
        "noErrors": true
      }
    },
    {
      "name": "Illegal mnemonics recognized",
      "type": "completion",
      "input": {
        "file": "mnemonics.asm",
        "line": 5,
        "character": 0
      },
      "expected": {
        "containsItems": [{"label": "jam"}, {"label": "slo"}, {"label": "dcp"}],
        "minItems": 40
      }
    }
  ]
}
```

### 2.2 Complex Parsing Test
```json
{
  "name": "Complex Parsing Regression Test",
  "description": "Test parser handles complex structures without breaking",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "complex.asm": ".const base = $1000\n.var counter = 0\nmainloop:\n    lda base,x\n    sta $d020\n    inx\n    cpx #$10\n    bne mainloop\n!loop:\n    inc $d021\n    jmp !loop-"
    }
  },
  "testCases": [
    {
      "name": "Labels parsed correctly",
      "type": "documentSymbol",
      "input": {
        "file": "complex.asm"
      },
      "expected": {
        "containsSymbols": ["base", "counter", "mainloop"],
        "minSymbols": 3
      }
    },
    {
      "name": "Complex expressions parsed",
      "type": "diagnostics",
      "input": {
        "file": "complex.asm"
      },
      "expected": {
        "maxErrors": 0,
        "maxWarnings": 2
      }
    }
  ]
}
```

## 3. Semantic Analysis Baseline Tests 🧠

### 3.1 Directive Processing Test
```json
{
  "name": "Directive Processing Regression Test",
  "description": "Verify existing directive processing remains functional",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "directives.asm": "* = $1000\n.const color = $02\n.var temp = $80\n.byte $01, $02, $03\n.word $1234, $5678\n.text \"hello\"\nstart:\n    lda #color\n    sta temp"
    }
  },
  "testCases": [
    {
      "name": "PC directive sets program counter",
      "type": "hover",
      "input": {
        "file": "directives.asm",
        "line": 6,
        "character": 0
      },
      "expected": {
        "containsText": "$1000",
        "addressInfo": true
      }
    },
    {
      "name": "Constants defined in symbol table",
      "type": "definition",
      "input": {
        "file": "directives.asm",
        "line": 7,
        "character": 9
      },
      "expected": {
        "definitionFound": true,
        "line": 1
      }
    },
    {
      "name": "Variables accessible",
      "type": "completion",
      "input": {
        "file": "directives.asm",
        "line": 8,
        "character": 8
      },
      "expected": {
        "containsItems": [{"label": "temp"}],
        "minItems": 1
      }
    }
  ]
}
```

### 3.2 Symbol Table Test
```json
{
  "name": "Symbol Table Regression Test",
  "description": "Ensure symbol resolution and scoping works correctly",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "symbols.asm": "global_label:\n    lda #$01\n.macro test_macro(param)\n    local_label:\n        lda param\n.endmacro\nmain:\n    jsr global_label\n    test_macro($ff)"
    }
  },
  "testCases": [
    {
      "name": "Global symbols accessible",
      "type": "completion",
      "input": {
        "file": "symbols.asm",
        "line": 7,
        "character": 8
      },
      "expected": {
        "containsItems": [{"label": "global_label"}],
        "minItems": 1
      }
    },
    {
      "name": "Macro definitions processed",
      "type": "documentSymbol",
      "input": {
        "file": "symbols.asm"
      },
      "expected": {
        "containsSymbols": ["test_macro", "global_label", "main"],
        "minSymbols": 3
      }
    }
  ]
}
```

## 4. Completion System Tests 💡

### 4.1 Memory Completion Regression
```json
{
  "name": "Memory Completion Regression Test",
  "description": "Ensure memory address completion works as before (existing test enhanced)",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "memory_regression.asm": "start:\n    lda #$\n    sta $D\n    inc $DC\n    lda $DD"
    }
  },
  "testCases": [
    {
      "name": "Memory completion after #$",
      "type": "completion",
      "input": {
        "file": "memory_regression.asm",
        "line": 1,
        "character": 10
      },
      "expected": {
        "containsItems": [{"label": "$D020"}, {"label": "$D021"}],
        "minItems": 15,
        "maxItems": 30,
        "excludesFunctions": true
      }
    },
    {
      "name": "Memory completion after partial address",
      "type": "completion",
      "input": {
        "file": "memory_regression.asm",
        "line": 2,
        "character": 9
      },
      "expected": {
        "containsItems": [{"label": "$D000"}, {"label": "$D020"}],
        "minItems": 10
      }
    },
    {
      "name": "CIA register completion",
      "type": "completion",
      "input": {
        "file": "memory_regression.asm",
        "line": 3,
        "character": 9
      },
      "expected": {
        "containsItems": [{"label": "$DC00"}, {"label": "$DC01"}],
        "minItems": 5
      }
    }
  ]
}
```

### 4.2 Context-Aware Completion Test
```json
{
  "name": "Context-Aware Completion Regression Test",
  "description": "Verify completion context detection remains accurate",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "context.asm": "start:\n    l\n    lda \n    lda #\n    lda #$\n    .c\n    sin("
    }
  },
  "testCases": [
    {
      "name": "Mnemonic completion at line start",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 1,
        "character": 5
      },
      "expected": {
        "containsItems": [{"label": "lda"}, {"label": "ldy"}],
        "prioritizes": "mnemonics",
        "minItems": 45
      }
    },
    {
      "name": "Address completion after mnemonic",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 2,
        "character": 8
      },
      "expected": {
        "containsItems": [{"label": "$D020"}],
        "excludes": ["lda", "inc"],
        "context": "operand"
      }
    },
    {
      "name": "Immediate value completion",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 3,
        "character": 8
      },
      "expected": {
        "containsItems": [{"label": "$01"}, {"label": "$FF"}],
        "context": "immediate"
      }
    },
    {
      "name": "Memory address completion after #$",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 4,
        "character": 9
      },
      "expected": {
        "containsItems": [{"label": "$D020"}],
        "excludesFunctions": true,
        "context": "memory"
      }
    },
    {
      "name": "Directive completion",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 5,
        "character": 6
      },
      "expected": {
        "containsItems": [{"label": ".const"}, {"label": ".var"}],
        "context": "directive"
      }
    },
    {
      "name": "Function completion",
      "type": "completion",
      "input": {
        "file": "context.asm",
        "line": 6,
        "character": 7
      },
      "expected": {
        "containsItems": [{"label": "sin"}, {"label": "cos"}],
        "context": "function"
      }
    }
  ]
}
```

## 5. Memory & Performance Tests ⚡

### 5.1 Performance Baseline Test
```json
{
  "name": "Performance Baseline Test",
  "description": "Measure response times to detect performance regressions",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "large_file.asm": "// Auto-generated large file with 1000+ lines"
    }
  },
  "testCases": [
    {
      "name": "File parsing performance",
      "type": "performance",
      "action": "textDocument/didOpen",
      "input": {
        "file": "large_file.asm"
      },
      "expected": {
        "parseTime": "<5000ms",
        "memoryUsage": "<100MB"
      }
    },
    {
      "name": "Completion response time",
      "type": "performance",
      "action": "textDocument/completion",
      "input": {
        "file": "large_file.asm",
        "line": 500,
        "character": 10
      },
      "expected": {
        "responseTime": "<500ms",
        "resultCount": ">10"
      }
    },
    {
      "name": "Hover response time",
      "type": "performance",
      "action": "textDocument/hover",
      "input": {
        "file": "large_file.asm",
        "line": 250,
        "character": 5
      },
      "expected": {
        "responseTime": "<200ms"
      }
    }
  ]
}
```

### 5.2 Memory Leak Test
```json
{
  "name": "Memory Leak Detection Test",
  "description": "Ensure no memory leaks in long-running server sessions",
  "setup": {
    "serverPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/6510lsp_server",
    "serverArgs": [],
    "rootPath": "/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/test-cases",
    "files": {
      "memory_test.asm": "start:\n    lda #$01"
    }
  },
  "testCases": [
    {
      "name": "Memory usage stability",
      "type": "memory",
      "action": "repeatOperations",
      "operations": [
        {"type": "textDocument/didOpen", "iterations": 100},
        {"type": "textDocument/completion", "iterations": 200},
        {"type": "textDocument/didClose", "iterations": 100}
      ],
      "expected": {
        "memoryGrowth": "<10MB",
        "finalMemory": "<200MB"
      }
    }
  ]
}
```

## Regression Test Execution Strategy

### Automated Test Runner Script
```bash
#!/bin/bash
# comprehensive-regression-test.sh

set -e

echo "🧪 Comprehensive Regression Test Suite"
echo "======================================="

TEST_DIR="test-cases"
SERVER_PATH="6510lsp_server"
RESULTS_DIR="regression-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create results directory
mkdir -p "$RESULTS_DIR/$TIMESTAMP"

# Test categories
TESTS=(
    "lsp-lifecycle-test.json"
    "configuration-loading-test.json"
    "mnemonic-recognition-test.json"
    "complex-parsing-test.json"
    "directive-processing-test.json"
    "symbol-table-test.json"
    "memory-completion-regression-test.json"
    "context-aware-completion-test.json"
    "performance-baseline-test.json"
    "memory-leak-test.json"
)

# Results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo "Starting regression test execution..."
echo ""

for test in "${TESTS[@]}"; do
    echo "Running: $test"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if ./test-client/test-client -suite "$TEST_DIR/$test" -server "$SERVER_PATH" \
       -output "$RESULTS_DIR/$TIMESTAMP/${test%.json}_result.json" > /dev/null 2>&1; then
        echo "✅ PASS: $test"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "❌ FAIL: $test"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
done

echo ""
echo "📊 Regression Test Summary"
echo "========================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Success Rate: $((PASSED_TESTS * 100 / TOTAL_TESTS))%"

# Performance baseline comparison
if [ -f "baseline-performance.json" ]; then
    echo ""
    echo "📈 Performance Comparison"
    echo "========================"
    # Compare current results with baseline
    # Implementation details...
fi

# Generate detailed report
echo ""
echo "📋 Detailed results saved to: $RESULTS_DIR/$TIMESTAMP/"

if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 All regression tests passed! Safe to proceed with implementation."
    exit 0
else
    echo "⚠️  $FAILED_TESTS test(s) failed. Review failures before proceeding."
    exit 1
fi
```

## Test Maintenance Strategy

### 1. Test Suite Evolution
- **Add new tests** for each new feature
- **Update existing tests** when requirements change
- **Remove obsolete tests** that no longer reflect reality
- **Version test suites** with the codebase

### 2. Performance Baselines
- **Establish baseline** before Phase 1 implementation
- **Update baselines** only after confirmed performance improvements
- **Alert on regressions** > 20% performance degradation
- **Track trends** over time

### 3. Test Data Management
- **Realistic test cases** based on actual user code
- **Edge case coverage** for boundary conditions
- **Error scenario testing** for robust error handling
- **Large file testing** for scalability validation

## Success Criteria

### Phase Gate Requirements
Each phase must achieve:
- **100% existing functionality** tests pass
- **95% new functionality** tests pass
- **No performance regression** > 20%
- **Zero memory leaks** detected
- **All critical paths** validated

### Long-term Maintenance
- **Monthly regression runs** on main branch
- **Pre-release validation** with full test suite
- **Performance trending** reports
- **Test coverage** > 90% of critical functionality

This comprehensive regression test suite ensures that improvements to the Kick Assembler semantic analysis strengthen the system without introducing any regressions, maintaining user confidence while advancing capabilities.