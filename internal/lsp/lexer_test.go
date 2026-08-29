package lsp

import (
	"strings"
	"testing"
)

// wantToken is the part of a ContextToken the tests assert on.
type wantToken struct {
	typ     TokenType
	literal string
	line    int
	column  int
}

// lexAll runs the lexer to EOF and returns every token. The iteration is bounded
// so a lexer that stops making progress fails the test instead of hanging it.
func lexAll(t *testing.T, input string) []*ContextToken {
	t.Helper()

	l := NewContextAwareLexer(input, "file:///test.asm", GetProcessorContext())
	var got []*ContextToken

	limit := len(input)*2 + 16
	for i := 0; ; i++ {
		if i > limit {
			t.Fatalf("lexer did not reach EOF after %d tokens for input %q", limit, input)
		}
		tok := l.NextToken()
		if tok.Type == TOKEN_EOF {
			return got
		}
		got = append(got, tok)
	}
}

func assertTokens(t *testing.T, input string, want []wantToken) {
	t.Helper()

	got := lexAll(t, input)
	if len(got) != len(want) {
		t.Errorf("input %q: got %d tokens, want %d", input, len(got), len(want))
		for i, g := range got {
			t.Logf("  got[%d] = {%s, %q, L%d, C%d}", i, g.Type.String(), g.Literal, g.Line, g.Column)
		}
		return
	}

	for i, w := range want {
		g := got[i]
		if g.Type != w.typ || g.Literal != w.literal || g.Line != w.line || g.Column != w.column {
			t.Errorf("input %q token %d:\n  got  {%s, %q, L%d, C%d}\n  want {%s, %q, L%d, C%d}",
				input, i,
				g.Type.String(), g.Literal, g.Line, g.Column,
				w.typ.String(), w.literal, w.line, w.column)
		}
	}
}

// --- positive cases ------------------------------------------------------

func TestLexerValidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []wantToken
	}{
		{
			name:  "instruction with immediate hex operand",
			input: "        lda #$05",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_NUMBER_HEX, "$05", 1, 14},
			},
		},
		{
			name:  "label and jump",
			input: "start:  jmp start",
			want: []wantToken{
				{TOKEN_LABEL, "start:", 1, 1},
				{TOKEN_MNEMONIC_CTRL, "jmp", 1, 9},
				{TOKEN_IDENTIFIER, "start", 1, 13},
			},
		},
		{
			name:  "indirect indexed addressing",
			input: "        lda ($fe),y",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_LPAREN, "(", 1, 13},
				{TOKEN_NUMBER_HEX, "$fe", 1, 14},
				{TOKEN_RPAREN, ")", 1, 17},
				{TOKEN_COMMA, ",", 1, 18},
				{TOKEN_IDENTIFIER, "y", 1, 19},
			},
		},
		{
			name:  "multi label definition and backward reference",
			input: "!loop:  bne !loop-",
			want: []wantToken{
				{TOKEN_MULTILABEL, "!loop:", 1, 1},
				{TOKEN_MNEMONIC_STD, "bne", 1, 9},
				{TOKEN_MULTILABEL_BACK, "!loop-", 1, 13},
			},
		},
		{
			// Anonymous multi labels have no name at all (manual 2.4).
			name:  "anonymous multi label definition",
			input: "!:      sta $d020",
			want: []wantToken{
				{TOKEN_MULTILABEL, "!:", 1, 1},
				{TOKEN_MNEMONIC_STD, "sta", 1, 9},
				{TOKEN_NUMBER_HEX, "$d020", 1, 13},
			},
		},
		{
			name:  "anonymous forward reference",
			input: "        bcc !+",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "bcc", 1, 9},
				{TOKEN_MULTILABEL_FWD, "!+", 1, 13},
			},
		},
		{
			name:  "anonymous backward reference",
			input: "        jmp !-",
			want: []wantToken{
				{TOKEN_MNEMONIC_CTRL, "jmp", 1, 9},
				{TOKEN_MULTILABEL_BACK, "!-", 1, 13},
			},
		},
		{
			// Repeating the sign skips labels: !+++ refers to the third one.
			name:  "repeated sign skips labels",
			input: "        jmp !+++",
			want: []wantToken{
				{TOKEN_MNEMONIC_CTRL, "jmp", 1, 9},
				{TOKEN_MULTILABEL_FWD, "!+++", 1, 13},
			},
		},
		{
			// The index register check used to consume the first letter without
			// putting it back, so any operand starting with X or Y lost it.
			name:  "operand starting with Y is not the index register",
			input: "        lda #YELLOW",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_IDENTIFIER, "YELLOW", 1, 14},
			},
		},
		{
			name:  "operand starting with X is not the index register",
			input: "        lda XCOORD",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_IDENTIFIER, "XCOORD", 1, 13},
			},
		},
		{
			name:  "underscore after Y does not make it a register",
			input: "        lda Y_POS",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_IDENTIFIER, "Y_POS", 1, 13},
			},
		},
		{
			name:  "the index register itself still works",
			input: "        lda $d020,y",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_NUMBER_HEX, "$d020", 1, 13},
				{TOKEN_COMMA, ",", 1, 18},
				{TOKEN_IDENTIFIER, "y", 1, 19},
			},
		},
		{
			name:  "boolean not is not a multi label",
			input: "#if !DEBUG",
			want: []wantToken{
				{TOKEN_DIRECTIVE_KICK_PRE, "#if", 1, 1},
				{TOKEN_LOGICAL_NOT, "!", 1, 5},
				{TOKEN_IDENTIFIER, "DEBUG", 1, 6},
			},
		},
		{
			name:  "not equal is not a multi label",
			input: "a = 1 != 2",
			want: []wantToken{
				{TOKEN_IDENTIFIER, "a", 1, 1},
				{TOKEN_EQUAL, "=", 1, 3},
				{TOKEN_NUMBER_DEC, "1", 1, 5},
				{TOKEN_NOT_EQUAL, "!=", 1, 7},
				{TOKEN_NUMBER_DEC, "2", 1, 10},
			},
		},
		{
			name:  "all number bases",
			input: "        lda #$ff",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_NUMBER_HEX, "$ff", 1, 14},
			},
		},
		{
			name:  "binary and octal literals in a directive",
			input: ".byte %1010, &777, 42",
			want: []wantToken{
				{TOKEN_DIRECTIVE_KICK_PRE, ".byte", 1, 1},
				{TOKEN_NUMBER_BIN, "%1010", 1, 7},
				{TOKEN_COMMA, ",", 1, 12},
				{TOKEN_NUMBER_OCT, "&777", 1, 14},
				{TOKEN_COMMA, ",", 1, 18},
				{TOKEN_NUMBER_DEC, "42", 1, 20},
			},
		},
		{
			name:  "string literal keeps embedded comment markers",
			input: `        .text "a // b"`,
			want: []wantToken{
				{TOKEN_DIRECTIVE_KICK_PRE, ".text", 1, 9},
				{TOKEN_STRING, `"a // b"`, 1, 15},
			},
		},
		{
			name:  "line comment with slashes",
			input: "        lda $d020 // note",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_NUMBER_HEX, "$d020", 1, 13},
				{TOKEN_COMMENT, "// note", 1, 19},
			},
		},
		{
			name:  "line comment with semicolon",
			input: "        lda $d020 ; note",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_NUMBER_HEX, "$d020", 1, 13},
				{TOKEN_COMMENT, "; note", 1, 19},
			},
		},
		{
			name:  "bitwise and modulo operators next to numbers",
			input: "a = 1 & 2 % 3 | 4",
			want: []wantToken{
				{TOKEN_IDENTIFIER, "a", 1, 1},
				{TOKEN_EQUAL, "=", 1, 3},
				{TOKEN_NUMBER_DEC, "1", 1, 5},
				{TOKEN_BITWISE_AND, "&", 1, 7},
				{TOKEN_NUMBER_DEC, "2", 1, 9},
				{TOKEN_MODULO, "%", 1, 11},
				{TOKEN_NUMBER_DEC, "3", 1, 13},
				{TOKEN_BITWISE_OR, "|", 1, 15},
				{TOKEN_NUMBER_DEC, "4", 1, 17},
			},
		},
		{
			name:  "builtin names stay identifiers so user symbols can shadow them",
			input: ".const RED = 2",
			want: []wantToken{
				{TOKEN_DIRECTIVE_KICK_PRE, ".const", 1, 1},
				{TOKEN_IDENTIFIER, "RED", 1, 8},
				{TOKEN_EQUAL, "=", 1, 12},
				{TOKEN_NUMBER_DEC, "2", 1, 14},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertTokens(t, tc.input, tc.want)
		})
	}
}

func TestLexerTracksLinesAcrossNewlines(t *testing.T) {
	input := "start:\n        lda #$01\n        rts\n"
	want := []wantToken{
		{TOKEN_LABEL, "start:", 1, 1},
		{TOKEN_MNEMONIC_STD, "lda", 2, 9},
		{TOKEN_HASH, "#", 2, 13},
		{TOKEN_NUMBER_HEX, "$01", 2, 14},
		{TOKEN_MNEMONIC_STD, "rts", 3, 9},
	}
	assertTokens(t, input, want)
}

// --- negative cases ------------------------------------------------------

