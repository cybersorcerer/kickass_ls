package lsp

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

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

	// Built-in Functions
	TOKEN_BUILTIN_MATH_FUNC
	TOKEN_BUILTIN_STRING_FUNC
	TOKEN_BUILTIN_FILE_FUNC
	TOKEN_BUILTIN_3D_FUNC

	// Built-in Constants
	TOKEN_BUILTIN_MATH_CONST
	TOKEN_BUILTIN_COLOR_CONST

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
	TOKEN_AT       // @ (program counter reference)
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
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// MnemonicInfo represents a mnemonic from mnemonic.json for lexer use
type MnemonicInfo struct {
	Mnemonic string `json:"mnemonic"`
	Type     string `json:"type"`
}

// DirectiveInfo represents a directive from kickass.json for lexer use
type DirectiveInfo struct {
	Directive string `json:"directive"`
}

// KickAssConfig represents the structure of the extended kickass.json
type KickAssConfig struct {
	Directives        []DirectiveInfo   `json:"directives"`
	BuiltinFunctions  []BuiltinFunction `json:"builtinFunctions"`
	BuiltinConstants  []BuiltinConstant `json:"builtinConstants"`
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

// loadMnemonicsFromJSON loads mnemonics from mnemonic.json and creates regex patterns
func loadMnemonicsFromJSON() map[TokenType]*regexp.Regexp {
	jsonPath := mnemonicJSONPath
	if jsonPath == "" {
		log.Error("FATAL: mnemonicJSONPath not set - must be initialized from $HOME/.config/6510lsp")
		os.Exit(1)
	}

	file, err := os.Open(jsonPath)
	if err != nil {
		log.Error("FATAL: Failed to open mnemonic.json at '%s': %v", jsonPath, err)
		log.Error("mnemonic.json is the Source of Truth and MUST be available at $HOME/.config/6510lsp")
		os.Exit(1)
	}
	defer file.Close()

	var mnemonics []MnemonicInfo
	if err := json.NewDecoder(file).Decode(&mnemonics); err != nil {
		log.Error("FATAL: Failed to parse mnemonic.json at '%s': %v", jsonPath, err)
		log.Error("mnemonic.json must contain valid JSON data")
		os.Exit(1)
	}

	// Group mnemonics by type
	stdOpcodes := []string{}
	ctrlOpcodes := []string{}
	illOpcodes := []string{}

	for _, mnemonic := range mnemonics {
		opcode := strings.ToLower(mnemonic.Mnemonic)
		switch mnemonic.Type {
		case "Transfer", "Arithmetic", "Logic", "Shift & Rotate", "Bit Test", "Flag", "Interrupt", "Comparison", "Decrement & Increment", "Other", "Stack":
			stdOpcodes = append(stdOpcodes, opcode)
		case "Jump":
			ctrlOpcodes = append(ctrlOpcodes, opcode)
		case "Illegal":
			illOpcodes = append(illOpcodes, opcode)
		}
	}

	// Create regex patterns
	regexes := make(map[TokenType]*regexp.Regexp)
	if len(stdOpcodes) > 0 {
		regexes[TOKEN_MNEMONIC_STD] = regexp.MustCompile(`^(?i)(` + strings.Join(stdOpcodes, "|") + `)\b`)
	}
	if len(ctrlOpcodes) > 0 {
		regexes[TOKEN_MNEMONIC_CTRL] = regexp.MustCompile(`^(?i)(` + strings.Join(ctrlOpcodes, "|") + `)\b`)
	}
	if len(illOpcodes) > 0 {
		regexes[TOKEN_MNEMONIC_ILL] = regexp.MustCompile(`^(?i)(` + strings.Join(illOpcodes, "|") + `)\b`)
	}

	log.Debug("Loaded mnemonics: %d std, %d ctrl, %d illegal", len(stdOpcodes), len(ctrlOpcodes), len(illOpcodes))
	return regexes
}

// createFallbackMnemonicRegexes provides empty fallback since JSON is now complete
func createFallbackMnemonicRegexes() map[TokenType]*regexp.Regexp {
	log.Error("JSON loading failed - mnemonic.json should be the only source of truth")
	return map[TokenType]*regexp.Regexp{}
}

// kickassJSONPath is set by the server to provide the correct path for kickass.json
var kickassJSONPath string

// mnemonicJSONPath is set by the server to provide the correct path for mnemonic.json
var mnemonicJSONPath string

// SetKickassJSONPath sets the path to kickass.json for lexer initialization
func SetKickassJSONPath(path string) {
	kickassJSONPath = path
	// Force re-initialization of token definitions when path changes
	tokenDefs = nil
}

// SetMnemonicJSONPath sets the path to mnemonic.json for lexer initialization
func SetMnemonicJSONPath(path string) {
	mnemonicJSONPath = path
	// Force re-initialization of token definitions when path changes
	tokenDefs = nil
}

// InitTokenDefs initializes token definitions after all JSON files are loaded
// MUST be called after SetMnemonicJSONPath and SetKickassJSONPath
func InitTokenDefs() {
	initTokenDefs()
}

// loadDirectivesFromJSON loads directives, functions and constants from kickass.json and creates regex patterns
func loadDirectivesFromJSON() map[TokenType]*regexp.Regexp {
	jsonPath := kickassJSONPath
	if jsonPath == "" {
		jsonPath = "kickass.json" // fallback
	}

	file, err := os.Open(jsonPath)
	if err != nil {
		log.Error("Failed to open kickass.json at %s: %v", jsonPath, err)
		return createFallbackDirectiveRegexes()
	}
	defer file.Close()

	var config KickAssConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Error("Failed to parse kickass.json: %v", err)
		return createFallbackDirectiveRegexes()
	}

	directives := config.Directives

	// Group directives by category
	preDirectives := []string{}    // #import, #importif etc
	flowDirectives := []string{}   // .if, .for, .while, .return
	asmDirectives := []string{}    // .align, .assert, .function, .macro etc
	dataDirectives := []string{}   // .byte, .const, .var etc
	textDirectives := []string{}   // .text, .te

	for _, directive := range directives {
		dir := strings.ToLower(directive.Directive)

		// Remove leading # or . for processing
		cleanDir := strings.TrimPrefix(strings.TrimPrefix(dir, "#"), ".")

		if strings.HasPrefix(directive.Directive, "#") {
			// Preprocessor directives
			preDirectives = append(preDirectives, cleanDir)
		} else {
			// Categorize by common patterns
			switch cleanDir {
			case "if", "for", "while", "return":
				flowDirectives = append(flowDirectives, cleanDir)
			case "by", "byte", "const", "dw", "dword", "enum", "fill", "fillword", "struct", "var", "wo", "word":
				dataDirectives = append(dataDirectives, cleanDir)
			case "te", "text":
				textDirectives = append(textDirectives, cleanDir)
			default:
				// Everything else goes to ASM
				asmDirectives = append(asmDirectives, cleanDir)
			}
		}
	}

	// Create regex patterns
	regexes := make(map[TokenType]*regexp.Regexp)
	if len(preDirectives) > 0 {
		regexes[TOKEN_DIRECTIVE_KICK_PRE] = regexp.MustCompile(`^#(` + strings.Join(preDirectives, "|") + `)\b`)
	}
	if len(flowDirectives) > 0 {
		regexes[TOKEN_DIRECTIVE_KICK_FLOW] = regexp.MustCompile(`^\.(?i)(` + strings.Join(flowDirectives, "|") + `)\b`)
	}
	if len(asmDirectives) > 0 {
		regexes[TOKEN_DIRECTIVE_KICK_ASM] = regexp.MustCompile(`^\.(?i)(` + strings.Join(asmDirectives, "|") + `)\b`)
	}
	if len(dataDirectives) > 0 {
		regexes[TOKEN_DIRECTIVE_KICK_DATA] = regexp.MustCompile(`^\.(?i)(` + strings.Join(dataDirectives, "|") + `)\b`)
	}
	if len(textDirectives) > 0 {
		regexes[TOKEN_DIRECTIVE_KICK_TEXT] = regexp.MustCompile(`^\.(?i)(` + strings.Join(textDirectives, "|") + `)\b`)
	}

	// Group built-in functions by category
	mathFunctions := []string{}
	stringFunctions := []string{}
	fileFunctions := []string{}
	d3Functions := []string{}

	for _, function := range config.BuiltinFunctions {
		funcName := strings.ToLower(function.Name)
		switch function.Category {
		case "math":
			mathFunctions = append(mathFunctions, funcName)
		case "string":
			stringFunctions = append(stringFunctions, funcName)
		case "file":
			fileFunctions = append(fileFunctions, funcName)
		case "3d":
			d3Functions = append(d3Functions, funcName)
		}
	}

	// Group built-in constants by category
	mathConstants := []string{}
	colorConstants := []string{}

	for _, constant := range config.BuiltinConstants {
		constName := strings.ToLower(constant.Name)
		switch constant.Category {
		case "math":
			mathConstants = append(mathConstants, constName)
		case "color":
			colorConstants = append(colorConstants, constName)
		}
	}

	// Add built-in function regexes
	if len(mathFunctions) > 0 {
		regexes[TOKEN_BUILTIN_MATH_FUNC] = regexp.MustCompile(`^(?i)(` + strings.Join(mathFunctions, "|") + `)\b`)
	}
	if len(stringFunctions) > 0 {
		regexes[TOKEN_BUILTIN_STRING_FUNC] = regexp.MustCompile(`^(?i)(` + strings.Join(stringFunctions, "|") + `)\b`)
	}
	if len(fileFunctions) > 0 {
		regexes[TOKEN_BUILTIN_FILE_FUNC] = regexp.MustCompile(`^(?i)(` + strings.Join(fileFunctions, "|") + `)\b`)
	}
	if len(d3Functions) > 0 {
		regexes[TOKEN_BUILTIN_3D_FUNC] = regexp.MustCompile(`^(?i)(` + strings.Join(d3Functions, "|") + `)\b`)
	}

	// Add built-in constant regexes
	if len(mathConstants) > 0 {
		regexes[TOKEN_BUILTIN_MATH_CONST] = regexp.MustCompile(`^(?i)(` + strings.Join(mathConstants, "|") + `)\b`)
	}
	if len(colorConstants) > 0 {
		regexes[TOKEN_BUILTIN_COLOR_CONST] = regexp.MustCompile(`^(?i)(` + strings.Join(colorConstants, "|") + `)\b`)
	}

	log.Debug("Loaded directives: %d pre, %d flow, %d asm, %d data, %d text", len(preDirectives), len(flowDirectives), len(asmDirectives), len(dataDirectives), len(textDirectives))
	log.Debug("Loaded functions: %d math, %d string, %d file, %d 3d", len(mathFunctions), len(stringFunctions), len(fileFunctions), len(d3Functions))
	log.Debug("Loaded constants: %d math, %d color", len(mathConstants), len(colorConstants))
	return regexes
}

