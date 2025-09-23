package lsp

import (
	"regexp"

	log "c64.nvim/internal/log"
)

// TokenType represents the type of a token.
type TokenType int

// Token represents a single token parsed from the input.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF

	// Literals
	TOKEN_LABEL
	TOKEN_IDENTIFIER // For identifiers that are not yet classified as labels, mnemonics, etc.

	// Comments
	TOKEN_COMMENT

	// Values
	TOKEN_NUMBER_HEX
	TOKEN_NUMBER_BIN
	TOKEN_NUMBER_DEC
	TOKEN_NUMBER_OCT
	TOKEN_STRING

	// Mnemonics
	TOKEN_MNEMONIC_STD
	TOKEN_MNEMONIC_CTRL
	TOKEN_MNEMONIC_ILL
	TOKEN_MNEMONIC_65C02

	// Directives
	TOKEN_DIRECTIVE_PC
	TOKEN_DIRECTIVE_KICK_PRE
	TOKEN_DIRECTIVE_KICK_FLOW
	TOKEN_DIRECTIVE_KICK_ASM
	TOKEN_DIRECTIVE_KICK_DATA
	TOKEN_DIRECTIVE_KICK_TEXT

	// Punctuation
	TOKEN_COLON    // :
	TOKEN_HASH     // #
	TOKEN_DOT      // .
	TOKEN_COMMA    // ,
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_ASTERISK // *
	TOKEN_SLASH    // /
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_EQUAL    // =
	TOKEN_LESS     // <
	TOKEN_GREATER  // >
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:             "ILLEGAL",
	TOKEN_EOF:                 "EOF",
	TOKEN_LABEL:               "LABEL",
	TOKEN_IDENTIFIER:          "IDENTIFIER",
	TOKEN_COMMENT:             "COMMENT",
	TOKEN_NUMBER_HEX:          "NUMBER_HEX",
	TOKEN_NUMBER_BIN:          "NUMBER_BIN",
	TOKEN_NUMBER_DEC:          "NUMBER_DEC",
	TOKEN_NUMBER_OCT:          "NUMBER_OCT",
	TOKEN_STRING:              "STRING",
	TOKEN_MNEMONIC_STD:        "MNEMONIC_STD",
	TOKEN_MNEMONIC_CTRL:       "MNEMONIC_CTRL",
	TOKEN_MNEMONIC_ILL:        "MNEMONIC_ILL",
	TOKEN_MNEMONIC_65C02:      "MNEMONIC_65C02",
	TOKEN_DIRECTIVE_PC:        "DIRECTIVE_PC",
	TOKEN_DIRECTIVE_KICK_PRE:  "DIRECTIVE_KICK_PRE",
	TOKEN_DIRECTIVE_KICK_FLOW: "DIRECTIVE_KICK_FLOW",
	TOKEN_DIRECTIVE_KICK_ASM:  "DIRECTIVE_KICK_ASM",
	TOKEN_DIRECTIVE_KICK_DATA: "DIRECTIVE_KICK_DATA",
	TOKEN_DIRECTIVE_KICK_TEXT: "DIRECTIVE_KICK_TEXT",
	TOKEN_COLON:               "COLON",
	TOKEN_HASH:                "HASH",
	TOKEN_DOT:                 "DOT",
	TOKEN_COMMA:               "COMMA",
	TOKEN_PLUS:                "PLUS",
	TOKEN_MINUS:               "MINUS",
	TOKEN_ASTERISK:            "ASTERISK",
	TOKEN_SLASH:               "SLASH",
	TOKEN_LPAREN:              "LPAREN",
	TOKEN_RPAREN:              "RPAREN",
	TOKEN_LBRACKET:            "LBRACKET",
	TOKEN_RBRACKET:            "RBRACKET",
	TOKEN_LBRACE:              "LBRACE",
	TOKEN_RBRACE:              "RBRACE",
	TOKEN_EQUAL:               "EQUAL",
	TOKEN_LESS:                "LESS",
	TOKEN_GREATER:             "GREATER",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// tokenDefinition holds a token type and the regex used to match it.
type tokenDefinition struct {
	tokenType TokenType
	regex     *regexp.Regexp
}