// TestLexerMalformedInput checks that broken source produces defined tokens
// rather than a panic, a hang, or silently shifted positions. Every case here
// is source a user can type while still editing.
func TestLexerMalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []wantToken
	}{
		{
			name:  "hex literal with a non hex digit is reported as illegal",
			input: "        lda #$0G",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_ILLEGAL, "$0G", 1, 14},
			},
		},
		{
			name:  "dollar sign without digits is illegal, not a number",
			input: "        lda #$",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_ILLEGAL, "$", 1, 14},
			},
		},
		{
			// The loop used to stop one character early, so the last one was
			// lexed again as a separate token.
			name:  "unterminated block comment consumes the rest of the input",
			input: "        /* unterminated",
			want: []wantToken{
				{TOKEN_COMMENT_BLOCK, "/* unterminated", 1, 9},
			},
		},
		{
			name:  "terminated block comment",
			input: "        /* note */ nop",
			want: []wantToken{
				{TOKEN_COMMENT_BLOCK, "/* note */", 1, 9},
				{TOKEN_MNEMONIC_STD, "nop", 1, 20},
			},
		},
		{
			name:  "unterminated character literal does not swallow the rest of the line",
			input: "        lda #'a",
			want: []wantToken{
				{TOKEN_MNEMONIC_STD, "lda", 1, 9},
				{TOKEN_HASH, "#", 1, 13},
				{TOKEN_ILLEGAL, "'", 1, 14},
				{TOKEN_IDENTIFIER, "a", 1, 15},
			},
		},
		{
			name:  "percent without binary digits is the modulo operator",
			input: "a = 5 % 3",
			want: []wantToken{
				{TOKEN_IDENTIFIER, "a", 1, 1},
				{TOKEN_EQUAL, "=", 1, 3},
				{TOKEN_NUMBER_DEC, "5", 1, 5},
				{TOKEN_MODULO, "%", 1, 7},
				{TOKEN_NUMBER_DEC, "3", 1, 9},
			},
		},
		{
			name:  "ampersand without octal digits is bitwise and",
			input: "a = 5 & 3",
			want: []wantToken{
				{TOKEN_IDENTIFIER, "a", 1, 1},
				{TOKEN_EQUAL, "=", 1, 3},
				{TOKEN_NUMBER_DEC, "5", 1, 5},
				{TOKEN_BITWISE_AND, "&", 1, 7},
				{TOKEN_NUMBER_DEC, "3", 1, 9},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertTokens(t, tc.input, tc.want)
		})
	}
}

func TestLexerTerminatesOnTruncatedInput(t *testing.T) {
	// These must not hang or panic. The token stream itself is not asserted;
	// the point is that the lexer always reaches EOF.
	inputs := []string{
		"",
		" ",
		"\n\n\n",
		`        .text "unterminated`,
		"        lda (",
		".macro foo() {",
		"$",
		"#",
		"!",
		"'",
	}

	for _, in := range inputs {
		t.Run(strings.TrimSpace(in), func(t *testing.T) {
			lexAll(t, in) // fails the test if EOF is never reached
		})
	}
}

// TestLexerColumnsMatchSource is the general guard against position drift.
// Whenever the lexer backtracks out of a candidate token it must undo the
// column along with the position; forgetting one of the two shifts every
// later token on the line and corrupts diagnostics and semantic highlighting.
//
// The check is generic: for every token whose literal appears verbatim in the
// source, the reported column must point at it. Tokens whose literal is not a
// substring of the line (character literals are converted to their numeric
// value) are skipped.
func TestLexerColumnsMatchSource(t *testing.T) {
	inputs := []string{
		"a = 1 & 2 % 3 | 4",
		"        lda #$05",
		"        lda #$",
		"        lda #'a",
		".byte %1010, &777, 42",
		"start:  jmp start",
		"        lda ($fe),y",
		".var bb = aa & 5678",
		".var cc = aa % 3",
		"!loop:  bne !loop-",
		"        lda #YELLOW",
		"        lda XCOORD,y",
		"        lda Y_POS",
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			lines := strings.Split(in, "\n")
			for _, tok := range lexAll(t, in) {
				if tok.Line < 1 || tok.Line > len(lines) {
					t.Errorf("token %q reports line %d, input has %d lines", tok.Literal, tok.Line, len(lines))
					continue
				}
				line := lines[tok.Line-1]
				if !strings.Contains(line, tok.Literal) {
					continue // literal was rewritten by the lexer, position is not comparable
				}
				start := tok.Column - 1
				end := start + len(tok.Literal)
				if start < 0 || end > len(line) {
					t.Errorf("token %q at column %d does not fit into line %q", tok.Literal, tok.Column, line)
					continue
				}
				if got := line[start:end]; got != tok.Literal {
					t.Errorf("column drift: token %q reports column %d, but the source has %q there\n  line: %q",
						tok.Literal, tok.Column, got, line)
				}
			}
		})
	}
}

// TestDirectiveCategorisationIsFlat documents a known gap rather than a bug in
// the lexer: kickass.json carries no "category" field for directives, so the
// switch in tokenizeDirective always falls through to its default and every
// directive becomes TOKEN_DIRECTIVE_KICK_PRE. The four other directive token
// types are currently unreachable.
//
// If categories are added to kickass.json, this test starts failing — that is
// the signal to update it and the expectations in TestLexerValidInput.
func TestDirectiveCategorisationIsFlat(t *testing.T) {
	for _, in := range []string{".byte 1", ".if (1)", ".macro m() {", ".text \"x\"", ".const A = 1"} {
		got := lexAll(t, in)
		if len(got) == 0 {
			t.Fatalf("input %q produced no tokens", in)
		}
		if got[0].Type != TOKEN_DIRECTIVE_KICK_PRE {
			t.Errorf("input %q: first token is %s, expected TOKEN_DIRECTIVE_KICK_PRE.\n"+
				"If kickass.json now has directive categories, update this test and TestLexerValidInput.",
				in, got[0].Type.String())
		}
	}
}
