package lsp

import (
	"strings"
	"testing"
)

// parseSource runs the full lexer/parser/analyzer pipeline the way the server
// does and returns the scope plus the diagnostics.
func parseSource(t *testing.T, src string) (*Scope, []Diagnostic) {
	t.Helper()
	scope, _, diags := ParseDocument("file:///test.asm", src)
	if scope == nil {
		t.Fatal("ParseDocument returned a nil scope")
	}
	return scope, diags
}

// errorsOnly keeps the diagnostics that mark the source as broken. Warnings and
// hints are style feedback and are not interesting for parser tests.
func errorsOnly(diags []Diagnostic) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Severity == SeverityError {
			out = append(out, d)
		}
	}
	return out
}

func assertNoErrors(t *testing.T, src string) {
	t.Helper()
	_, diags := parseSource(t, src)
	errs := errorsOnly(diags)
	if len(errs) == 0 {
		return
	}
	t.Errorf("valid source produced %d error(s):", len(errs))
	for _, e := range errs {
		t.Errorf("  line %d col %d: %s", e.Range.Start.Line+1, e.Range.Start.Character+1, e.Message)
	}
	t.Logf("source:\n%s", src)
}

func assertHasSymbols(t *testing.T, scope *Scope, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, ok := scope.FindSymbol(n); !ok {
			t.Errorf("symbol %q not found in the scope tree", n)
		}
	}
}

// --- preprocessor blocks -------------------------------------------------

// TestParserPreprocessorBlocks guards a defect where the statement terminator
// check only looked for a "." prefix on the literal. Preprocessor statements
// start with "#", so an instruction consumed the following #endif as its
// operand, which produced "Unexpected token '#endif' in expression" and left
// the #if block counted as unclosed.
func TestParserPreprocessorBlocks(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "if block closed after instructions",
			src:  "#define DEBUG\n#if DEBUG\n    nop\n    nop\n#endif\n",
		},
		{
			name: "if block closed after a directive",
			src:  "#define DEBUG\n#if DEBUG\n    .byte 1, 2, 3\n#endif\n",
		},
		{
			name: "if else endif around instructions",
			src:  "#define DEBUG\n#if DEBUG\n    nop\n#else\n    rts\n#endif\n",
		},
		{
			name: "two consecutive blocks",
			src:  "#define A\n#define B\n#if A\n    nop\n#endif\n#if B\n    rts\n#endif\n",
		},
		{
			name: "importonce followed by an instruction",
			src:  "#importonce\n    rts\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertNoErrors(t, tc.src)
		})
	}
}

// TestParserImmediateOperandIsNotMistakenForADirective is the counterpart to
// the fix above: the "#" of an immediate operand must still be parsed as part
// of the instruction, not treated as the start of a new statement.
func TestParserImmediateOperandStillParses(t *testing.T) {
	src := "*=$0801\nstart:\n    lda #$01\n    ldx #$02\n    rts\n"
	assertNoErrors(t, src)
}

// TestParserReportsUnclosedPreprocessorBlock is the negative case: a missing
// #endif must still be reported.
func TestParserReportsUnclosedPreprocessorBlock(t *testing.T) {
	src := "#define DEBUG\n#if DEBUG\n    nop\n"
	_, diags := parseSource(t, src)

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "Unclosed #if") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing #endif was not reported; diagnostics: %v", messages(diags))
	}
}

// --- enum blocks ---------------------------------------------------------

