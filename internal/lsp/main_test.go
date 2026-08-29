package lsp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// repoRoot is the directory holding the Source of Truth JSON files. Tests read
// them straight from the repository so they do not depend on an installed
// configuration in $HOME/.config/kickass_ls.
const repoRoot = "../.."

// TestMain loads the ProcessorContext once for the whole package. The lexer
// needs it to recognise mnemonics, directives and preprocessor statements.
func TestMain(m *testing.M) {
	if err := loadTestProcessorContext(); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func loadTestProcessorContext() error {
	paths := []string{"mnemonic.json", "kickass.json", "c64memory.json"}
	for _, p := range paths {
		full := filepath.Join(repoRoot, p)
		if _, err := os.Stat(full); err != nil {
			return fmt.Errorf("source of truth %s not readable: %w", full, err)
		}
	}

	return LoadProcessorContext(
		filepath.Join(repoRoot, "mnemonic.json"),
		filepath.Join(repoRoot, "kickass.json"),
		filepath.Join(repoRoot, "c64memory.json"),
	)
}

// TestProcessorContextLoaded is a guard: every other test in this package
// assumes the Source of Truth files were parsed into a populated context. If
// a JSON key or a struct tag drifts apart again, this fails first and names
// the table that came back empty, instead of letting dozens of unrelated
// assertions fail with confusing messages.
func TestProcessorContextLoaded(t *testing.T) {
	ctx := GetProcessorContext()
	if ctx == nil {
		t.Fatal("GetProcessorContext() = nil, want a loaded context")
	}

	tables := []struct {
		name string
		got  int
	}{
		{"mnemonics", len(ctx.AllMnemonics)},
		{"directives", len(ctx.Directives)},
		{"preprocessor statements", len(ctx.PreprocessorStatements)},
		{"builtin functions", len(ctx.Functions)},
		{"builtin constants", len(ctx.Constants)},
		{"memory regions", len(ctx.MemoryRegions)},
	}

	for _, tbl := range tables {
		if tbl.got == 0 {
			t.Errorf("ProcessorContext.%s is empty — check the JSON keys and struct tags", tbl.name)
		}
	}
}

// TestMnemonicClassification pins the categorisation to the type field in
// mnemonic.json. The loader used to OR the type check together with hardcoded
// opcode lists; those lists duplicated data the Source of Truth already carries
// and had gone stale (they named 18 of the 46 illegal opcodes).
func TestMnemonicClassification(t *testing.T) {
	ctx := GetProcessorContext()

	for name, m := range ctx.AllMnemonics {
		var category map[string]*EnhancedMnemonicInfo
		switch m.Type {
		case "Illegal":
			category = ctx.IllegalMnemonics
		case "Jump", "Branch", "Return":
			category = ctx.ControlMnemonics
		default:
			category = ctx.StandardMnemonics
		}
		if _, ok := category[name]; !ok {
			t.Errorf("%s (type %q) is missing from the table its type selects", name, m.Type)
		}
	}

	total := len(ctx.StandardMnemonics) + len(ctx.IllegalMnemonics) + len(ctx.ControlMnemonics)
	if total != len(ctx.AllMnemonics) {
		t.Errorf("category tables hold %d mnemonics, AllMnemonics holds %d", total, len(ctx.AllMnemonics))
	}
}
