# 🎮 C64.nvim - 6502/6510 Assembly Language Server

<p align="center">
  <img src="https://img.shields.io/badge/language-6502%2F6510%20Assembly-blue" alt="Language">
  <img src="https://img.shields.io/badge/assembler-Kick%20Assembler-orange" alt="Kick Assembler">
  <img src="https://img.shields.io/badge/platform-Commodore%20C64-red" alt="Platform">
  <img src="https://img.shields.io/badge/LSP-Language%20Server%20Protocol-green" alt="LSP">
  <img src="https://img.shields.io/badge/editor-Neovim%200.11+-purple" alt="Neovim">
</p>

A comprehensive **Language Server Protocol (LSP)** implementation for **6502/6510 assembly language** development, specifically designed for **Commodore 64** programming with **Kick Assembler** support. This project provides intelligent code analysis, semantic validation, and development assistance for retro computing enthusiasts and developers working with modern C64 development tools.

## ✨ Features

### 🧠 **Advanced Semantic Analysis**
- **Multi-pass semantic analysis** with 6-pass architecture
- **Program Counter (PC) tracking** through all statements
- **Symbol resolution** with forward reference support
- **Address calculation** and label resolution
- **Cross-reference analysis** with scope awareness

### 🎯 **6502/C64 Specialized Features**
- **Branch distance validation** (-128 to +127 byte range)
- **Zero-page optimization** detection and hints
- **Illegal opcode detection** with warnings
- **Hardware bug detection** (JMP indirect bug, etc.)
- **Memory layout analysis** (ROM/RAM/I/O regions)
- **Magic number detection** for C64-specific addresses

### 🔧 **Code Quality & Optimization**
- **Dead code detection** and unreachable code analysis
- **Unused symbol warnings** with smart filtering
- **Style guide enforcement** with configurable rules
- **Memory access pattern analysis**
- **CPU flag dependency tracking**

### ⚡ **Kick Assembler Integration**
- **Complete directive support** for modern Kick Assembler syntax
- **Variable and constant tracking** (`.var`, `.const`)
- **Macro and pseudocommand** parsing and validation
- **Namespace and scope** awareness
- **Function definition** support with parameter validation
- **Import directive** handling (`#import`, `.import`)
- **Advanced control flow** (`.if`, `.for`, `.while`)
- **Built-in function** completion and documentation

### ⚙️ **Flexible Configuration**
- **Multiple configuration profiles**: `default`, `legacy`, `minimal`
- **Project-specific configuration** via `.6510lsp.json`
- **Runtime configuration updates** without restart
- **Configurable warnings and hints** for all features

## 🚀 Quick Start

### Prerequisites
- **Go 1.25.1+** for building the language server
- **Neovim 0.11+** with native LSP support
- Basic knowledge of 6502/6510 assembly language

### Installation

1. **Clone and build the language server:**
```bash
git clone <repository-url> c64.nvim
cd c64.nvim
go build -o 6510lsp_server .
```

2. **Install Neovim configuration:**
```bash
# Copy the LSP configuration to your Neovim config
cp lua/6510_ls.lua ~/.config/nvim/lsp/6510_ls.lua

# Add to your init.lua or use the provided lsp.lua
```

