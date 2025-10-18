# Kick Assembler Language Server

A Language Server Protocol (LSP) implementation for Kick Assembler, providing modern IDE features for 6502/6510 assembly development on Commodore 64.

**Note:** This is the first release of the Kick Assembler Language Server. While the server has been extensively tested on macOS, installation and functionality on Windows and Linux have not been thoroughly tested yet. Your feedback and bug reports are greatly appreciated!

Made with love for the retro computing community.

## Table of Contents

- [Kick Assembler Language Server](#kick-assembler-language-server)
	- [Table of Contents](#table-of-contents)
	- [Features](#features)
	- [Installation](#installation)
		- [Quick Install (Recommended)](#quick-install-recommended)
		- [Manual Installation](#manual-installation)
		- [Building from Source](#building-from-source)
	- [Editor Configuration](#editor-configuration)
		- [Neovim](#neovim)
		- [Visual Studio Code](#visual-studio-code)
		- [Other LSP-Compatible Editors](#other-lsp-compatible-editors)
	- [Language Features](#language-features)
		- [Diagnostics](#diagnostics)
		- [Code Completion](#code-completion)
		- [Hover Information](#hover-information)
		- [Go to Definition](#go-to-definition)
		- [Document Symbols](#document-symbols)
		- [Semantic Highlighting](#semantic-highlighting)
	- [Project Structure](#project-structure)
	- [Development](#development)
		- [Prerequisites](#prerequisites)
		- [Building](#building)
		- [Testing](#testing)
	- [Server Configuration](#server-configuration)
		- [Configuration Profiles](#configuration-profiles)
		- [Available Settings](#available-settings)
			- [General Analysis](#general-analysis)
			- [6502-Specific Features](#6502-specific-features)
				- [Zero Page Optimization](#zero-page-optimization)
				- [Branch Distance Validation](#branch-distance-validation)
				- [Illegal Opcode Detection](#illegal-opcode-detection)
			- [Hardware Bug Detection](#hardware-bug-detection)
			- [Memory Layout Analysis](#memory-layout-analysis)
			- [Code Quality Features](#code-quality-features)
				- [Magic Number Detection](#magic-number-detection)
				- [Dead Code Detection](#dead-code-detection)
				- [Style Guide Enforcement](#style-guide-enforcement)
		- [Configuration Examples](#configuration-examples)
			- [Neovim (nvim-lspconfig)](#neovim-nvim-lspconfig)
			- [Minimal Profile (Only Critical Errors)](#minimal-profile-only-critical-errors)
			- [Legacy Code Profile (Less Strict)](#legacy-code-profile-less-strict)
		- [Project-Specific Configuration](#project-specific-configuration)
		- [Command-Line Flags](#command-line-flags)
	- [Configuration Files](#configuration-files)
		- [kickass.json](#kickassjson)
		- [mnemonic.json](#mnemonicjson)
		- [c64memory.json](#c64memoryjson)
	- [Contributing](#contributing)
	- [License](#license)

## Features

The Kick Assembler Language Server provides comprehensive language support for 6502/6510 assembly programming with Kick Assembler syntax:

- **Context-aware parsing** - Deep understanding of Kick Assembler directives, macros, functions, and namespaces
- **Real-time diagnostics** - Instant feedback on syntax errors, invalid mnemonics, and addressing mode violations
- **Intelligent completion** - Context-sensitive suggestions for mnemonics, directives, labels, constants, and C64 memory-mapped registers
- **Hover documentation** - Detailed information for mnemonics, directives, functions, and hardware registers
- **Symbol navigation** - Jump to definition for labels, constants, variables, functions, macros, and pseudocommands
- **Semantic highlighting** - Syntax-aware token classification for better code visualization
- **C64 memory map integration** - Built-in knowledge of VIC-II, SID, CIA registers with hardware-specific hints
- **Multi-pass analysis** - Accurate program counter tracking and forward reference resolution

## Installation

### Quick Install (Recommended)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/cybersorcerer/kickass_ls/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/cybersorcerer/kickass_ls/main/install.ps1 | iex
```

The installer will:
- Download the latest release for your platform
- Install the server binary to `~/.local/bin/` (Unix) or `%LOCALAPPDATA%\kickass_ls\bin\` (Windows)
- Install configuration files to `~/.config/kickass_ls/` (Unix) or `%LOCALAPPDATA%\kickass_ls\config\` (Windows)
- Update your PATH if needed

### Manual Installation

1. Download the appropriate archive for your platform from the [releases page](https://github.com/cybersorcerer/kickass_ls/releases):
   - macOS Intel: `kickass_ls-vX.X.X-darwin-amd64.tar.gz`
   - macOS Apple Silicon: `kickass_ls-vX.X.X-darwin-arm64.tar.gz`
   - Linux x86_64: `kickass_ls-vX.X.X-linux-amd64.tar.gz`
   - Linux ARM64: `kickass_ls-vX.X.X-linux-arm64.tar.gz`
   - Windows x86_64: `kickass_ls-vX.X.X-windows-amd64.zip`

2. Extract the archive:
   ```bash
   # Unix/macOS
   tar -xzf kickass_ls-vX.X.X-*.tar.gz
   cd kickass_ls-vX.X.X-*/

   # Windows (PowerShell)
   Expand-Archive kickass_ls-vX.X.X-windows-amd64.zip
   cd kickass_ls-vX.X.X-windows-amd64
   ```

3. Run the included installation script:
   ```bash
   # Unix/macOS
   ./install.sh

   # Windows (PowerShell)
   .\install.ps1
   ```

### Building from Source

**Prerequisites:**
- Go 1.21 or later
- Make (optional, for convenience)

**Build steps:**
```bash
# Clone the repository
git clone https://github.com/cybersorcerer/kickass_ls.git
cd kickass_ls

# Build for your current platform
make build

# Or build for all platforms
make build-all

# Install locally
make install
```

The server binary will be installed to `~/.local/bin/kickass_ls` and configuration files to `~/.config/kickass_ls/`.

## Editor Configuration

### Neovim

Using `nvim-lspconfig`:

```lua
local lspconfig = require('lspconfig')

-- Add kickass_ls to the LSP config
local configs = require('lspconfig.configs')
if not configs.kickass_ls then
  configs.kickass_ls = {
    default_config = {
      cmd = { 'kickass_ls' },
      filetypes = { 'kickasm', 'asm' },
      root_dir = lspconfig.util.root_pattern('.git', '.kickass'),
      settings = {},
    },
  }
end

-- Setup the LSP
lspconfig.kickass_ls.setup({
  on_attach = function(client, bufnr)
    -- Your keybindings and configuration here
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, { buffer = bufnr })
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, { buffer = bufnr })
    vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, { buffer = bufnr })
  end,
  capabilities = require('cmp_nvim_lsp').default_capabilities(),
})
```

**File type detection:**

Add to your `~/.config/nvim/filetype.lua`:
```lua
vim.filetype.add({
  extension = {
    asm = 'kickasm',
    s = 'kickasm',
  },
  pattern = {
    ['.*%.asm'] = 'kickasm',
  },
})
```

### Visual Studio Code

Create or edit `.vscode/settings.json` in your project:

```json
{
  "kickasm.lsp.enabled": true,
  "kickasm.lsp.serverPath": "/path/to/kickass_ls"
}
```

Note: A dedicated VSCode extension is planned but not yet available.

### Other LSP-Compatible Editors

The Kick Assembler Language Server implements the Language Server Protocol and should work with any LSP-compatible editor. Configure your editor to:

1. Launch `kickass_ls` as the language server
2. Associate `.asm` files with the Kick Assembler language
3. Set the workspace root to your project directory

## Language Features

### Diagnostics

Real-time error detection and validation:

- **Invalid mnemonics** - Unknown or misspelled 6502/6510 instructions
- **Addressing mode violations** - Instructions used with unsupported addressing modes
- **Branch distance errors** - Relative branches exceeding +127/-128 byte range
- **Invalid encodings** - Unrecognized encoding names in `.encoding` directive
- **Syntax errors** - Malformed expressions, directives, or statements

### Code Completion

Context-aware completions for:

- **Mnemonics** - All standard and illegal 6502/6510 opcodes with addressing mode hints
- **Directives** - Kick Assembler directives (`.byte`, `.word`, `.const`, `.var`, `.macro`, etc.)
- **Labels** - Local and namespace-qualified labels
- **Constants and variables** - Defined with `.const` and `.var`
- **Functions and macros** - User-defined and built-in functions
- **C64 memory map** - VIC-II registers ($D000-$D02E), SID registers ($D400-$D418), CIA, Color RAM ($D800)
- **Built-in constants** - Predefined colors, screen codes, and system constants

### Hover Information

Detailed documentation on hover:

- **Mnemonics** - Instruction description, addressing modes, cycle counts
- **Directives** - Syntax and usage examples
- **Hardware registers** - Register function, bit fields, hardware-specific warnings (e.g., "CLEARED ON READ" for collision registers)
- **Functions** - Parameter types, return values, and descriptions
- **Labels and symbols** - Value, type, and scope information

### Go to Definition

Jump to definition for:

- Labels
- Constants (`.const`)
- Variables (`.var`)
- Functions (`.function`)
- Macros (`.macro`)
- Pseudocommands (`.pseudocommand`)
- Namespace members

### Document Symbols

Hierarchical symbol outline showing:

- Namespaces
- Functions and macros
- Constants and variables
- Labels
- Organized by scope and nesting

### Semantic Highlighting

Syntax-aware token classification for:

- Mnemonics (standard, illegal, control flow)
- Directives
- Labels
- Constants and variables
- Functions and macros
- Registers and addressing modes
- Numbers (hex, binary, decimal)
- Strings and comments

## Project Structure

```
.
├── internal/lsp/           # LSP server implementation
│   ├── server.go          # LSP protocol handlers
│   ├── context_aware_lexer.go   # Tokenizer
│   ├── context_aware_parser.go  # Parser and AST
│   ├── analyze.go         # Semantic analysis
│   ├── semantic.go        # Semantic tokens
│   ├── completion.go      # Code completion
│   └── hover.go           # Hover information
├── kickass_cl/            # Test client
│   ├── main.go           # CLI interface
│   ├── client.go         # LSP client
│   ├── protocol.go       # Protocol types
│   └── runner.go         # Test runner
├── test-cases/           # Test suites
│   ├── regression-test/  # Regression tests
│   ├── 0.9.0-baseline/   # Baseline tests
│   └── test-files/       # Integration tests
├── kickass.json          # Kick Assembler directives
├── mnemonic.json         # 6502/6510 mnemonics
├── c64memory.json        # C64 memory map
├── Makefile              # Build automation
└── main.go               # Server entry point
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional)
- Git

### Building

```bash
# Build server and test client
make build

# Build for specific platform
make darwin/arm64
make linux/amd64
make windows/amd64

# Build for all platforms
make build-all

# Create release packages
make release
```

### Testing

The project uses integration tests with the test client:

```bash
# Run integration tests
make test-integration

# Run all regression tests
./run-regression-tests.sh

# Run specific test suite
build/kickass_cl -suite test-cases/regression-test/regression-suite.json -server build/kickass_ls
```

**Test client options:**
- `-suite <file>` - Run JSON test suite
- `-server <path>` - Path to server binary
- `-debug` - Enable debug logging

## Server Configuration

The Kick Assembler Language Server is highly configurable. You can enable or disable various code analysis features and adjust their behavior to match your workflow.

### Configuration Profiles

The server supports runtime configuration through LSP settings. You can configure it in your editor or create project-specific settings.

### Available Settings

All settings are organized under the `kickass_ls` namespace:

#### General Analysis

- **warnUnusedLabels** (boolean, default: `true`)
  - Show warnings for labels that are defined but never used
  - Helps identify dead code and typos

#### 6502-Specific Features

##### Zero Page Optimization

- **zeroPageOptimization.enabled** (boolean, default: `true`)
- **zeroPageOptimization.showHints** (boolean, default: `true`)
  - Suggests using zero-page addressing when accessing $00-$FF addresses
  - Example: `LDA $0080` → Hint: "Consider zero-page addressing: LDA $80"
  - Saves 1 byte and 1 cycle per instruction

##### Branch Distance Validation

- **branchDistanceValidation.enabled** (boolean, default: `true`)
- **branchDistanceValidation.showWarnings** (boolean, default: `true`)
  - Validates that relative branches stay within -128 to +127 byte range
  - Shows exact distance and suggests using JMP for longer distances

##### Illegal Opcode Detection

- **illegalOpcodeDetection.enabled** (boolean, default: `true`)
- **illegalOpcodeDetection.showWarnings** (boolean, default: `true`)
  - Detects use of illegal/undocumented 6502 opcodes (SLO, RLA, etc.)
  - Warns about stability and compatibility issues

#### Hardware Bug Detection

- **hardwareBugDetection.enabled** (boolean, default: `true`)
- **hardwareBugDetection.showWarnings** (boolean, default: `true`)
- **hardwareBugDetection.jmpIndirectBug** (boolean, default: `true`)
  - Detects the famous JMP ($xxFF) page boundary bug
  - Example: `JMP ($10FF)` → Warning: "JMP indirect wraps to $1000 instead of $1100"
  - Critical for avoiding hard-to-debug issues

#### Memory Layout Analysis

- **memoryLayoutAnalysis.enabled** (boolean, default: `true`)
- **memoryLayoutAnalysis.showIOAccess** (boolean, default: `true`)
- **memoryLayoutAnalysis.showStackWarnings** (boolean, default: `true`)
- **memoryLayoutAnalysis.showROMWriteWarnings** (boolean, default: `true`)
  - Analyzes memory access patterns
  - Warns about writes to ROM areas ($A000-$BFFF, $E000-$FFFF)
  - Detects stack issues ($0100-$01FF)
  - Highlights I/O register access ($D000-$DFFF)

#### Code Quality Features

##### Magic Number Detection

- **magicNumberDetection.enabled** (boolean, default: `true`)
- **magicNumberDetection.showHints** (boolean, default: `true`)
- **magicNumberDetection.c64Addresses** (boolean, default: `true`)
  - Detects hardcoded numbers that should be named constants
  - Recognizes common C64 addresses ($D020, $D021, etc.)
  - Example: `STA $D020` → Hint: "Consider using constant: BORDER_COLOR"

##### Dead Code Detection

- **deadCodeDetection.enabled** (boolean, default: `true`)
- **deadCodeDetection.showWarnings** (boolean, default: `true`)
  - Finds unreachable code after unconditional jumps
  - Detects code after RTS with no label

##### Style Guide Enforcement

- **styleGuideEnforcement.enabled** (boolean, default: `true`)
- **styleGuideEnforcement.showHints** (boolean, default: `true`)
- **styleGuideEnforcement.upperCaseConstants** (boolean, default: `true`)
- **styleGuideEnforcement.descriptiveLabels** (boolean, default: `true`)
  - Suggests UPPER_CASE naming for constants
  - Warns about very short label names (< 3 characters)

### Configuration Examples

#### Neovim (nvim-lspconfig)

```lua
local lspconfig = require('lspconfig')

lspconfig.kickass_ls.setup({
  settings = {
    kickass_ls = {
      -- Enable all features (default profile)
      warnUnusedLabels = true,
      zeroPageOptimization = {
        enabled = true,
        showHints = true,
      },
      branchDistanceValidation = {
        enabled = true,
        showWarnings = true,
      },
      illegalOpcodeDetection = {
        enabled = true,
        showWarnings = true,
      },
      hardwareBugDetection = {
        enabled = true,
        showWarnings = true,
        jmpIndirectBug = true,
      },
      memoryLayoutAnalysis = {
        enabled = true,
        showIOAccess = true,
        showStackWarnings = true,
        showROMWriteWarnings = true,
      },
      magicNumberDetection = {
        enabled = true,
        showHints = true,
        c64Addresses = true,
      },
      deadCodeDetection = {
        enabled = true,
        showWarnings = true,
      },
      styleGuideEnforcement = {
        enabled = true,
        showHints = true,
        upperCaseConstants = true,
        descriptiveLabels = true,
      },
    },
  },
})
```

#### Minimal Profile (Only Critical Errors)

```lua
settings = {
  kickass_ls = {
    warnUnusedLabels = false,
    zeroPageOptimization = { enabled = false },
    branchDistanceValidation = { enabled = true, showWarnings = true },
    illegalOpcodeDetection = { enabled = false },
    hardwareBugDetection = { enabled = true, showWarnings = true },
    memoryLayoutAnalysis = { enabled = false },
    magicNumberDetection = { enabled = false },
    deadCodeDetection = { enabled = false },
    styleGuideEnforcement = { enabled = false },
  },
}
```

#### Legacy Code Profile (Less Strict)

```lua
settings = {
  kickass_ls = {
    warnUnusedLabels = false,
    zeroPageOptimization = { enabled = true, showHints = false },
    branchDistanceValidation = { enabled = true, showWarnings = true },
    illegalOpcodeDetection = { enabled = false },
    hardwareBugDetection = { enabled = true, showWarnings = true },
    memoryLayoutAnalysis = { enabled = false },
    magicNumberDetection = { enabled = false },
    deadCodeDetection = { enabled = false },
    styleGuideEnforcement = { enabled = false },
  },
}
```

### Project-Specific Configuration

Create a `.kickass_ls.json` file in your project root:

```json
{
  "kickass_ls": {
    "warnUnusedLabels": true,
    "zeroPageOptimization": {
      "enabled": true,
      "showHints": true
    },
    "styleGuideEnforcement": {
      "enabled": false
    }
  }
}
```

The server will automatically use project-specific settings when found.

### Command-Line Flags

When starting the server directly (outside LSP mode):

- `--debug` - Enable debug logging to `~/.local/share/kickass_ls/log/kickass_ls.log`
- `--warn-unused-labels` - Enable unused label warnings (can also be set via LSP settings)

Example:

```bash
kickass_ls --debug
```

## Configuration Files

The language server uses three JSON configuration files located in `~/.config/kickass_ls/`:

### kickass.json

Defines Kick Assembler directives with syntax, parameters, and descriptions. Used for:
- Directive validation
- Code completion
- Hover documentation

### mnemonic.json

Defines all 6502/6510 mnemonics including:
- Standard opcodes
- Illegal/undocumented opcodes
- Addressing modes
- Cycle counts
- Flags affected

### c64memory.json

C64 memory map with hardware registers:
- VIC-II registers ($D000-$D02E)
- SID registers ($D400-$D418)
- CIA registers ($DC00-$DCFF, $DD00-$DDFF)
- Color RAM ($D800-$DBE7)
- Kernal ROM addresses
- Hardware-specific tips and warnings

These files are the single source of truth for the language server. Custom configurations can be added by editing these files.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Development guidelines:**
- Follow Go best practices and conventions
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting PR

## License

Copyright 2025 Ronny Funk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

```text
http://www.apache.org/licenses/LICENSE-2.0
```

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

---

**Note:** This is a Language Server implementation. It provides the backend language intelligence. You will need to configure your editor/IDE to use this server for Kick Assembler files.

For issues, feature requests, or questions, please visit the [GitHub repository](https://github.com/cybersorcerer/kickass_ls).