// TestParserEnumBlocks covers .enum, which per the manual takes no name and
// defines its members as constants in the surrounding scope.
func TestParserEnumBlocks(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		symbols []string
	}{
		{
			name:    "explicit values",
			src:     ".enum {BLACK = 0, WHITE = 1, RED = 2}\n",
			symbols: []string{"BLACK", "WHITE", "RED"},
		},
		{
			name:    "implicit values",
			src:     ".enum {PLAYER, ENEMY, BULLET}\n",
			symbols: []string{"PLAYER", "ENEMY", "BULLET"},
		},
		{
			name:    "members spread over several lines",
			src:     ".enum {\n    IDLE = 0,\n    RUNNING = 1\n}\n",
			symbols: []string{"IDLE", "RUNNING"},
		},
		{
			// Comment tokens inside the block used to end up in the "expected
			// enum member identifier" error branch.
			name:    "trailing comments between members",
			src:     ".enum {\n    PLAYER,      // 0\n    ENEMY1,      // 1\n    BULLET       // 3\n}\n",
			symbols: []string{"PLAYER", "ENEMY1", "BULLET"},
		},
		{
			name:    "hex values",
			src:     ".enum {SCREEN = $0400, CHARSET = $2000}\n",
			symbols: []string{"SCREEN", "CHARSET"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertNoErrors(t, tc.src)
			scope, _ := parseSource(t, tc.src)
			assertHasSymbols(t, scope, tc.symbols...)
		})
	}
}

// TestParserEnumRequiresBrace is the negative case. .enum takes no name, so a
// name between the directive and the brace has to be reported.
func TestParserEnumRequiresBrace(t *testing.T) {
	_, diags := parseSource(t, ".enum Colors {BLACK = 0}\n")
	if len(errorsOnly(diags)) == 0 {
		t.Errorf("named enum was accepted, expected an error; diagnostics: %v", messages(diags))
	}
}

// --- multi labels --------------------------------------------------------

// TestAnonymousMultiLabels covers the anonymous form from manual 2.4. The lexer
// required a name after the '!', so "!:" and "!+" fell through to the boolean
// not operator. A "jmp !-" then consumed the following line as its operand,
// which cascaded into "Unexpected token" errors, undefined symbols and a flood
// of unreachable code warnings.
func TestAnonymousMultiLabels(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "skip forward over an instruction",
			src:  "*=$0801\n        bcc !+\n        eor #$c5\n!:      sta $02\n        rts\n",
		},
		{
			name: "loop backwards",
			src:  "*=$0801\n        ldx #0\n!:      inx\n        bne !-\n        rts\n",
		},
		{
			name: "named and anonymous side by side",
			src: "*=$0801\n" +
				"        ldx #0\n" +
				"!:      lda $1000,x\n" +
				"        beq !done+\n" +
				"        inx\n" +
				"        jmp !-\n" +
				"!done:\n" +
				"        rts\n",
		},
		{
			name: "manual example",
			src: "*=$0801\n" +
				"        ldx #10\n" +
				"!loop:\n" +
				"        jmp !+\n" +
				"        nop\n" +
				"        nop\n" +
				"!:      jmp !+\n" +
				"        nop\n" +
				"        nop\n" +
				"!:\n" +
				"        dex\n" +
				"        bne !loop-\n",
		},
		{
			name: "repeated sign",
			src:  "*=$0801\n        jmp !+++\n!:      nop\n!:      nop\n!:\n        rts\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertNoErrors(t, tc.src)
		})
	}
}

// TestBooleanNotIsNotAMultiLabel is the counterpart: making '!' start a multi
// label must not swallow the boolean not operator.
func TestBooleanNotIsNotAMultiLabel(t *testing.T) {
	tests := []string{
		"#define DEBUG\n#if !DEBUG\n    nop\n#endif\n",
		".const A = 1\n.const B = !A\n",
		".const A = 1\n.const B = A != 2\n",
	}

	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			assertNoErrors(t, src)
		})
	}
}

// TestMultiLabelSkipCount covers the sign counting. Repeating the sign skips
// instances, so "!+++" is not the same reference as "!+".
func TestMultiLabelSkipCount(t *testing.T) {
	tests := []struct {
		literal string
		want    int
	}{
		{"!loop+", 1},
		{"!loop-", 1},
		{"!+", 1},
		{"!-", 1},
		{"!++", 2},
		{"!+++", 3},
		{"!--", 2},
		{"!loop++", 2},
		{"!loop:", 1}, // a definition has no sign
		{"!:", 1},
		{"", 1},
	}

	for _, tc := range tests {
		t.Run(tc.literal, func(t *testing.T) {
			if got := multiLabelSkipCount(tc.literal); got != tc.want {
				t.Errorf("multiLabelSkipCount(%q) = %d, want %d", tc.literal, got, tc.want)
			}
		})
	}
}

