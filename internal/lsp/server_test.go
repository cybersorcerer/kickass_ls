package lsp

import "testing"

// --- getWordAtPosition ---------------------------------------------------

// TestGetWordAtPosition covers the word boundaries hover and go-to-definition
// rely on. The tricky character is '#': it belongs to the word for preprocessor
// statements but marks an immediate operand everywhere else, where including it
// made the lookup search for "#MYCONST" and find nothing.
func TestGetWordAtPosition(t *testing.T) {
	tests := []struct {
		name string
		line string
		char int
		want string
	}{
		// positive: ordinary identifiers, labels and directives
		{"identifier", "        lda MYCONST", 14, "MYCONST"},
		{"label definition", "start:", 2, "start"},
		{"directive keeps its dot", ".byte 1, 2", 2, ".byte"},
		{"qualified name keeps its dot", "        lda Colors.RED", 20, "Colors.RED"},
		{"mnemonic", "        lda #$05", 9, "lda"},

		// the '#' rule
		{"preprocessor statement keeps the hash", `#import "file.asm"`, 3, "#import"},
		{"indented preprocessor statement keeps the hash", `    #importonce`, 8, "#importonce"},
		{"immediate operand drops the hash", "        lda #MYCONST", 14, "MYCONST"},
		{"cursor on the hash of an immediate operand", "        lda #MYCONST", 12, "MYCONST"},
		{"cursor on the hash of a preprocessor statement", "#define DEBUG", 0, "#define"},

		// negative: nothing to return
		{"position past end of line", "lda", 10, ""},
		{"negative position", "lda", -1, ""},
		{"empty line", "", 0, ""},
		{"hash with nothing after it", "        lda #", 12, ""},
		{"whitespace only", "     ", 2, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := getWordAtPosition(tc.line, tc.char); got != tc.want {
				t.Errorf("getWordAtPosition(%q, %d) = %q, want %q", tc.line, tc.char, got, tc.want)
			}
		})
	}
}

// --- GenerateSignatureHelp -----------------------------------------------

// TestGenerateSignatureHelpBounds guards an out of range read: the cursor was
// clamped to len(line) and the backwards scan then indexed line[len(line)].
func TestGenerateSignatureHelpBounds(t *testing.T) {
	scope, _ := parseSource(t, ".function f(x) {\n    .return x\n}\n")

	lines := []string{
		"",
		"        lda #$00",
		"        .var v = f(1",
		"        .var v = f(1)",
		"(",
		")",
	}

	for _, line := range lines {
		for _, char := range []int{-1, 0, len(line), len(line) + 1, len(line) + 50} {
			// A panic here fails the test; the return values are not asserted
			// because they depend on the line contents.
			GenerateSignatureHelp(scope, line, char)
		}
	}
}
