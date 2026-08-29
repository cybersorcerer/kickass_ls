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

// --- getTokenAtPosition --------------------------------------------------

// TestGetTokenAtPosition covers the token hover and go-to-definition work on.
// Multi labels need their own recognition: neither '!' nor '+' nor '-' is a
// word character, so the anonymous forms were invisible and go-to-definition
// on "!+" or "!-" silently did nothing.
func TestGetTokenAtPosition(t *testing.T) {
	tests := []struct {
		name string
		line string
		char int
		want string
	}{
		// anonymous multi labels
		{"cursor on the bang of a forward reference", "        bcc !+", 12, "!+"},
		{"cursor on the sign of a forward reference", "        bcc !+", 13, "!+"},
		{"anonymous backward reference", "        jmp !-", 13, "!-"},
		{"anonymous definition", "!:      sta $02", 0, "!:"},
		{"repeated sign", "        jmp !+++", 14, "!+++"},

		// named multi labels
		{"named definition", "!loop:  inc $d020", 3, "!loop:"},
		{"named backward reference", "        bne !loop-", 15, "!loop-"},
		{"named forward reference", "        beq !done+", 15, "!done+"},

		// not multi labels
		{"boolean not", "#if !DEBUG", 6, "DEBUG"},
		{"not equal", "a != 2", 2, ""},
		{"plain label", "start:  rts", 2, "start"},
		{"mnemonic", "        lda #$01", 9, "lda"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := getTokenAtPosition(tc.line, tc.char); got != tc.want {
				t.Errorf("getTokenAtPosition(%q, %d) = %q, want %q", tc.line, tc.char, got, tc.want)
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
