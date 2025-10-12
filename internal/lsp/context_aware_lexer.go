package lsp

import (
	"fmt"
	"strings"
	"sync"

	log "c64.nvim/internal/log"
)

// Context-Aware Lexer and Parser Structures for 6510/C64/Kick Assembler

// LexerState represents the current state of the context-aware lexer
type LexerState int

const (
	StateNormal LexerState = iota
	StateDirective
	StateExpression
	StateStringLiteral
	StateForLoop
	StateBlock
	StateConditional
	StateInstruction      // NEW: Inside 6510 instruction
	StateAddressingMode   // NEW: Parsing addressing mode
	StateOperand         // NEW: Parsing instruction operand
	StateKickFunction    // NEW: Inside Kick Assembler function
	StateKickConstant    // NEW: Kick Assembler constant context
)

// String returns the string representation of the lexer state
func (s LexerState) String() string {
	switch s {
	case StateNormal:
		return "Normal"
	case StateDirective:
		return "Directive"
	case StateExpression:
		return "Expression"
	case StateStringLiteral:
		return "StringLiteral"
	case StateForLoop:
		return "ForLoop"
	case StateBlock:
		return "Block"
	case StateConditional:
		return "Conditional"
	case StateInstruction:
		return "Instruction"
	case StateAddressingMode:
		return "AddressingMode"
	case StateOperand:
		return "Operand"
	case StateKickFunction:
		return "KickFunction"
	case StateKickConstant:
		return "KickConstant"
	default:
		return "Unknown"
	}
}

// ProcessorContext holds all 6510/C64/Kick Assembler specific context
type ProcessorContext struct {
	// Mnemonics from mnemonic.json
	StandardMnemonics map[string]*EnhancedMnemonicInfo `json:"standard_mnemonics"`
	IllegalMnemonics  map[string]*EnhancedMnemonicInfo `json:"illegal_mnemonics"`
	ControlMnemonics  map[string]*EnhancedMnemonicInfo `json:"control_mnemonics"`

	// Kick Assembler directives from kickass.json
	Directives             map[string]*KickDirectiveInfo `json:"directives"`
	PreprocessorStatements map[string]*KickDirectiveInfo `json:"preprocessor_statements"`
	Functions              map[string]*FunctionInfo      `json:"functions"`
	Constants              map[string]*ConstantInfo      `json:"constants"`

	// C64 memory regions from c64memory.json
	MemoryRegions    []*MemoryRegion `json:"memory_regions"`
	MemoryMap        map[uint16]*MemoryRegion `json:"memory_map"` // For fast address lookups

	// Cached lookups for performance
	AllMnemonics     map[string]*EnhancedMnemonicInfo `json:"-"` // Combined mnemonics cache
	DirectiveNames   []string                         `json:"-"` // For completion
	FunctionNames    []string                         `json:"-"` // For completion
	ConstantNames    []string                         `json:"-"` // For completion

	mutex sync.RWMutex
}

// EnhancedMnemonicInfo represents a 6510 mnemonic with all addressing modes for context-aware parser
type EnhancedMnemonicInfo struct {
	Name            string                 `json:"mnemonic"`
	Description     string                 `json:"description"`
	Type            string                 `json:"type"`
	AddressingModes []*AddressingModeInfo  `json:"addressing_modes"`
}

// AddressingModeInfo represents a specific addressing mode for a mnemonic
type AddressingModeInfo struct {
	Opcode          string `json:"opcode"`
	Mode            string `json:"addressing_mode"`  // "Immediate", "Absolute", "Zero Page", etc.
	AssemblerFormat string `json:"assembler_format"` // "LDA #nn", "LDA nnnn", etc.
	Length          int    `json:"length"`
	Cycles          string `json:"cycles"`
}

// StatementSourceType indicates which table in kickass.json the statement comes from
type StatementSourceType int

const (
	SourceUnknown StatementSourceType = iota
	SourceDirective
	SourcePreprocessor
	SourceFunction
	SourceConstant
)

// KickDirectiveInfo represents a Kick Assembler directive for context-aware parser
type KickDirectiveInfo struct {
	Name        string              `json:"directive"`
	Description string              `json:"description"`
	Signature   string              `json:"signature"`
	Examples    []string            `json:"examples"`
	Category    string              `json:"category,omitempty"` // "data", "flow", "pre", etc.
	SourceType  StatementSourceType `json:"-"`                  // Which table this came from
}

// FunctionInfo represents a Kick Assembler built-in function
type FunctionInfo struct {
	Name        string   `json:"function"`
	Description string   `json:"description"`
	Signature   string   `json:"signature"`
	Examples    []string `json:"examples"`
	ReturnType  string   `json:"return_type,omitempty"`
	Category    string   `json:"category,omitempty"` // "math", "string", "file", etc.
}

// ConstantInfo represents a Kick Assembler built-in constant
type ConstantInfo struct {
	Name        string `json:"constant"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category,omitempty"` // "math", "color", etc.
}

