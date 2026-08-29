package lsp

import (
	"strings"
	"testing"
)

// mapProvider serves documents from memory so the import graph and the
// splicing can be tested without touching the filesystem.
type mapProvider map[string]string

func (m mapProvider) Read(uri string) (string, bool) {
	text, ok := m[uri]
	return text, ok
}

// newTestWorkspace builds a workspace over in-memory documents. The keys are
// plain paths; they are turned into file URIs so resolution behaves as it does
// against a real project.
func newTestWorkspace(t *testing.T, files map[string]string) (*Workspace, map[string]string) {
	t.Helper()

	provider := mapProvider{}
	uris := map[string]string{}
	for name, text := range files {
		uri := "file:///project/" + name
		provider[uri] = text
		uris[name] = uri
	}

	// An empty root disables the filesystem scan; the documents are loaded
	// explicitly instead.
	ws := NewWorkspace("", provider)
	for _, uri := range uris {
		ws.Load(uri)
	}
	// Resolution matches by base name against already loaded files, so a second
	// pass gives every file its edges once all of them are known.
	for _, uri := range uris {
		ws.Invalidate(uri)
		ws.Load(uri)
	}
	return ws, uris
}

// unitMnemonics returns the mnemonics of a translation unit in order, which is
// enough to tell whether the statements were spliced in the right place.
func unitMnemonics(program *Program) []string {
	var out []string
	for _, stmt := range program.Statements {
		if instr, ok := stmt.(*InstructionStatement); ok {
			out = append(out, strings.ToLower(instr.Token.Literal))
		}
	}
	return out
}

// --- splicing ------------------------------------------------------------

func TestAssembleUnitSplicesInSourceOrder(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm":   "*=$0801\n        nop\n#import \"middle.asm\"\n        rts\n",
		"middle.asm": "        inx\n#import \"leaf.asm\"\n        iny\n",
		"leaf.asm":   "        dex\n",
	})

	program, members, diags := ws.AssembleUnit(uris["main.asm"])

	if errs := errorsOnly(diags); len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %s", e.Message)
		}
	}

	want := []string{"nop", "inx", "dex", "iny", "rts"}
	got := unitMnemonics(program)
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}

	if len(members) != 3 {
		t.Errorf("unit has %d members, want 3: %v", len(members), members)
	}
}

func TestAssembleUnitReportsMissingImport(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm": "*=$0801\n#import \"nowhere.asm\"\n        rts\n",
	})

	_, _, diags := ws.AssembleUnit(uris["main.asm"])

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "not found") {
			found = true
			if d.URI != uris["main.asm"] {
				t.Errorf("diagnostic points at %q, want %q", d.URI, uris["main.asm"])
			}
		}
	}
	if !found {
		t.Errorf("missing import was not reported: %v", messages(diags))
	}
}

func TestAssembleUnitDetectsCycle(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"a.asm": "        nop\n#import \"b.asm\"\n",
		"b.asm": "        inx\n#import \"a.asm\"\n",
	})

	_, _, diags := ws.AssembleUnit(uris["a.asm"])

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "cycle") {
			found = true
		}
	}
	if !found {
		t.Errorf("import cycle was not reported: %v", messages(diags))
	}
}

func TestAssembleUnitHonoursImportOnce(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm": "#import \"lib.asm\"\n#import \"lib.asm\"\n        rts\n",
		"lib.asm":  "#importonce\n        nop\n",
	})

	program, _, _ := ws.AssembleUnit(uris["main.asm"])

	count := 0
	for _, m := range unitMnemonics(program) {
		if m == "nop" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("lib.asm contributed %d times, want 1 (it declares #importonce)", count)
	}
}

func TestAssembleUnitWithoutImportOnceIncludesTwice(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm": "#import \"lib.asm\"\n#import \"lib.asm\"\n        rts\n",
		"lib.asm":  "        nop\n",
	})

	program, _, _ := ws.AssembleUnit(uris["main.asm"])

	count := 0
	for _, m := range unitMnemonics(program) {
		if m == "nop" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("lib.asm contributed %d times, want 2 without #importonce", count)
	}
}

// --- graph ---------------------------------------------------------------

func TestRootsFor(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm":      "#import \"constants.asm\"\n        rts\n",
		"constants.asm": ".const MAX = 10\n",
		"scratch.asm":   "        nop\n",
	})

	tests := []struct {
		file string
		want string
	}{
		{"main.asm", "main.asm"},       // imports, is imported by nobody
		{"constants.asm", "main.asm"},  // analysed as part of its importer
		{"scratch.asm", "scratch.asm"}, // nobody imports it, so it is its own unit
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			got := ws.PrimaryRootFor(uris[tc.file])
			if got != uris[tc.want] {
				t.Errorf("PrimaryRootFor(%s) = %s, want %s", tc.file, got, uris[tc.want])
			}
		})
	}
}