// TestLookupMultiLabelSkip checks that the resolver steps over the requested
// number of instances instead of always taking the nearest one.
func TestLookupMultiLabelSkip(t *testing.T) {
	ctx := NewAnalysisContext()
	for _, addr := range []int64{0x1000, 0x1010, 0x1020, 0x1030} {
		ctx.DefinedMultiLabels[""] = append(ctx.DefinedMultiLabels[""],
			&Symbol{Kind: MultiLabel, Address: addr})
	}

	tests := []struct {
		name      string
		from      int64
		direction rune
		skip      int
		want      int64
		found     bool
	}{
		{"forward nearest", 0x1005, '+', 1, 0x1010, true},
		{"forward second", 0x1005, '+', 2, 0x1020, true},
		{"forward third", 0x1005, '+', 3, 0x1030, true},
		{"forward past the end", 0x1005, '+', 4, 0, false},
		{"backward nearest", 0x1025, '-', 1, 0x1020, true},
		{"backward second", 0x1025, '-', 2, 0x1010, true},
		{"backward third", 0x1025, '-', 3, 0x1000, true},
		{"backward past the start", 0x1025, '-', 4, 0, false},
		{"skip below one is treated as one", 0x1005, '+', 0, 0x1010, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sym, found := ctx.lookupMultiLabel("", tc.direction, tc.from, tc.skip)
			if found != tc.found {
				t.Fatalf("found = %v, want %v", found, tc.found)
			}
			if found && sym.Address != tc.want {
				t.Errorf("resolved to $%04X, want $%04X", sym.Address, tc.want)
			}
		})
	}
}

// --- symbol kinds --------------------------------------------------------

// TestBuildScopeSymbolKinds guards the rule that a symbol's kind follows from
// the directive that declared it. Deriving it from the shape of the statement
// instead used to drop two cases silently: ".label name = value" carries a
// value and therefore never reached the .label branch, and an enum member
// without an explicit value has neither a value nor a block.
func TestBuildScopeSymbolKinds(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		symbol string
		kind   SymbolKind
	}{
		{"const with value", ".const MAX = 10\n", "MAX", Constant},
		{"var with value", ".var counter = 0\n", "counter", Variable},
		{"label with value", ".label SCREEN = $0400\n", "SCREEN", Label},
		{"code label", "start:\n    rts\n", "start", Label},
		{"macro with block", ".macro clear() {\n    rts\n}\n", "clear", Macro},
		{"function with block", ".function f(x) {\n    .return x\n}\n", "f", Function},
		{"pseudocommand", ".pseudocommand mov src : dst {\n    rts\n}\n", "mov", PseudoCommand},
		{"enum member with value", ".enum {RED = 2}\n", "RED", Constant},
		{"enum member without value", ".enum {PLAYER}\n", "PLAYER", Constant},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope, _ := parseSource(t, tc.src)
			sym, ok := scope.FindSymbol(tc.symbol)
			if !ok {
				t.Fatalf("symbol %q was not defined by %q", tc.symbol, tc.src)
			}
			if sym.Kind != tc.kind {
				t.Errorf("symbol %q has kind %s, want %s", tc.symbol, sym.Kind.String(), tc.kind.String())
			}
		})
	}
}

// --- namespaces ----------------------------------------------------------

