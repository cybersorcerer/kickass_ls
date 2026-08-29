package lsp

import (
	"strings"
	"testing"
)

func format(t *testing.T, in string) string {
	t.Helper()
	out, err := FormatDocument(in, DefaultFormattingConfig())
	if err != nil {
		t.Fatalf("FormatDocument(%q) returned error: %v", in, err)
	}
	return out
}

// --- positive cases ------------------------------------------------------

// TestFormatDocumentLayout pins the layout the formatter produces. These are
// golden expectations: a deliberate change to indentation or comment alignment
// has to update them, an accidental one fails here.
func TestFormatDocumentLayout(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "instructions are indented, labels and origin stay at column 0",
			in:   "*=$0801\nstart:\nlda #$00\nrts\n",
			want: "*= $0801\nstart:\n        lda #$00\n        rts\n",
		},
		{
			name: "block contents get one extra indent level",
			in:   ".macro foo() {\nlda #$00\nrts\n}\n",
			want: ".macro foo() {\n                lda #$00\n                rts\n}\n",
		},
		{
			name: "end of line comments are aligned to the configured column",
			in:   "start:\nlda #$00 // laden\nrts ; zurueck\n",
			want: "start:\n        lda #$00                        // laden\n        rts                             ; zurueck\n",
		},
		{
			name: "blank lines are preserved",
			in:   "a:\n\nb:\n\n",
			want: "a:\n\nb:\n\n",
		},
		{
			name: "empty document stays empty",
			in:   "",
			want: "",
		},
		{
			name: "a lone newline is preserved",
			in:   "\n",
			want: "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := format(t, tc.in); got != tc.want {
				t.Errorf("FormatDocument(%q)\n  got  %q\n  want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatDocumentDisabledReturnsInputUnchanged(t *testing.T) {
	cfg := DefaultFormattingConfig()
	cfg.Enabled = false

	in := "start:\nlda #$00\n"
	got, err := FormatDocument(in, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != in {
		t.Errorf("with Enabled=false the document must not change\n  got  %q\n  want %q", got, in)
	}
}

// --- negative cases ------------------------------------------------------

// TestFormatDocumentPreservesStringLiterals guards the defect where the comment
// scanner looked for // and ; anywhere in the line. A string containing either
// was split into "code" and "comment", padded with alignment whitespace on
// reassembly, and the result no longer assembled.
func TestFormatDocumentPreservesStringLiterals(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		keepAll []string
	}{
		{
			name:    "double slash inside a string",
			in:      "        .text \"a // b\"\n",
			keepAll: []string{`"a // b"`},
		},
		{
			name:    "semicolon inside a string",
			in:      "        .text \"x;y\"\n",
			keepAll: []string{`"x;y"`},
		},
		{
			name:    "url inside a string",
			in:      "        .text \"http://example.com\"\n",
			keepAll: []string{`"http://example.com"`},
		},
		{
			name:    "string followed by a real comment",
			in:      "        .text \"a;b\" // note\n",
			keepAll: []string{`"a;b"`, "// note"},
		},
		{
			name:    "escaped quote inside a string",
			in:      "        .text \"a\\\"b;c\"\n",
			keepAll: []string{`"a\"b;c"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := format(t, tc.in)
			for _, want := range tc.keepAll {
				if !strings.Contains(got, want) {
					t.Errorf("formatting destroyed %s\n  in:  %q\n  out: %q", want, tc.in, got)
				}
			}
		})
	}
}

// TestFormatDocumentTrailingNewline guards the defect where every branch of the
// reconstruction terminated its line instead of separating lines. A document
// ending in a newline grew by one byte on every format, so "format on save"
// appended a blank line each time.
func TestFormatDocumentTrailingNewline(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"document ending in a newline", "*=$0801\nstart:\nrts\n"},
		{"document without a trailing newline", "*=$0801\nstart:\nrts"},
		{"document ending in a blank line", "*=$0801\nrts\n\n"},
		{"comment only document", "// nur ein kommentar\n"},
		{"block end as last line", ".macro m() {\nrts\n}\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := format(t, tc.in)
			wantN := strings.Count(tc.in, "\n")
			gotN := strings.Count(got, "\n")
			if gotN != wantN {
				t.Errorf("newline count changed: got %d, want %d\n  in:  %q\n  out: %q",
					gotN, wantN, tc.in, got)
			}
		})
	}
}

// TestFormatDocumentIsIdempotent is the property that catches the whole class
// of "formatting slowly rewrites the file" defects: formatting an already
// formatted document must be a no-op.
func TestFormatDocumentIsIdempotent(t *testing.T) {
	inputs := []string{
		"",
		"\n",
		"*=$0801\nstart:\nlda #$00\nrts\n",
		"*=$0801\nstart:\nlda #$00\nrts",
		".macro foo() {\nlda #$00\nrts\n}\n",
		"start:\nlda #$00 // laden\nrts ; zurueck\n",
		"        .text \"a // b\"\n",
		"a:\n\nb:\n\n",
		"// nur ein kommentar\n",
		".namespace n {\nlabel:\nrts\n}\n",
	}

	for _, in := range inputs {
		t.Run(strings.ReplaceAll(in, "\n", "\\n"), func(t *testing.T) {
			once := format(t, in)
			twice := format(t, once)
			if once != twice {
				t.Errorf("second format changed the document\n  1st: %q\n  2nd: %q", once, twice)
			}
		})
	}
}

// --- findCommentStart ----------------------------------------------------

// findCommentStart is the helper the formatter and the reference search rely on
// to tell a comment from a comment marker inside a string. It is tested
// directly because both callers break in different, hard to diagnose ways when
// it is wrong.
func TestFindCommentStart(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		// positive: a comment is present and must be found
		{"double slash comment", "        lda #$00 // note", 17},
		{"semicolon comment", "        lda #$00 ; note", 17},
		{"comment at start of line", "// note", 0},
		{"comment directly after a string", `.text "abc" ; note`, 12},
		{"semicolon before a later double slash", "lda ; a // b", 4},

		// negative: no comment, or only comment markers inside a literal
		{"no comment at all", "        lda #$00", -1},
		{"double slash inside a string", `.text "a // b"`, -1},
		{"semicolon inside a string", `.text "x;y"`, -1},
		{"url inside a string", `.text "http://example.com"`, -1},
		{"escaped quote does not end the string", `.text "a\"b;c"`, -1},
		{"single quoted character literal", "lda #';'", -1},
		{"empty line", "", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := findCommentStart(tc.line); got != tc.want {
				t.Errorf("findCommentStart(%q) = %d, want %d", tc.line, got, tc.want)
			}
		})
	}
}
