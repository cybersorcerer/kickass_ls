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

// TestInstructionFactsFromSourceOfTruth guards the two lookups that replaced
// hardcoded opcode lists in analyze.go: zero page support now comes from the
// addressing_modes table, write access from the writes_memory flag.
func TestInstructionFactsFromSourceOfTruth(t *testing.T) {
	zeroPage := map[string]bool{
		"LDA": true, "STA": true, "BIT": true,
		"LAX": true, // illegal opcode, but it does have a zero page mode
		"JMP": false, "INX": false, "RTS": false, "NOPE": false,
	}
	for mnemonic, want := range zeroPage {
		if got := supportsZeroPage(mnemonic); got != want {
			t.Errorf("supportsZeroPage(%q) = %v, want %v", mnemonic, got, want)
		}
	}

	var a SemanticAnalyzer
	writes := map[string]bool{
		"STA": true, "STX": true, "STY": true,
		"INC": true, "DEC": true,
		"ASL": true, "LSR": true, "ROL": true, "ROR": true,
		"LDA": false, "CMP": false, "INX": false, "JSR": false, "NOPE": false,
	}
	for mnemonic, want := range writes {
		if got := a.isWriteInstruction(mnemonic); got != want {
			t.Errorf("isWriteInstruction(%q) = %v, want %v", mnemonic, got, want)
		}
	}
}

// TestAlwaysReachableDirectives guards the third list that moved into the
// Source of Truth. The hardcoded version named ".data", ".byt" and ".tx",
// none of which are Kick Assembler directives, while the real short forms
// ".by", ".te" and ".dw" were missing and warned falsely.
func TestAlwaysReachableDirectives(t *testing.T) {
	var a SemanticAnalyzer
	cases := map[string]bool{
		".byte": true, ".const": true, ".macro": true, ".encoding": true,
		".by": true, ".te": true, ".dw": true,
		".align": true, ".print": true, ".segment": true,
		".for": false, ".while": false, ".if": false, "else": false,
		".pseudopc": false, ".zp": false,
		".nosuchdirective": false,
	}
	for directive, want := range cases {
		if got := a.isAlwaysReachableDirective(directive); got != want {
			t.Errorf("isAlwaysReachableDirective(%q) = %v, want %v", directive, got, want)
		}
	}
}