// TestNamespaceScopeRange guards the block end position. BlockStatement.EndToken
// was never assigned, so every namespace scope ended at line -1 and no line
// lookup ever matched it, which hid namespace local symbols from completion.
func TestNamespaceScopeRange(t *testing.T) {
	// 0: .namespace vic {
	// 1:     .label border = $d020
	// 2: }
	src := ".namespace vic {\n    .label border = $d020\n}\n"
	scope, _ := parseSource(t, src)

	ns := scope.FindNamespace("vic")
	if ns == nil {
		t.Fatal("namespace vic did not produce a child scope")
	}
	if ns.Range.Start.Line != 0 {
		t.Errorf("namespace starts at line %d, want 0", ns.Range.Start.Line)
	}
	if ns.Range.End.Line != 2 {
		t.Errorf("namespace ends at line %d, want 2 (the closing brace)", ns.Range.End.Line)
	}

	// A line inside the block has to resolve to the namespace scope, otherwise
	// completion inside a namespace never sees its own symbols.
	if got := scope.findInnermostScope(1); got != ns {
		t.Errorf("findInnermostScope(1) returned %q, want %q", got.Name, ns.Name)
	}
}

// TestNamespaceRedeclarationReusesScope: the manual (9.3) states that declaring
// a namespace a second time continues the existing one. It must not be reported
// as a duplicate symbol, and symbols from both blocks belong to one scope.
func TestNamespaceRedeclarationReusesScope(t *testing.T) {
	src := ".namespace part1 {\n" +
		"init:\n" +
		"    rts\n" +
		"}\n" +
		".namespace part1 {\n" +
		"exec:\n" +
		"    rts\n" +
		"}\n"

	assertNoErrors(t, src)

	scope, _ := parseSource(t, src)
	assertHasSymbols(t, scope, "part1.init", "part1.exec")

	count := 0
	for _, child := range scope.Children {
		if child.Name == "part1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("namespace part1 produced %d scopes, want 1", count)
	}
}

// TestDuplicateSymbolStillReported is the negative case for the change above:
// only namespaces may be redeclared, everything else keeps its error.
func TestDuplicateSymbolStillReported(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"duplicate label", "start:\n    rts\nstart:\n    rts\n"},
		{"duplicate constant", ".const MAX = 1\n.const MAX = 2\n"},
		{"duplicate macro", ".macro m() {\n    rts\n}\n.macro m() {\n    rts\n}\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := parseSource(t, tc.src)
			found := false
			for _, d := range errorsOnly(diags) {
				if strings.Contains(d.Message, "already defined") {
					found = true
				}
			}
			if !found {
				t.Errorf("duplicate definition was not reported; diagnostics: %v", messages(diags))
			}
		})
	}
}

// --- helpers -------------------------------------------------------------

func messages(diags []Diagnostic) []string {
	out := make([]string, 0, len(diags))
	for _, d := range diags {
		out = append(out, d.Message)
	}
	return out
}

// TestFunctionParametersAreSymbols covers hover and go-to-definition inside a
// function or macro body. The parameters used to exist only as a []string on
// the symbol, so a reference to one resolved to nothing.
func TestFunctionParametersAreSymbols(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		body   int // a line inside the body, zero based
		params []string
	}{
		{"function", ".function square(x) {\n    .return x * x\n}\n", 1, []string{"x"}},
		{"function with two parameters", ".function add(a, b) {\n    .return a + b\n}\n", 1, []string{"a", "b"}},
		{"macro", ".macro store(value, target) {\n    lda #value\n}\n", 1, []string{"value", "target"}},
		{"pseudocommand", ".pseudocommand mov src : dst {\n    lda src\n}\n", 1, []string{"src", "dst"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope, _ := parseSource(t, tc.src)

			// From the root the parameters must stay invisible; they belong to
			// the body, not to the file.
			for _, name := range tc.params {
				if _, found := scope.Symbols[name]; found {
					t.Errorf("parameter %q leaked into the outer scope", name)
				}
			}

			inner := scope.findInnermostScope(tc.body)
			if inner == scope {
				t.Fatalf("line %d did not resolve to a body scope", tc.body)
			}
			for _, name := range tc.params {
				symbol, found := inner.FindSymbol(name)
				if !found {
					t.Errorf("parameter %q is not a symbol in the body scope", name)
					continue
				}
				if symbol.Kind != Parameter {
					t.Errorf("parameter %q has kind %s, want parameter", name, symbol.Kind.String())
				}
			}
		})
	}
}