3. **Configure your Neovim setup** (see [Neovim Integration](#neovim-integration))

## 🛠️ Configuration

### Project Configuration
Create a `.6510lsp.json` file in your project root:

```json
{
  "6510lsp": {
    "warnUnusedLabels": true,
    "zeroPageOptimization": {
      "enabled": true,
      "showHints": true
    },
    "branchDistanceValidation": {
      "enabled": true,
      "showWarnings": true
    },
    "illegalOpcodeDetection": {
      "enabled": true,
      "showWarnings": true
    },
    "hardwareBugDetection": {
      "enabled": true,
      "showWarnings": true,
      "jmpIndirectBug": true
    },
    "memoryLayoutAnalysis": {
      "enabled": true,
      "showIOAccess": true,
      "showStackWarnings": true,
      "showROMWriteWarnings": true
    },
    "magicNumberDetection": {
      "enabled": true,
      "showHints": true,
      "c64Addresses": true
    },
    "deadCodeDetection": {
      "enabled": true,
      "showWarnings": true
    },
    "styleGuideEnforcement": {
      "enabled": true,
      "showHints": true,
      "upperCaseConstants": true,
      "descriptiveLabels": true
    }
  }
}
```

### Configuration Profiles

The language server comes with three built-in profiles:

#### 🎯 **Default Profile** - Full feature set
All analysis features enabled with comprehensive warnings and hints.

#### 🏛️ **Legacy Profile** - Balanced for older codebases
Reduced strictness for working with existing legacy code.

#### 🔧 **Minimal Profile** - Critical errors only
Only essential error detection for minimal overhead.

## 🖥️ Neovim Integration

### Modern Neovim Setup (0.11+)
Place this in `~/.config/nvim/lsp/6510_ls.lua`:

```lua
-- 6510 LSP Configuration with safe dependency loading
local log = {}
local capabilities = vim.lsp.protocol.make_client_capabilities()

-- Try to load optional dependencies
local ok_log, plenary_log = pcall(require, "plenary.log")
if ok_log then
    log = plenary_log.new({
        plugin = "lsp.6510_ls",
        level = "info",
    })
else
    -- Fallback logging
    log = {
        debug = function(msg) end,
        info = function(msg) print("[6510LSP] " .. msg) end,
        warn = function(msg) vim.notify("[6510LSP] " .. msg, vim.log.levels.WARN) end,
        error = function(msg) vim.notify("[6510LSP] " .. msg, vim.log.levels.ERROR) end,
    }
end

-- Configuration profiles and functions...
-- (See the full configuration file in the repository)
```

Add to your `~/.config/nvim/lua/ronny/core/lsp.lua` (or equivalent):

```lua
-- Auto-start LSP for 6510 assembly files
vim.api.nvim_create_autocmd({"BufEnter", "BufWinEnter"}, {
    pattern = {"*.asm", "*.s", "*.6510", "*.inc", "*.a"},
    group = vim.api.nvim_create_augroup("6510_lsp_start", { clear = true }),
    callback = function()
        local clients = vim.lsp.get_clients({ name = "6510lsp" })
        if #clients == 0 then
            local config_file = vim.fn.stdpath("config") .. "/lsp/6510_ls.lua"
            if vim.fn.filereadable(config_file) == 1 then
                local ok, config = pcall(function()
                    return dofile(config_file)
                end)
                if ok and config then
                    vim.lsp.start(vim.tbl_extend("force", config.lsp_config, {
                        name = "6510lsp",
                        root_dir = vim.fs.root(0, config.lsp_config.root_markers) or vim.fn.getcwd(),
                    }))
                end
            end
        end
    end,
})
```

### Key Bindings
The LSP provides these 6510-specific key bindings:

- `<leader>ct` - Toggle C64 LSP hints on/off
- `<leader>cs` - Show current LSP configuration profile

### Commands
- `:C64LspProfile <profile>` - Switch between configuration profiles
- `:C64LspToggleHints` - Toggle hints for optimization features
- `:C64LspRestart` - Restart the language server

## 📂 Project Structure

```
c64.nvim/
├── 6510lsp_server          # Compiled language server binary
├── main.go                 # Main entry point
├── go.mod                  # Go module definition
├── internal/               # Internal Go packages
│   ├── lsp/               # LSP implementation
│   │   ├── server.go      # LSP server core
│   │   ├── analyze.go     # Semantic analysis engine
│   │   └── handlers.go    # LSP message handlers
│   └── log/               # Logging utilities
├── instructions/           # 6502/6510 instruction definitions
│   ├── adc.json          # Individual instruction metadata
│   ├── sta.json          # ...
│   └── ...               # Complete instruction set
├── lua/                   # Neovim Lua configuration
│   ├── 6510_ls.lua       # LSP client configuration
│   └── lsp.lua           # Integration helpers
├── test/                  # Test assembly files
├── .6510lsp.json         # Project-specific configuration
├── lsp-config-example.json # Configuration template
└── README.md             # This file
```

## 🧪 Semantic Analysis Architecture

The language server implements a sophisticated **6-pass semantic analysis** system:

### Pass 1: Symbol Collection & Address Calculation
- PC tracking through all statements
- Label address calculation
- Directive processing (`.pc`, `.byte`, `.word`, etc.)
- Forward reference collection

### Pass 2: Forward Reference Resolution
- Symbol reference resolution
- Address dependency calculation
- Circular dependency detection

### Pass 3: Enhanced Usage Analysis
- Symbol usage tracking
- Scope-aware cross-references
- Comment filtering

### Pass 4: 6502/C64 Specialized Analysis
- Branch distance validation
- Zero-page access optimization
- Illegal opcode warnings
- Memory access pattern analysis
- CPU flag dependency tracking

### Pass 5: Optimization Hints
- Dead code detection
- Unreachable code detection
- Redundant instruction analysis

### Pass 6: Final Validation & Reporting
- Error consolidation
- Warning prioritization
- Diagnostic generation

## 🎨 Supported File Types

The language server recognizes these file extensions:
- `.asm` - Assembly source files (Kick Assembler standard)
- `.s` - Assembly source files (Unix convention)
- `.6510` - 6510-specific assembly files
- `.inc` - Include files (Kick Assembler includes)
- `.a` - Assembly files (alternative extension)

## ⚡ Kick Assembler Support

This language server is specifically designed for **Kick Assembler**, the most popular modern assembler for C64 development. It provides comprehensive support for Kick Assembler's advanced features:

### 📋 **Currently Supported Directives**

| Directive | Support Level | Description |
|-----------|---------------|-------------|
| `.const` | ✅ **Full** | Constant definitions with scope tracking |
| `.var` | ✅ **Full** | Variable declarations and assignments |
| `.eval` | ✅ **Full** | Expression evaluation |
| `.enum` | ✅ **Full** | Enumeration definitions |
| `.label` | ✅ **Full** | Label declarations and multi-labels |
| `.byte/.word/.dword` | ✅ **Full** | Data generation directives |
| `.text` | ✅ **Full** | String data generation |
| `.fill` | ✅ **Full** | Memory fill operations |
| `#import` | ✅ **Full** | Preprocessor imports |
| `.import` | ✅ **Full** | Binary/C64 file imports |
| `.macro` | ✅ **Full** | Macro definitions and calls |
| `.pseudocommand` | ✅ **Full** | Custom command definitions |
| `.function` | ✅ **Full** | User-defined functions |
| `.namespace` | ✅ **Full** | Namespace containers |
| `.if/.else` | ✅ **Full** | Conditional compilation |
| `.for/.while` | 🔄 **Planned** | Loop constructs |
| `.align` | ✅ **Full** | Memory alignment |
| `.assert/.asserterror` | 🔄 **Planned** | Testing directives |
| `.break` | ✅ **Full** | Debug breakpoints |
| `.cpu` | ✅ **Full** | CPU instruction set selection |
| `.print/.printnow` | ✅ **Full** | Assembly-time output |

### 🚀 **Advanced Kick Assembler Features**

#### **Variable and Constant System**
```kickassembler
.const SCREEN = $0400        // Tracked as constant
.var score = 0               // Tracked as variable
.eval score = score + 100    // Expression evaluation
```

#### **Macro and Pseudocommand Support**
```kickassembler
.macro SetColor(color) {     // Full macro parsing
    lda #color
    sta $d020
}

.pseudocommand mov src:tar { // Custom command syntax
    lda src
    sta tar
}
```

#### **Namespace Awareness**
```kickassembler
.namespace Graphics {        // Scope isolation
    .function drawPixel(x, y) { ... }
    .var currentColor = 1
}
```

#### **Import System**
```kickassembler
#import "MyLibrary.asm"           // Preprocessor imports
.import binary "Music.bin"        // Binary data imports
.import c64 "charset.c64", $400   // C64 file format
```

### 🎯 **Planned Kick Assembler Enhancements**

#### **Phase 1: Advanced Control Flow** (Q1 2024)
- ✅ `.for` loop parsing and validation
- ✅ `.while` loop support
- ✅ Nested control structure validation
- ✅ Loop variable tracking

#### **Phase 2: Enhanced Testing Support** (Q2 2024)
- ✅ `.assert` directive validation
- ✅ `.asserterror` testing support
- ✅ Test result reporting
- ✅ Conditional test execution

#### **Phase 3: Advanced Expression Engine** (Q3 2024)
- ✅ Built-in function library completion
- ✅ Mathematical expression validation
- ✅ String manipulation functions
- ✅ List and vector operations

#### **Phase 4: Project Management** (Q4 2024)
- ✅ Multi-file project analysis
- ✅ Dependency graph generation
- ✅ Cross-file symbol resolution
- ✅ Build configuration support

### 💡 **Kick Assembler Best Practices**

The language server enforces and suggests Kick Assembler best practices:
- **Consistent naming conventions** for variables and constants
- **Proper scope usage** with namespaces
- **Efficient macro design** with parameter validation
- **Memory-aware coding** with alignment suggestions
- **Modular development** with proper import structure

## 🐛 Troubleshooting

### Common Issues

**LSP not starting:**
- Ensure the `6510lsp_server` binary is executable
- Check the path in your LSP configuration
- Verify file extensions are recognized

**Configuration not loading:**
- Check `.6510lsp.json` syntax with a JSON validator
- Ensure the file is in your project root or a parent directory
- Review `:messages` in Neovim for error details

**Missing features:**
- Verify your configuration profile enables the desired features
- Check if project-specific configuration overrides global settings

### Debugging

Enable debug logging in your LSP configuration:
```lua
log = plenary_log.new({
    plugin = "lsp.6510_ls",
    level = "debug",  -- Change from "info" to "debug"
})
```

Check LSP logs:
```bash
tail -f 6510lsp.log
```

## 🤝 Contributing

Contributions are welcome! Areas of particular interest:

- **Additional 6502 analysis features**
- **Enhanced C64-specific validations**
- **Support for other 6502-based platforms**
- **Performance optimizations**
- **Documentation improvements**

### Development Setup

1. Clone the repository
2. Install Go 1.25.1+
3. Run tests: `go test ./...`
4. Build: `go build -o 6510lsp_server .`

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- **Kick Assembler** - Mads Nielsen's excellent modern C64 assembler
- **6502.org** - Comprehensive 6502 documentation and community
- **C64 Wiki** - Detailed C64 hardware and software information
- **Language Server Protocol** - Microsoft's LSP specification
- **Neovim LSP** - Built-in LSP client implementation
- **C64 Scene** - The vibrant retro computing community

---

<p align="center">
Made with ❤️ for the retro computing community
</p>