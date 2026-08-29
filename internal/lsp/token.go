package lsp

// TokenType represents the type of a token.
type TokenType int

// Token represents a single token parsed from the input.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
	// File is the document the token was read from. Line and Column are
	// relative to that file, so after several files have been spliced into one
	// translation unit this is what tells a diagnostic where it belongs.
	File string
}

const (
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF

	// Literals
	TOKEN_LABEL
	TOKEN_MULTILABEL      // Multi-label definition (!label:)
	TOKEN_MULTILABEL_FWD  // Multi-label forward reference (!label+)
	TOKEN_MULTILABEL_BACK // Multi-label backward reference (!label-)
	TOKEN_IDENTIFIER      // For identifiers that are not yet classified as labels, mnemonics, etc.

	// Comments
	TOKEN_COMMENT
	TOKEN_COMMENT_BLOCK

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

	// Flow control keywords
	TOKEN_ELSE // else keyword for .if directives

	// Bitwise operators
	TOKEN_LEFT_SHIFT  // <<
	TOKEN_RIGHT_SHIFT // >>
	TOKEN_BITWISE_AND // &
	TOKEN_BITWISE_OR  // |
	TOKEN_BITWISE_XOR // ^
	TOKEN_BITWISE_NOT // ~
	TOKEN_MODULO      // % (only produced when % is not a binary literal prefix)

	// Comparison operators (manual table 5.1)
	TOKEN_EQUAL_EQUAL   // ==
	TOKEN_NOT_EQUAL     // !=
	TOKEN_LESS_EQUAL    // <=
	TOKEN_GREATER_EQUAL // >=

	// Boolean operators (manual table 5.2)
	TOKEN_LOGICAL_NOT // !
	TOKEN_LOGICAL_AND // &&
	TOKEN_LOGICAL_OR  // ||

	// Conditional operator
	TOKEN_QUESTION // ? (paired with TOKEN_COLON for the ternary)

	// Built-in Functions
	TOKEN_BUILTIN_MATH_FUNC
	TOKEN_BUILTIN_STRING_FUNC
	TOKEN_BUILTIN_FILE_FUNC
	TOKEN_BUILTIN_3D_FUNC

	// Built-in Constants
	TOKEN_BUILTIN_MATH_CONST
	TOKEN_BUILTIN_COLOR_CONST

	// Punctuation
	TOKEN_COLON     // :
	TOKEN_HASH      // #
	TOKEN_DOT       // .
	TOKEN_COMMA     // ,
	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_ASTERISK  // *
	TOKEN_SLASH     // /
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
	TOKEN_EQUAL     // =
	TOKEN_LESS      // <
	TOKEN_GREATER   // >
	TOKEN_AT        // @ (program counter reference)
	TOKEN_SEMICOLON // ; (for .for loops)
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:             "ILLEGAL",
	TOKEN_EOF:                 "EOF",
	TOKEN_LABEL:               "LABEL",
	TOKEN_MULTILABEL:          "MULTILABEL",
	TOKEN_MULTILABEL_FWD:      "MULTILABEL_FWD",
	TOKEN_MULTILABEL_BACK:     "MULTILABEL_BACK",
	TOKEN_IDENTIFIER:          "IDENTIFIER",
	TOKEN_COMMENT:             "COMMENT",
	TOKEN_COMMENT_BLOCK:       "COMMENT_BLOCK",
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
	TOKEN_ELSE:                "ELSE",
	TOKEN_LEFT_SHIFT:          "LEFT_SHIFT",
	TOKEN_RIGHT_SHIFT:         "RIGHT_SHIFT",
	TOKEN_BITWISE_AND:         "BITWISE_AND",
	TOKEN_BITWISE_OR:          "BITWISE_OR",
	TOKEN_BITWISE_XOR:         "BITWISE_XOR",
	TOKEN_BITWISE_NOT:         "BITWISE_NOT",
	TOKEN_MODULO:              "MODULO",
	TOKEN_EQUAL_EQUAL:         "EQUAL_EQUAL",
	TOKEN_NOT_EQUAL:           "NOT_EQUAL",
	TOKEN_LESS_EQUAL:          "LESS_EQUAL",
	TOKEN_GREATER_EQUAL:       "GREATER_EQUAL",
	TOKEN_LOGICAL_NOT:         "LOGICAL_NOT",
	TOKEN_LOGICAL_AND:         "LOGICAL_AND",
	TOKEN_LOGICAL_OR:          "LOGICAL_OR",
	TOKEN_QUESTION:            "QUESTION",
	TOKEN_BUILTIN_MATH_FUNC:   "BUILTIN_MATH_FUNC",
	TOKEN_BUILTIN_STRING_FUNC: "BUILTIN_STRING_FUNC",
	TOKEN_BUILTIN_FILE_FUNC:   "BUILTIN_FILE_FUNC",
	TOKEN_BUILTIN_3D_FUNC:     "BUILTIN_3D_FUNC",
	TOKEN_BUILTIN_MATH_CONST:  "BUILTIN_MATH_CONST",
	TOKEN_BUILTIN_COLOR_CONST: "BUILTIN_COLOR_CONST",
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
	TOKEN_AT:                  "AT",
	TOKEN_SEMICOLON:           "SEMICOLON",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// BuiltinFunction represents a built-in function from kickass.json
type BuiltinFunction struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Signature   string   `json:"signature"`
	Examples    []string `json:"examples"`
}

// BuiltinConstant represents a built-in constant from kickass.json
type BuiltinConstant struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Value       string   `json:"value"`
	Examples    []string `json:"examples"`
}