// The order of these definitions is important for correct matching.
var tokenDefs = []tokenDefinition{
	{TOKEN_COMMENT, regexp.MustCompile(`^(//.*|;.*|/\*.*?\*/)`)}, // Corrected escaping for /* */
	{TOKEN_NUMBER_HEX, regexp.MustCompile(`^#?\$[0-9a-zA-Z]+`)},  // Corrected escaping for $
	{TOKEN_NUMBER_BIN, regexp.MustCompile(`^#?%[0-1]+`)},         // Corrected escaping for %
	{TOKEN_NUMBER_DEC, regexp.MustCompile(`^#?[0-9]+`)},
	{TOKEN_NUMBER_OCT, regexp.MustCompile(`^#?&[0-7]+`)}, // Corrected escaping for &
	{TOKEN_STRING, regexp.MustCompile(`^"(\\|[^\"])*"`)}, // Corrected escaping for " and "
	{TOKEN_MNEMONIC_STD, regexp.MustCompile(`^(?i)(adc|and|asl|bit|clc|cld|cli|clv|cmp|cpx|cpy|dec|dex|dey|eor|inc|inx|iny|lda|ldx|ldy|lsr|nop|ora|pha|php|pla|plp|rol|ror|sbc|sec|sed|sei|sta|stx|sty|ta x|txa|tay|tya|tsx|txs)\b`)},
	{TOKEN_MNEMONIC_CTRL, regexp.MustCompile(`^(?i)(bcc|bcs|beq|bmi|bne|bpl|brk|bvc|bvs|jmp|jsr|rti|rts)\b`)},
	{TOKEN_MNEMONIC_ILL, regexp.MustCompile(`^(?i)(slo|rla|sre|rra|sax|lax|dcp|isc|anc|asr|arr|sbx|dop|top|jam)\b`)},
	{TOKEN_MNEMONIC_65C02, regexp.MustCompile(`^(?i)((bbr|bbs|rmb|smb)[0-7]|trb|tsb|phx|phy|plx|ply|stz|bra)\b`)},
	{TOKEN_DIRECTIVE_KICK_PRE, regexp.MustCompile(`^#(define|elif|else|endif|if|import|importif|importonce|undef)\b`)},
	{TOKEN_DIRECTIVE_KICK_FLOW, regexp.MustCompile(`^\.(?i)(for|if|while|return)\b`)},
	{TOKEN_DIRECTIVE_KICK_ASM, regexp.MustCompile(`^\.(?i)(align|assert|asserterror|break|cpu|define|disk|encoding|error|errorif|eval|file|filemodify|filenamespace|function|import|importonce|label|lohifill|m acro|memblock|modify|namespace|pc|plugin|print|printnow|pseudocommand|pseudopc|segment|segmentdef|segmentout|zp)\b`)},
	{TOKEN_DIRECTIVE_KICK_DATA, regexp.MustCompile(`^\.(?i)(by|byte|const|dw|dword|enum|fill|fillword|struct|var|wo|word)\b`)},
	{TOKEN_DIRECTIVE_KICK_TEXT, regexp.MustCompile(`^\.(?i)(te|text)\b`)},
	{TOKEN_DIRECTIVE_PC, regexp.MustCompile(`^(\*=)`)},              // Corrected escaping for *=
	{TOKEN_LABEL, regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*):`)}, // Corrected escaping for :
	{TOKEN_IDENTIFIER, regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)},
	{TOKEN_COLON, regexp.MustCompile(`^:`)}, // Corrected escaping for :
	{TOKEN_HASH, regexp.MustCompile(`^#`)},
	{TOKEN_DOT, regexp.MustCompile(`^\.`)}, // Corrected escaping for .
	{TOKEN_COMMA, regexp.MustCompile(`^,`)},
	{TOKEN_PLUS, regexp.MustCompile(`^\+`)}, // Corrected escaping for +
	{TOKEN_MINUS, regexp.MustCompile(`^-`)},
	{TOKEN_ASTERISK, regexp.MustCompile(`^\*`)}, // Corrected escaping for *
	{TOKEN_SLASH, regexp.MustCompile(`^/`)},
	{TOKEN_LPAREN, regexp.MustCompile(`^\(`)},   // Corrected escaping for (
	{TOKEN_RPAREN, regexp.MustCompile(`^\)`)},   // Corrected escaping for )
	{TOKEN_LBRACKET, regexp.MustCompile(`^\[`)}, // Corrected escaping for [
	{TOKEN_RBRACKET, regexp.MustCompile(`^\]`)}, // Corrected escaping for ]
	{TOKEN_LBRACE, regexp.MustCompile(`^\{`)},   // Corrected escaping for {
	{TOKEN_RBRACE, regexp.MustCompile(`^\}`)},   // Corrected escaping for }
	{TOKEN_EQUAL, regexp.MustCompile(`^=`)},
	{TOKEN_LESS, regexp.MustCompile(`^<`)},
	{TOKEN_GREATER, regexp.MustCompile(`^>`)},
}

// Lexer holds the state of the lexical analysis.
type Lexer struct {
	input        string
	position     int // current position in input (points to current char)
	readPosition int // current reading position in input (after current char)
	line         int
	column       int
}

// NewLexer creates a new Lexer.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 1}
	return l
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.position >= len(l.input) {
		return Token{Type: TOKEN_EOF, Literal: "", Line: l.line, Column: l.column}
	}

	remainingInput := l.input[l.position:]

	for _, def := range tokenDefs {
		match := def.regex.FindString(remainingInput)
		if match != "" {
			tok := Token{
				Type:    def.tokenType,
				Literal: match,
				Line:    l.line,
				Column:  l.column,
			}
			l.advance(len(match))
			return tok
		}
	}

	// If no token is matched, we have an illegal character.
	log.Warn("Illegal character found at %d:%d", l.line, l.column)
	tok := Token{
		Type:    TOKEN_ILLEGAL,
		Literal: string(l.input[l.position]),
		Line:    l.line,
		Column:  l.column,
	}
	l.advance(1)
	return tok
}

func (l *Lexer) advance(n int) {
	for i := 0; i < n; i++ {
		if l.position < len(l.input) && l.input[l.position] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.position++
	}
	l.readPosition = l.position
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) && isWhitespace(l.input[l.position]) {
		l.advance(1)
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