func TestPrimaryRootIsDeterministicWithTwoImporters(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"zeta.asm":   "#import \"shared.asm\"\n        rts\n",
		"alpha.asm":  "#import \"shared.asm\"\n        rts\n",
		"shared.asm": ".const MAX = 10\n",
	})

	roots := ws.RootsFor(uris["shared.asm"])
	if len(roots) != 2 {
		t.Fatalf("shared.asm has %d roots, want 2: %v", len(roots), roots)
	}
	if got := ws.PrimaryRootFor(uris["shared.asm"]); got != uris["alpha.asm"] {
		t.Errorf("primary root is %s, want the lowest path %s", got, uris["alpha.asm"])
	}
}

// --- cross file symbols --------------------------------------------------

// TestImportedSymbolsResolve is the reason for the whole feature: a constant
// defined in an imported file must be known in the importer, so that it neither
// counts as unused nor as undefined.
func TestImportedSymbolsResolve(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm":      "*=$0801\n#import \"constants.asm\"\nstart:\n        lda #SCREEN_WIDTH\n        rts\n",
		"constants.asm": ".const SCREEN_WIDTH = 40\n",
	})

	program, _, diags := ws.AssembleUnit(uris["main.asm"])
	scope, defs := buildScopeFromAST(program, uris["main.asm"])
	analyzer := NewSemanticAnalyzerForUnit(scope, ws.UnitSources(uris["main.asm"]))
	diags = append(diags, defs...)
	diags = append(diags, analyzer.Analyze(program)...)

	if _, ok := scope.FindSymbol("SCREEN_WIDTH"); !ok {
		t.Fatal("SCREEN_WIDTH from constants.asm is not visible in the unit")
	}

	for _, d := range diags {
		if strings.Contains(d.Message, "SCREEN_WIDTH") {
			t.Errorf("imported constant still reported: %s", d.Message)
		}
	}
}

// TestDiagnosticsCarryTheirOwnFile guards the routing: a problem in the
// imported file must not be reported against the root.
func TestDiagnosticsCarryTheirOwnFile(t *testing.T) {
	ws, uris := newTestWorkspace(t, map[string]string{
		"main.asm": "*=$0801\n#import \"lib.asm\"\nstart:\n        rts\n",
		"lib.asm":  ".const UNUSED_HERE = 1\n",
	})

	program, _, _ := ws.AssembleUnit(uris["main.asm"])
	scope, _ := buildScopeFromAST(program, uris["main.asm"])
	analyzer := NewSemanticAnalyzerForUnit(scope, ws.UnitSources(uris["main.asm"]))

	for _, d := range analyzer.Analyze(program) {
		if strings.Contains(d.Message, "UNUSED_HERE") {
			if d.URI != uris["lib.asm"] {
				t.Errorf("diagnostic for a symbol from lib.asm carries URI %q, want %q",
					d.URI, uris["lib.asm"])
			}
			return
		}
	}
	t.Skip("no diagnostic produced for the unused constant; routing not exercised")
}

// --- URI helpers ---------------------------------------------------------

func TestURIPathRoundTrip(t *testing.T) {
	tests := []struct {
		uri  string
		path string
	}{
		{"file:///project/src/main.asm", "/project/src/main.asm"},
		{"file:///with%20space/main.asm", "/with space/main.asm"},
		{"/plain/path.asm", "/plain/path.asm"},
	}

	for _, tc := range tests {
		t.Run(tc.uri, func(t *testing.T) {
			if got := URIToPath(tc.uri); got != tc.path {
				t.Errorf("URIToPath(%q) = %q, want %q", tc.uri, got, tc.path)
			}
		})
	}

	if got := PathToURI("/project/src/main.asm"); got != "file:///project/src/main.asm" {
		t.Errorf("PathToURI = %q", got)
	}
	if got := PathToURI("file:///already/a/uri.asm"); got != "file:///already/a/uri.asm" {
		t.Errorf("PathToURI must not double-prefix: %q", got)
	}
}

// TestParseNumericValueHandlesHex guards a helper that reported every hex
// address as "not numeric": strconv.Atoi parses base 10 only, so prefixing the
// string with "0x" made it fail rather than succeed.
func TestParseNumericValueHandlesHex(t *testing.T) {
	ctx := GetProcessorContext()
	if ctx == nil {
		t.Fatal("no processor context")
	}

	tests := []struct {
		input string
		want  int
	}{
		{"$d020", 0xd020},
		{"$80", 0x80},
		{"0xd020", 0xd020},
		{"%1010", 0b1010},
		{"42", 42},
		{"#$10", 0x10},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ctx.parseNumericValue(tc.input)
			if err != nil {
				t.Fatalf("parseNumericValue(%q) failed: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseNumericValue(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}

	if _, err := ctx.parseNumericValue("someLabel"); err == nil {
		t.Error("a label must not parse as a number")
	}
}
