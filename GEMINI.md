# Project Overview

This project is a Language Server for 6502/6510 assembly language, specifically tailored for Commodore C64 development. It provides features like hover information (opcode descriptions, symbol values), code completion (opcodes, labels), go-to-definition for symbols, find references, and real-time diagnostics (duplicate labels, unknown opcodes, invalid addressing modes). The core logic is implemented in Go, and it uses a `mnemonic.json` file to store detailed information about 6502/6510 opcodes.

## Building and Running

This project is a standard Go application.

### Build

To build the language server executable:

```bash
go build -o 6510lsp_server ./6510lsp_server
```

This will create an executable named `6510lsp_server` (or `6510lsp_server.exe` on Windows) in the project root directory.

### Run

The language server communicates via standard input/output, as is typical for LSP servers. It is usually run by a compatible editor or IDE that supports the Language Server Protocol.

To run it manually (for testing or debugging):

```bash
./6510lsp_server
```

You can enable debug logging by running:

```bash
./6510lsp_server --debug
```

## Development Conventions

- **Language:** Go
- **Language Server Protocol (LSP):** The server implements a subset of the LSP to provide language features for 6502/6510 Kick Assembler.
- **Data Management:** CPU mnemonic data, including opcodes, descriptions, and addressing modes, is stored in `mnemonic.json`.
- **Logging:** The project uses an internal `log` package for structured logging, supporting `INFO` and `DEBUG` levels.
- **Code Structure:** The Go codebase follows conventional Go project layout, with core LSP logic encapsulated within the `internal/lsp` package.