// createFallbackDirectiveRegexes provides hardcoded regexes as fallback
func createFallbackDirectiveRegexes() map[TokenType]*regexp.Regexp {
	log.Warn("Using fallback hardcoded directive regexes")
	return map[TokenType]*regexp.Regexp{
		TOKEN_DIRECTIVE_KICK_PRE:  regexp.MustCompile(`^#(define|elif|else|endif|if|import|importif|importonce|undef)\b`),
		TOKEN_DIRECTIVE_KICK_FLOW: regexp.MustCompile(`^\.(?i)(for|if|while|return)\b`),
		TOKEN_DIRECTIVE_KICK_ASM:  regexp.MustCompile(`^\.(?i)(align|assert|asserterror|break|cpu|define|disk|encoding|error|errorif|eval|file|filemodify|filenamespace|function|import|importonce|label|lohifill|macro|memblock|modify|namespace|pc|plugin|print|printnow|pseudocommand|pseudopc|segment|segmentdef|segmentout|zp)\b`),
		TOKEN_DIRECTIVE_KICK_DATA: regexp.MustCompile(`^\.(?i)(by|byte|const|dw|dword|enum|fill|fillword|struct|var|wo|word)\b`),
		TOKEN_DIRECTIVE_KICK_TEXT: regexp.MustCompile(`^\.(?i)(te|text)\b`),
	}
}