// MemoryRegion represents a C64 memory region
type MemoryRegion struct {
	Address     uint16            `json:"address"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`    // "System", "Graphics", "Sound", etc.
	Type        string            `json:"type"`        // "register", "ram", "rom", etc.
	Size        int               `json:"size"`
	Description string            `json:"description"`
	Access      string            `json:"access"`      // "read", "write", "read/write"
	BitFields   map[string]string `json:"bit_fields,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
	Tips        []string          `json:"tips,omitempty"`
}

// LexerContext represents the context stack for the lexer
type LexerContext struct {
	State      LexerState                 `json:"state"`
	Directive  string                     `json:"directive,omitempty"`  // Current directive (.byte, .for, etc.)
	Depth      int                        `json:"depth"`                // Nesting depth
	Parameters map[string]interface{}     `json:"parameters,omitempty"` // Context-specific parameters
}

// ContextAwareLexer represents the new context-aware lexer
type ContextAwareLexer struct {
	input           string
	position        int
	line            int
	column          int
	contextStack    []LexerContext
	processorCtx    *ProcessorContext
	debugMode       bool

	// Token buffer for lookahead
	tokenBuffer     []*ContextToken
	bufferPos       int

	// Parenthesis depth tracking for ; handling in .for loops
	parenDepth      int

	mutex           sync.RWMutex
}

// ContextToken represents a token with enhanced context information
type ContextToken struct {
	Type        TokenType       `json:"type"`
	Literal     string          `json:"literal"`
	Line        int             `json:"line"`
	Column      int             `json:"column"`
	Context     LexerContext    `json:"context"`     // NEW: Context information
	Metadata    *TokenMetadata  `json:"metadata"`    // NEW: Additional semantic info
}

// TokenMetadata contains additional semantic information about tokens
type TokenMetadata struct {
	// Directive context
	IsPartOfDirective   bool   `json:"is_part_of_directive,omitempty"`
	DirectiveName       string `json:"directive_name,omitempty"`
	ParameterIndex      int    `json:"parameter_index,omitempty"`
	ExpressionDepth     int    `json:"expression_depth,omitempty"`

	// 6510 Instruction context
	IsInstruction       bool                     `json:"is_instruction,omitempty"`
	MnemonicInfo        *EnhancedMnemonicInfo    `json:"mnemonic_info,omitempty"`
	AddressingMode      string                   `json:"addressing_mode,omitempty"`
	IsOperand          bool                      `json:"is_operand,omitempty"`
	OperandType        string                    `json:"operand_type,omitempty"` // "immediate", "absolute", "zeropage", etc.

	// Memory context
	MemoryRegion       *MemoryRegion        `json:"memory_region,omitempty"`
	IsMemoryAddress    bool                 `json:"is_memory_address,omitempty"`
	AddressType        string               `json:"address_type,omitempty"` // "zeropage", "absolute", "relative"

	// Kick Assembler context
	IsKickFunction     bool                 `json:"is_kick_function,omitempty"`
	FunctionInfo       *FunctionInfo        `json:"function_info,omitempty"`
	IsKickConstant     bool                 `json:"is_kick_constant,omitempty"`
	ConstantInfo       *ConstantInfo        `json:"constant_info,omitempty"`

	// Validation hints
	ValidationHints    []string             `json:"validation_hints,omitempty"`
	Suggestions        []string             `json:"suggestions,omitempty"`
}

// Global processor context instance
var globalProcessorContext *ProcessorContext
var processorContextMutex sync.RWMutex

// GetProcessorContext returns the global processor context
func GetProcessorContext() *ProcessorContext {
	processorContextMutex.RLock()
	defer processorContextMutex.RUnlock()
	return globalProcessorContext
}

// LoadProcessorContext loads all JSON configuration into the processor context
func LoadProcessorContext(mnemonicPath, kickassPath, c64MemoryPath string) error {
	processorContextMutex.Lock()
	defer processorContextMutex.Unlock()

	ctx := &ProcessorContext{
		StandardMnemonics:      make(map[string]*EnhancedMnemonicInfo),
		IllegalMnemonics:       make(map[string]*EnhancedMnemonicInfo),
		ControlMnemonics:       make(map[string]*EnhancedMnemonicInfo),
		Directives:             make(map[string]*KickDirectiveInfo),
		PreprocessorStatements: make(map[string]*KickDirectiveInfo),
		Functions:              make(map[string]*FunctionInfo),
		Constants:              make(map[string]*ConstantInfo),
		MemoryMap:              make(map[uint16]*MemoryRegion),
		AllMnemonics:           make(map[string]*EnhancedMnemonicInfo),
	}

	// Load mnemonics
	if err := ctx.loadMnemonics(mnemonicPath); err != nil {
		return err
	}

	// Load Kick Assembler data
	if err := ctx.loadKickAssemblerData(kickassPath); err != nil {
		return err
	}

	// Load C64 memory map
	if err := ctx.loadC64Memory(c64MemoryPath); err != nil {
		return err
	}

	// Build caches
	ctx.buildCaches()

	globalProcessorContext = ctx
	return nil
}

// Helper methods for ProcessorContext are implemented in context_aware_loader.go

// NewContextAwareLexer creates a new context-aware lexer instance
func NewContextAwareLexer(input string, processorCtx *ProcessorContext) *ContextAwareLexer {
	return &ContextAwareLexer{
		input:        input,
		position:     0,
		line:         1,
		column:       1,
		contextStack: []LexerContext{{State: StateNormal, Depth: 0}},
		processorCtx: processorCtx,
		debugMode:    IsParserDebugModeEnabled(),
		tokenBuffer:  make([]*ContextToken, 0, 10), // Lookahead buffer
	}
}

// SetDebugMode enables or disables debug mode
func (l *ContextAwareLexer) SetDebugMode(enabled bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.debugMode = enabled
}

// Context management methods
func (l *ContextAwareLexer) PushContext(state LexerState, directive string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	newContext := LexerContext{
		State:     state,
		Directive: directive,
		Depth:     len(l.contextStack),
		Parameters: make(map[string]interface{}),
	}

	l.contextStack = append(l.contextStack, newContext)

	if l.debugMode {
		log.Debug("Pushed lexer context: %s (directive: %s, depth: %d)",
			state.String(), directive, newContext.Depth)
	}
}

func (l *ContextAwareLexer) PopContext() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.contextStack) > 1 {
		popped := l.contextStack[len(l.contextStack)-1]
		l.contextStack = l.contextStack[:len(l.contextStack)-1]

		if l.debugMode {
			log.Debug("Popped lexer context: %s (depth: %d)",
				popped.State.String(), popped.Depth)
		}
	}
}

func (l *ContextAwareLexer) CurrentContext() LexerContext {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if len(l.contextStack) == 0 {
		return LexerContext{State: StateNormal, Depth: 0}
	}
	return l.contextStack[len(l.contextStack)-1]
}

// Main tokenization method - context-aware tokenization
func (l *ContextAwareLexer) NextToken() *ContextToken {
	l.skipWhitespace()

	// Check for EOF
	if l.position >= len(l.input) {
		return l.createToken(TOKEN_EOF, "", l.column, nil)
	}

	// Get current context
	ctx := l.CurrentContext()

	if l.debugMode {
		log.Debug("NextToken at Line %d, Col %d, State: %s, Stack depth: %d",
			l.line, l.column, ctx.State.String(), len(l.contextStack))
	}

	remaining := l.input[l.position:]

	// Handle comments (always highest priority)
	if strings.HasPrefix(remaining, "//") {
		return l.readLineComment()
	}
	if strings.HasPrefix(remaining, "/*") {
		return l.readBlockComment()
	}
	// Semicolon comment (assembly style)
	// Important: ; is a comment EXCEPT inside parentheses (for .for loops)
	// .for loops use ; as separator: .for (init; test; increment)
	// But inline comments work: lda #$00  ; load zero
	if strings.HasPrefix(remaining, ";") {
		// Only treat as separator if we're inside parentheses
		if l.parenDepth == 0 {
			return l.readLineComment() // Semicolon comment
		}
		// Otherwise, let it be tokenized as TOKEN_SEMICOLON for .for loops
	}

	// Context-aware tokenization based on current state
	switch ctx.State {
	case StateStringLiteral:
		return l.tokenizeString()

	case StateDirective:
		return l.tokenizeDirectiveContent()

	case StateExpression:
		return l.tokenizeExpression()

	case StateInstruction:
		return l.tokenizeInstruction()

	case StateAddressingMode:
		return l.tokenizeAddressingMode()

	case StateOperand:
		return l.tokenizeOperand()

	default: // StateNormal, StateForLoop, StateBlock, StateConditional
		return l.tokenizeNormal()
	}
}

// skipWhitespace skips over whitespace characters
// In StateOperand/StateInstruction/StateDirective, we don't skip newlines because they signal end
func (l *ContextAwareLexer) skipWhitespace() {
	ctx := l.CurrentContext()

	for l.position < len(l.input) {
		ch := l.input[l.position]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else if ch == '\n' {
			// In operand, instruction, or directive state, newline is significant - don't skip it
			if ctx.State == StateOperand || ctx.State == StateInstruction || ctx.State == StateDirective {
				if l.debugMode {
					log.Debug("skipWhitespace: NOT skipping newline in state %s at Line %d, Col %d", ctx.State.String(), l.line, l.column)
				}
				break
			}
			if l.debugMode {
				log.Debug("skipWhitespace: Skipping newline in state %s at Line %d, Col %d, resetting column to 1", ctx.State.String(), l.line, l.column)
			}
			l.advance()
			l.line++
			l.column = 1
		} else {
			break
		}
	}
}

// advance moves the lexer position forward
func (l *ContextAwareLexer) advance() {
	if l.position < len(l.input) {
		l.position++
		l.column++
	}
}

// peek returns the character at current position without advancing
func (l *ContextAwareLexer) peek() byte {
	if l.position < len(l.input) {
		return l.input[l.position]
	}
	return 0
}

// peekAhead returns the character n positions ahead
func (l *ContextAwareLexer) peekAhead(n int) byte {
	pos := l.position + n
	if pos < len(l.input) {
		return l.input[pos]
	}
	return 0
}

// readLineComment reads a line comment starting with //
func (l *ContextAwareLexer) readLineComment() *ContextToken {
	startLine := l.line
	startCol := l.column
	start := l.position

	// Skip //
	l.advance()
	l.advance()

	// Read until end of line
	for l.position < len(l.input) && l.peek() != '\n' {
		l.advance()
	}

	literal := l.input[start:l.position]
	token := &ContextToken{
		Type:    TOKEN_COMMENT,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
		Context: l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}

	return token
}

// readBlockComment reads a block comment starting with /* and ending with */
func (l *ContextAwareLexer) readBlockComment() *ContextToken {
	startLine := l.line
	startCol := l.column
	start := l.position

	// Skip /*
	l.advance()
	l.advance()

	// Read until */ is found
	for l.position+1 < len(l.input) {
		if l.peek() == '*' && l.peekAhead(1) == '/' {
			l.advance() // skip *
			l.advance() // skip /
			break
		}
		// Track line numbers inside block comments
		if l.peek() == '\n' {
			l.advance()
			l.line++
			l.column = 1
		} else {
			l.advance()
		}
	}

	literal := l.input[start:l.position]
	token := &ContextToken{
		Type:    TOKEN_COMMENT_BLOCK,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
		Context: l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}

	return token
}

// createToken creates a context token with the current context
func (l *ContextAwareLexer) createToken(tokenType TokenType, literal string, startCol int, metadata *TokenMetadata) *ContextToken {
	if metadata == nil {
		metadata = &TokenMetadata{}
	}

	return &ContextToken{
		Type:     tokenType,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,  // Use start column, not current column
		Context:  l.CurrentContext(),
		Metadata: metadata,
	}
}

// tokenizeNormal handles tokenization in normal state
func (l *ContextAwareLexer) tokenizeNormal() *ContextToken {
	remaining := l.input[l.position:]

	// Check for preprocessor directives (#define, #undef, #import, #importif)
	if strings.HasPrefix(remaining, "#") {
		// Check if it matches any preprocessor statement from kickass.json
		if l.processorCtx != nil && l.processorCtx.PreprocessorStatements != nil {
			for directiveName := range l.processorCtx.PreprocessorStatements {
				if strings.HasPrefix(remaining, directiveName) {
					return l.tokenizePreprocessorDirective()
				}
			}
		}
	}

	// Check for directives first (start with .)
	if strings.HasPrefix(remaining, ".") {
		return l.tokenizeDirective()
	}

	// Check for program counter directive (*=)
	if strings.HasPrefix(remaining, "*=") {
		return l.tokenizeProgramCounter()
	}

	// Check for labels (identifier followed by :)
	if token := l.tryTokenizeLabel(); token != nil {
		return token
	}

	// Check for mnemonics (6510 instructions)
	if token := l.tryTokenizeMnemonic(); token != nil {
		// Enter instruction state
		l.PushContext(StateInstruction, "")
		return token
	}

	// Check for numbers
	if token := l.tryTokenizeNumber(); token != nil {
		return token
	}

	// Check for strings
	if l.peek() == '"' {
		l.PushContext(StateStringLiteral, "")
		return l.tokenizeString()
	}

	// Check for identifiers (variables, labels references, etc.)
	if token := l.tryTokenizeIdentifier(); token != nil {
		return token
	}

	// Check for operators and punctuation
	return l.tokenizeOperatorOrPunctuation()
}

// tokenizeDirective tokenizes a Kick Assembler directive
func (l *ContextAwareLexer) tokenizeDirective() *ContextToken {
	start := l.position
	startCol := l.column

	// Read the directive name
	l.advance() // skip .
	for l.position < len(l.input) {
		ch := l.peek()
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			l.advance()
		} else {
			break
		}
	}

	literal := l.input[start:l.position]
	directiveName := strings.ToLower(literal)

	// Look up directive info from processor context
	var tokenType TokenType = TOKEN_DIRECTIVE_KICK_PRE // default

	if l.processorCtx != nil {
		directiveInfo := l.processorCtx.GetDirectiveInfo(directiveName)
		if directiveInfo != nil {
			// Categorize by directive category
			switch directiveInfo.Category {
			case "flow":
				tokenType = TOKEN_DIRECTIVE_KICK_FLOW
			case "data":
				tokenType = TOKEN_DIRECTIVE_KICK_DATA
			case "asm":
				tokenType = TOKEN_DIRECTIVE_KICK_ASM
			case "text":
				tokenType = TOKEN_DIRECTIVE_KICK_TEXT
			default:
				tokenType = TOKEN_DIRECTIVE_KICK_PRE
			}
		}
	}

	// Push directive context
	l.PushContext(StateDirective, directiveName)

	metadata := &TokenMetadata{
		IsPartOfDirective: true,
		DirectiveName:     directiveName,
	}

	return &ContextToken{
		Type:     tokenType,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: metadata,
	}
}

// tokenizePreprocessorDirective tokenizes preprocessor directives (#define, #undef, #import, #importif)
func (l *ContextAwareLexer) tokenizePreprocessorDirective() *ContextToken {
	start := l.position
	startCol := l.column

	// Read the directive name (starts with #)
	l.advance() // skip #
	for l.position < len(l.input) {
		ch := l.peek()
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			l.advance()
		} else {
			break
		}
	}

	literal := l.input[start:l.position]
	directiveName := strings.ToLower(literal)

	// Push directive context
	l.PushContext(StateDirective, directiveName)

	metadata := &TokenMetadata{
		IsPartOfDirective: true,
		DirectiveName:     directiveName,
	}

	return &ContextToken{
		Type:     TOKEN_DIRECTIVE_KICK_PRE, // Preprocessor directive
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: metadata,
	}
}

// tokenizeDirectiveContent handles tokens inside a directive
func (l *ContextAwareLexer) tokenizeDirectiveContent() *ContextToken {
	// Check for end of directive (newline or special constructs)
	if l.peek() == '\n' {
		l.PopContext() // Exit directive context
		l.advance()
		l.line++
		l.column = 1
		return l.NextToken()
	}

	// Check for block start (for .for, .if, etc.)
	if l.peek() == '{' {
		l.PopContext() // Exit directive context
		l.PushContext(StateBlock, "")
		return l.tokenizeOperatorOrPunctuation()
	}

	// Check for expressions in parentheses
	if l.peek() == '(' {
		l.PushContext(StateExpression, "")
		return l.tokenizeOperatorOrPunctuation()
	}

	// Check for numbers
	if token := l.tryTokenizeNumber(); token != nil {
		return token
	}

	// Check for strings
	if l.peek() == '"' {
		l.PushContext(StateStringLiteral, "")
		return l.tokenizeString()
	}

	// Check for identifiers
	if token := l.tryTokenizeIdentifier(); token != nil {
		return token
	}

	// Operators and punctuation
	return l.tokenizeOperatorOrPunctuation()
}

// tokenizeExpression handles tokens inside expressions
func (l *ContextAwareLexer) tokenizeExpression() *ContextToken {
	// Check for expression end
	if l.peek() == ')' {
		l.PopContext() // Exit expression context
		return l.tokenizeOperatorOrPunctuation()
	}

	// Check for nested expressions
	if l.peek() == '(' {
		l.PushContext(StateExpression, "")
		return l.tokenizeOperatorOrPunctuation()
	}

	// Check for numbers
	if token := l.tryTokenizeNumber(); token != nil {
		return token
	}

	// Check for strings
	if l.peek() == '"' {
		l.PushContext(StateStringLiteral, "")
		return l.tokenizeString()
	}

	// Check for identifiers/functions/constants
	if token := l.tryTokenizeIdentifier(); token != nil {
		return token
	}

	// Operators and punctuation
	return l.tokenizeOperatorOrPunctuation()
}

// tokenizeInstruction handles tokenization after a mnemonic
func (l *ContextAwareLexer) tokenizeInstruction() *ContextToken {
	// After mnemonic, we expect operand or end of line
	if l.peek() == '\n' {
		l.PopContext() // Exit instruction context
		l.advance()
		l.line++
		l.column = 1
		return l.NextToken()
	}

	// Enter operand parsing
	l.PopContext() // Exit instruction context
	l.PushContext(StateOperand, "")
	return l.tokenizeOperand()
}

// tokenizeOperand handles instruction operands with addressing mode detection
func (l *ContextAwareLexer) tokenizeOperand() *ContextToken {
	// Check for end of operand (newline or comment)
	if l.peek() == '\n' || (l.peek() == '/' && l.peekAhead(1) == '/') {
		l.PopContext() // Exit operand context, return to normal
		if l.peek() == '\n' {
			l.advance()
			l.line++
			l.column = 1
		}
		return l.NextToken()
	}

	// Detect addressing mode
	ch := l.peek()

	// Immediate addressing (#)
	if ch == '#' {
		token := l.tokenizeOperatorOrPunctuation()
		token.Metadata.IsOperand = true
		token.Metadata.OperandType = "immediate"
		return token
	}

	// Indirect addressing (parentheses)
	if ch == '(' {
		l.PushContext(StateAddressingMode, "")
		token := l.tokenizeOperatorOrPunctuation()
		token.Metadata.IsOperand = true
		token.Metadata.OperandType = "indirect"
		return token
	}

	// Comma - check if followed by X or Y for indexed addressing
	if ch == ',' {
		token := l.tokenizeOperatorOrPunctuation()
		// Peek ahead to see if X or Y follows
		l.skipWhitespace()
		nextCh := l.peek()
		if nextCh == 'X' || nextCh == 'x' || nextCh == 'Y' || nextCh == 'y' {
			// Mark this as indexed addressing comma
			token.Metadata.IsOperand = true
			token.Metadata.OperandType = "indexed_separator"
		}
		return token
	}

	// Numbers or identifiers (absolute, zeropage, or labels)
	if token := l.tryTokenizeNumber(); token != nil {
		token.Metadata.IsOperand = true
		token.Metadata.OperandType = "address"
		return token
	}

	// Check for X or Y register (for indexed addressing)
	if ch == 'X' || ch == 'x' || ch == 'Y' || ch == 'y' {
		startCol := l.column
		register := l.peek()
		l.advance()

		// Make sure it's just a single letter (not part of a longer identifier)
		if l.position >= len(l.input) || !isAlphaNumeric(l.peek()) {
			token := &ContextToken{
				Type:     TOKEN_IDENTIFIER, // Could create TOKEN_INDEX_REGISTER
				Literal:  string(register),
				Line:     l.line,
				Column:   startCol,
				Context:  l.CurrentContext(),
				Metadata: &TokenMetadata{
					IsOperand:   true,
					OperandType: "index_register",
				},
			}
			// After index register, exit operand state
			l.PopContext()
			return token
		}
	}

	if token := l.tryTokenizeIdentifier(); token != nil {
		token.Metadata.IsOperand = true
		token.Metadata.OperandType = "label"
		return token
	}

	// Operators (for indexed addressing ,X ,Y)
	return l.tokenizeOperatorOrPunctuation()
}

// tokenizeAddressingMode handles tokens inside indirect addressing
func (l *ContextAwareLexer) tokenizeAddressingMode() *ContextToken {
	// Check for end of indirect addressing
	if l.peek() == ')' {
		l.PopContext() // Exit addressing mode context
		return l.tokenizeOperatorOrPunctuation()
	}

	// Numbers or identifiers
	if token := l.tryTokenizeNumber(); token != nil {
		return token
	}

	if token := l.tryTokenizeIdentifier(); token != nil {
		return token
	}

	return l.tokenizeOperatorOrPunctuation()
}

// tokenizeString handles string literals
func (l *ContextAwareLexer) tokenizeString() *ContextToken {
	start := l.position
	startCol := l.column

	l.advance() // skip opening "

	// Read until closing " or end of line
	for l.position < len(l.input) {
		ch := l.peek()
		if ch == '"' {
			l.advance() // include closing "
			break
		} else if ch == '\\' {
			l.advance() // skip escape char
			if l.position < len(l.input) {
				l.advance() // skip escaped char
			}
		} else if ch == '\n' {
			// Unterminated string
			break
		} else {
			l.advance()
		}
	}

	literal := l.input[start:l.position]

	// Pop string literal context
	l.PopContext()

	return &ContextToken{
		Type:     TOKEN_STRING,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}
}

// tokenizeProgramCounter handles *= directive
func (l *ContextAwareLexer) tokenizeProgramCounter() *ContextToken {
	startCol := l.column
	l.advance() // *
	l.advance() // =

	return &ContextToken{
		Type:     TOKEN_DIRECTIVE_PC,
		Literal:  "*=",
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}
}

// tryTokenizeLabel attempts to tokenize a label (identifier:)
func (l *ContextAwareLexer) tryTokenizeLabel() *ContextToken {
	// Look ahead to see if this is a label
	saved := l.position
	savedCol := l.column

	if l.debugMode && l.line >= 80 && l.line <= 110 {
		log.Debug("tryTokenizeLabel: ENTER Line %d, l.position=%d, l.column=%d", l.line, l.position, l.column)
	}

	// Read identifier
	if !isAlpha(l.peek()) && l.peek() != '_' {
		return nil
	}

	tempPos := l.position
	for tempPos < len(l.input) {
		ch := l.input[tempPos]
		if isAlphaNumeric(ch) || ch == '_' {
			tempPos++
		} else {
			break
		}
	}

	// Check if followed by :
	if tempPos < len(l.input) && l.input[tempPos] == ':' {
		// It's a label
		startCol := l.column
		literal := l.input[l.position : tempPos+1]

		// Advance position
		for l.position <= tempPos {
			l.advance()
		}

		return &ContextToken{
			Type:     TOKEN_LABEL,
			Literal:  literal,
			Line:     l.line,
			Column:   startCol,
			Context:  l.CurrentContext(),
			Metadata: &TokenMetadata{},
		}
	}

	// Not a label, restore position
	l.position = saved

	if l.debugMode && l.line >= 80 && l.line <= 110 {
		log.Debug("tryTokenizeLabel: EXIT (not a label) Line %d, l.position=%d, l.column=%d, savedCol=%d", l.line, l.position, l.column, savedCol)
	}

	return nil
}

// tryTokenizeMnemonic attempts to tokenize a 6510 mnemonic
func (l *ContextAwareLexer) tryTokenizeMnemonic() *ContextToken {
	if !isAlpha(l.peek()) {
		return nil
	}

	start := l.position
	startCol := l.column

	// Read 3 characters (mnemonics are always 3 chars)
	var mnemonic [3]byte
	for i := 0; i < 3 && l.position < len(l.input); i++ {
		ch := l.peek()
		if isAlpha(ch) {
			mnemonic[i] = ch
			l.advance()
		} else {
			// Not a mnemonic - restore position AND column
			l.position = start
			l.column = startCol
			return nil
		}
	}

	// Check if it's a word boundary (space, newline, comment, etc.)
	if l.position < len(l.input) {
		ch := l.peek()
		if isAlphaNumeric(ch) || ch == '_' {
			// Part of longer identifier, not a mnemonic - restore position AND column
			l.position = start
			l.column = startCol
			return nil
		}
	}

	mnemonicStr := strings.ToUpper(string(mnemonic[:]))

	// Look up in processor context
	if l.processorCtx != nil {
		mnemonicInfo := l.processorCtx.GetMnemonicInfo(mnemonicStr)
		if mnemonicInfo != nil {
			tokenType := TOKEN_MNEMONIC_STD
			isIllegal := l.processorCtx.IsIllegalMnemonic(mnemonicStr)

			if isIllegal {
				tokenType = TOKEN_MNEMONIC_ILL
			} else if mnemonicInfo.Type == "Jump" {
				tokenType = TOKEN_MNEMONIC_CTRL
			}

			metadata := &TokenMetadata{
				IsInstruction: true,
				MnemonicInfo:  mnemonicInfo,
			}

			return &ContextToken{
				Type:     tokenType,
				Literal:  l.input[start:l.position],
				Line:     l.line,
				Column:   startCol,
				Context:  l.CurrentContext(),
				Metadata: metadata,
			}
		}
	}

	// Not a valid mnemonic, restore position AND column
	l.position = start
	l.column = startCol
	return nil
}

// tryTokenizeNumber attempts to tokenize a number
func (l *ContextAwareLexer) tryTokenizeNumber() *ContextToken {
	start := l.position
	startCol := l.column
	ch := l.peek()

	// Optional # prefix for immediate values
	if ch == '#' {
		l.advance()
		ch = l.peek()
	}

	var tokenType TokenType
	var literal string

	// Hex number ($xxxx)
	if ch == '$' {
		l.advance()
		hasValidHexDigits := false
		for l.position < len(l.input) && isHexDigit(l.peek()) {
			l.advance()
			hasValidHexDigits = true
		}

		// Check if followed by invalid characters (like G-Z)
		nextCh := l.peek()
		if (nextCh >= 'G' && nextCh <= 'Z') || (nextCh >= 'g' && nextCh <= 'z') || (nextCh >= '0' && nextCh <= '9' && !hasValidHexDigits) {
			// This is an invalid hex number like $NE, $GG, $XY
			// Read the rest of the invalid identifier
			for l.position < len(l.input) {
				ch := l.peek()
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
					l.advance()
				} else {
					break
				}
			}
			literal = l.input[start:l.position]
			if l.debugMode {
				log.Warn("Invalid hex number found at %d:%d: %s", l.line, startCol, literal)
			}
			return l.createToken(TOKEN_ILLEGAL, literal, startCol, nil)
		}

		if hasValidHexDigits {
			literal = l.input[start:l.position]
			tokenType = TOKEN_NUMBER_HEX
		} else {
			// Just '$' with nothing after - not a number token
			l.position = start
			return nil
		}
	} else if ch == '%' {
		// Binary number (%xxxx)
		l.advance()
		digitStart := l.position
		for l.position < len(l.input) && (l.peek() == '0' || l.peek() == '1') {
			l.advance()
		}
		if l.position > digitStart {
			literal = l.input[start:l.position]
			tokenType = TOKEN_NUMBER_BIN
		} else {
			l.position = start
			return nil
		}
	} else if ch == '&' {
		// Octal number (&xxxx)
		l.advance()
		digitStart := l.position
		for l.position < len(l.input) && isOctalDigit(l.peek()) {
			l.advance()
		}
		if l.position > digitStart {
			literal = l.input[start:l.position]
			tokenType = TOKEN_NUMBER_OCT
		} else {
			l.position = start
			return nil
		}
	} else if ch == '\'' {
		// Character literal ('A') - converts to ASCII value
		l.advance() // skip opening '
		if l.position >= len(l.input) {
			l.position = start
			return nil
		}
		charByte := l.peek() // the character itself
		l.advance()

		// Check for closing '
		if l.peek() != '\'' {
			l.position = start
			return nil
		}
		l.advance() // skip closing '

		// Convert character to its ASCII decimal value
		// Return the numeric value as string, not 'A'
		asciiValue := fmt.Sprintf("%d", charByte)

		return &ContextToken{
			Type:     TOKEN_NUMBER_DEC,
			Literal:  asciiValue, // e.g. "65" for 'A'
			Line:     l.line,
			Column:   startCol,
			Context:  l.CurrentContext(),
			Metadata: &TokenMetadata{},
		}
	} else if isDigit(ch) {
		// Decimal number
		for l.position < len(l.input) && isDigit(l.peek()) {
			l.advance()
		}
		// Check for decimal point
		if l.peek() == '.' && isDigit(l.peekAhead(1)) {
			l.advance() // .
			for l.position < len(l.input) && isDigit(l.peek()) {
				l.advance()
			}
		}
		literal = l.input[start:l.position]
		tokenType = TOKEN_NUMBER_DEC
	} else {
		// Not a number
		l.position = start
		return nil
	}

	return &ContextToken{
		Type:     tokenType,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}
}

// tryTokenizeIdentifier attempts to tokenize an identifier
func (l *ContextAwareLexer) tryTokenizeIdentifier() *ContextToken {
	if !isAlpha(l.peek()) && l.peek() != '_' && l.peek() != '!' {
		return nil
	}

	start := l.position
	startCol := l.column

	if l.debugMode && l.line >= 80 && l.line <= 110 {
		log.Debug("tryTokenizeIdentifier: Line %d, l.column=%d, startCol=%d, start position=%d", l.line, l.column, startCol, start)
	}

	// Read identifier (allow dots for member access like Colors.BLUE)
	l.advance()
	for l.position < len(l.input) {
		ch := l.peek()
		if isAlphaNumeric(ch) || ch == '_' || ch == '.' {
			l.advance()
		} else {
			break
		}
	}

	literal := l.input[start:l.position]

	// Check for special keywords
	if literal == "else" {
		return &ContextToken{
			Type:     TOKEN_ELSE,
			Literal:  literal,
			Line:     l.line,
			Column:   startCol,
			Context:  l.CurrentContext(),
			Metadata: &TokenMetadata{},
		}
	}

	// Check if it's a Kick Assembler function or constant
	metadata := &TokenMetadata{}
	tokenType := TOKEN_IDENTIFIER

	if l.processorCtx != nil {
		// Check for function
		if funcInfo := l.processorCtx.GetFunctionInfo(literal); funcInfo != nil {
			metadata.IsKickFunction = true
			metadata.FunctionInfo = funcInfo
			tokenType = TOKEN_BUILTIN_MATH_FUNC // Default, could categorize by function category
		}

		// Check for constant
		if constInfo := l.processorCtx.GetConstantInfo(literal); constInfo != nil {
			metadata.IsKickConstant = true
			metadata.ConstantInfo = constInfo
			tokenType = TOKEN_BUILTIN_MATH_CONST // Default, could categorize by constant category
		}
	}

	return &ContextToken{
		Type:     tokenType,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: metadata,
	}
}

// tokenizeOperatorOrPunctuation handles operators and punctuation
func (l *ContextAwareLexer) tokenizeOperatorOrPunctuation() *ContextToken {
	startCol := l.column
	ch := l.peek()

	// Multi-character operators
	if ch == '<' && l.peekAhead(1) == '<' {
		startCol := l.column
		l.advance()
		l.advance()
		return l.createToken(TOKEN_LEFT_SHIFT, "<<", startCol, nil)
	}
	if ch == '>' && l.peekAhead(1) == '>' {
		startCol := l.column
		l.advance()
		l.advance()
		return l.createToken(TOKEN_RIGHT_SHIFT, ">>", startCol, nil)
	}

	// Single character operators/punctuation
	var tokenType TokenType
	var literal string

	switch ch {
	case ':':
		tokenType = TOKEN_COLON
	case '#':
		tokenType = TOKEN_HASH
	case '.':
		tokenType = TOKEN_DOT
	case ',':
		tokenType = TOKEN_COMMA
	case '+':
		tokenType = TOKEN_PLUS
	case '-':
		tokenType = TOKEN_MINUS
	case '*':
		tokenType = TOKEN_ASTERISK
	case '/':
		tokenType = TOKEN_SLASH
	case '(':
		tokenType = TOKEN_LPAREN
		l.parenDepth++
	case ')':
		tokenType = TOKEN_RPAREN
		if l.parenDepth > 0 {
			l.parenDepth--
		}
	case '[':
		tokenType = TOKEN_LBRACKET
	case ']':
		tokenType = TOKEN_RBRACKET
	case '{':
		tokenType = TOKEN_LBRACE
	case '}':
		tokenType = TOKEN_RBRACE
		// Pop block context when we see closing brace
		ctx := l.CurrentContext()
		if l.debugMode {
			log.Debug("tokenizeOperatorOrPunctuation: Found }, currentContext=%s, stack depth=%d",
				ctx.State.String(), len(l.contextStack))
		}
		if ctx.State == StateBlock {
			l.PopContext()
			if l.debugMode {
				log.Debug("tokenizeOperatorOrPunctuation: Popped StateBlock, new stack depth=%d", len(l.contextStack))
			}
		}
	case '=':
		tokenType = TOKEN_EQUAL
	case '<':
		tokenType = TOKEN_LESS
	case '>':
		tokenType = TOKEN_GREATER
	case '@':
		tokenType = TOKEN_AT
	case ';':
		tokenType = TOKEN_SEMICOLON
	case '&':
		tokenType = TOKEN_BITWISE_AND
	case '|':
		tokenType = TOKEN_BITWISE_OR
	case '^':
		tokenType = TOKEN_BITWISE_XOR
	case '%':
		tokenType = TOKEN_MODULO
	default:
		// Unknown character
		startCol := l.column
		l.advance()
		return l.createToken(TOKEN_ILLEGAL, string(ch), startCol, nil)
	}

	l.advance()
	literal = string(ch)

	return &ContextToken{
		Type:     tokenType,
		Literal:  literal,
		Line:     l.line,
		Column:   startCol,
		Context:  l.CurrentContext(),
		Metadata: &TokenMetadata{},
	}
}

// Character classification helpers
func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}