// tokenDefinition holds a token type and the regex used to match it.
type tokenDefinition struct {
	tokenType TokenType
	regex     *regexp.Regexp
}

// The order of these definitions is important for correct matching.
var tokenDefs []tokenDefinition

// initTokenDefs initializes tokenDefs with mnemonic and directive regexes loaded from JSON
func initTokenDefs() {
	mnemonicRegexes := loadMnemonicsFromJSON()
	directiveRegexes := loadDirectivesFromJSON()

	tokenDefs = []tokenDefinition{
	{TOKEN_COMMENT, regexp.MustCompile(`^//.*`)},                   // Handle // comments
	{TOKEN_COMMENT, regexp.MustCompile(`^;.*`)},                    // Handle ; comments
	{TOKEN_COMMENT, regexp.MustCompile(`^/\*.*?\*/`)},               // Handle /* */ comments
	{TOKEN_NUMBER_HEX, regexp.MustCompile(`^#?\$[0-9a-zA-Z]+`)},  // Corrected escaping for $
	{TOKEN_NUMBER_BIN, regexp.MustCompile(`^#?%[0-1]+`)},         // Corrected escaping for %
	{TOKEN_NUMBER_DEC, regexp.MustCompile(`^#?[0-9]+(\.[0-9]+)?`)},
	{TOKEN_NUMBER_OCT, regexp.MustCompile(`^#?&[0-7]+`)}, // Corrected escaping for &
	{TOKEN_STRING, regexp.MustCompile(`^"(\\|[^\"])*"`)}, // Corrected escaping for " and "
	}

	// Add dynamic mnemonic regexes from JSON
	if regex, exists := mnemonicRegexes[TOKEN_MNEMONIC_STD]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_MNEMONIC_STD, regex})
	}
	if regex, exists := mnemonicRegexes[TOKEN_MNEMONIC_CTRL]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_MNEMONIC_CTRL, regex})
	}
	if regex, exists := mnemonicRegexes[TOKEN_MNEMONIC_ILL]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_MNEMONIC_ILL, regex})
	}

	// Add dynamic directive regexes from JSON
	if regex, exists := directiveRegexes[TOKEN_DIRECTIVE_KICK_PRE]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_DIRECTIVE_KICK_PRE, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_DIRECTIVE_KICK_FLOW]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_DIRECTIVE_KICK_FLOW, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_DIRECTIVE_KICK_ASM]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_DIRECTIVE_KICK_ASM, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_DIRECTIVE_KICK_DATA]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_DIRECTIVE_KICK_DATA, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_DIRECTIVE_KICK_TEXT]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_DIRECTIVE_KICK_TEXT, regex})
	}

	// Add dynamic built-in function regexes from JSON
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_MATH_FUNC]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_MATH_FUNC, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_STRING_FUNC]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_STRING_FUNC, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_FILE_FUNC]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_FILE_FUNC, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_3D_FUNC]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_3D_FUNC, regex})
	}

	// Add dynamic built-in constant regexes from JSON
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_MATH_CONST]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_MATH_CONST, regex})
	}
	if regex, exists := directiveRegexes[TOKEN_BUILTIN_COLOR_CONST]; exists {
		tokenDefs = append(tokenDefs, tokenDefinition{TOKEN_BUILTIN_COLOR_CONST, regex})
	}

	// Continue with other token definitions
	tokenDefs = append(tokenDefs, []tokenDefinition{
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
	{TOKEN_AT, regexp.MustCompile(`^@`)},
	}...)
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
	// Token definitions MUST be initialized by JSON loading before any lexer creation
	if len(tokenDefs) == 0 {
		log.Error("FATAL: tokenDefs not initialized - JSON files must be loaded first")
		log.Error("Call initTokenDefs() after loading all JSON files from $HOME/.config/6510lsp")
		os.Exit(1)
	}
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
