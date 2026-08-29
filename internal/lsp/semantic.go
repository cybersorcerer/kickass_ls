// internal/lsp/semantic.go
package lsp

import (
	"c64.nvim/internal/log"
)

// generateSemanticTokens creates semantic tokens for syntax highlighting
func generateSemanticTokens(uri string, text string) []int {
	log.Debug("generateSemanticTokens: Generating tokens for URI: %s", uri)

	// Get the symbol tree for context
	symbolStore.RLock()
	tree, exists := symbolStore.trees[uri]
	symbolStore.RUnlock()

	if !exists {
		// Analysing the unit fills the symbol store for every document in it.
		// Assigning to a fresh tree here instead used to leave the outer one
		// nil, so every identifier fell back to "variable".
		log.Debug("generateSemanticTokens: No symbol tree found, analysing the unit")
		analyzeUnit(uri)

		symbolStore.RLock()
		tree = symbolStore.trees[uri]
		symbolStore.RUnlock()
	}

	// Create context-aware lexer to tokenize the text
	lexer := NewContextAwareLexer(text, uri, globalProcessorContext)
	tokens := []int{}

	// Solution A: Dual Position Tracking (Single Pass)
	lastEmittedLine := 0
	lastEmittedChar := 0
	tokenCount := 0

	for {
		ctxToken := lexer.NextToken()
		if ctxToken.Type == TOKEN_EOF {
			break
		}

		// Convert ContextToken to Token for compatibility
		token := Token{
			Type:    ctxToken.Type,
			Literal: ctxToken.Literal,
			Line:    ctxToken.Line,
			Column:  ctxToken.Column,
		}

		// Convert 1-based to 0-based coordinates
		line := token.Line - 1
		char := token.Column - 1

		// Skip invalid positions
		if line < 0 || char < 0 {
			continue
		}

		// Get semantic token type
		tokenType, modifiers := getSemanticTokenType(token, tree)
		if tokenType == -1 {
			// Skip this token - don't update lastEmitted positions
			continue
		}

		// Special handling for hex/binary numbers with prefix ($d020, %10101010)
		// Split into two tokens: prefix (operator) + number
		if (token.Type == TOKEN_NUMBER_HEX || token.Type == TOKEN_NUMBER_BIN) && len(token.Literal) > 1 {
			prefix := token.Literal[0]
			if prefix == '$' || prefix == '%' {
				// Emit prefix as operator
				deltaLine := line - lastEmittedLine
				deltaChar := char
				if deltaLine == 0 {
					deltaChar = char - lastEmittedChar
				}
				tokens = append(tokens, deltaLine, deltaChar, 1, SemanticTokenOperator, 0)
				tokenCount++
				lastEmittedLine = line
				lastEmittedChar = char

				// Emit number part (without prefix)
				deltaLine = 0
				deltaChar = 1 // 1 char after prefix
				tokenLength := len(token.Literal) - 1
				tokens = append(tokens, deltaLine, deltaChar, tokenLength, tokenType, modifiers)
				tokenCount++
				lastEmittedLine = line
				lastEmittedChar = char + 1
				continue
			}
		}

		// Calculate delta from last EMITTED token
		deltaLine := line - lastEmittedLine
		deltaChar := char
		if deltaLine == 0 {
			deltaChar = char - lastEmittedChar
		}

		// Add token: [deltaLine, deltaChar, length, tokenType, modifiers]
		tokenLength := len(token.Literal)
		tokens = append(tokens, deltaLine, deltaChar, tokenLength, tokenType, modifiers)

		tokenCount++

		// Update last emitted position (start of THIS token)
		lastEmittedLine = line
		lastEmittedChar = char
	}

	log.Debug("generateSemanticTokens: Generated %d tokens", len(tokens)/5)

	return tokens
}

// getSemanticTokenType determines the semantic token type for a given token
func getSemanticTokenType(token Token, tree *Scope) (int, int) {
	switch token.Type {
	case TOKEN_MNEMONIC_STD, TOKEN_MNEMONIC_CTRL, TOKEN_MNEMONIC_ILL, TOKEN_MNEMONIC_65C02:
		return SemanticTokenMnemonic, 0 // Mnemonics (LDA, STA, JMP, etc.)

	case TOKEN_DIRECTIVE_PC, TOKEN_DIRECTIVE_KICK_FLOW,
		TOKEN_DIRECTIVE_KICK_ASM, TOKEN_DIRECTIVE_KICK_DATA, TOKEN_DIRECTIVE_KICK_TEXT:
		return SemanticTokenDirective, SemanticTokenModifierDeclaration // Directives (.byte, .const, etc.)

	case TOKEN_DIRECTIVE_KICK_PRE:
		// Preprocessor directives (#import, #define, etc.)
		if len(token.Literal) > 0 && token.Literal[0] == '#' {
			return SemanticTokenPreprocessor, 0
		}
		return SemanticTokenDirective, SemanticTokenModifierDeclaration // Other directives

	case TOKEN_ELSE:
		return SemanticTokenKeyword, 0 // else keyword for .if directives

	case TOKEN_NUMBER_HEX, TOKEN_NUMBER_BIN, TOKEN_NUMBER_DEC, TOKEN_NUMBER_OCT:
		return SemanticTokenNumber, 0 // Numbers

	case TOKEN_COMMENT, TOKEN_COMMENT_BLOCK:
		return SemanticTokenComment, 0 // Comments (line and block)

	case TOKEN_STRING:
		return SemanticTokenString, 0 // Strings

	case TOKEN_LABEL:
		return SemanticTokenLabel, 0 // Labels

	case TOKEN_IDENTIFIER:
		// Check if it's a known symbol
		if tree != nil {
			if symbol, found := tree.FindSymbol(token.Literal); found {
				return getTokenTypeForSymbol(symbol.Kind), 0
			}
		}
		return SemanticTokenVariable, 0 // Default to variable

	case TOKEN_BUILTIN_MATH_FUNC, TOKEN_BUILTIN_STRING_FUNC, TOKEN_BUILTIN_FILE_FUNC, TOKEN_BUILTIN_3D_FUNC:
		return SemanticTokenFunction, SemanticTokenModifierReadonly // Built-in functions

	case TOKEN_BUILTIN_MATH_CONST, TOKEN_BUILTIN_COLOR_CONST:
		return SemanticTokenVariable, SemanticTokenModifierReadonly // Built-in constants

	case TOKEN_HASH, TOKEN_LESS, TOKEN_GREATER, TOKEN_PLUS, TOKEN_MINUS,
		TOKEN_ASTERISK, TOKEN_SLASH, TOKEN_EQUAL,
		TOKEN_LEFT_SHIFT, TOKEN_RIGHT_SHIFT,
		TOKEN_BITWISE_AND, TOKEN_BITWISE_OR, TOKEN_BITWISE_XOR, TOKEN_BITWISE_NOT,
		TOKEN_MODULO,
		TOKEN_EQUAL_EQUAL, TOKEN_NOT_EQUAL, TOKEN_LESS_EQUAL, TOKEN_GREATER_EQUAL,
		TOKEN_LOGICAL_NOT, TOKEN_LOGICAL_AND, TOKEN_LOGICAL_OR,
		TOKEN_QUESTION:
		return SemanticTokenOperator, 0 // Operators

	case TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACKET, TOKEN_RBRACKET,
		TOKEN_LBRACE, TOKEN_RBRACE, TOKEN_COMMA, TOKEN_COLON, TOKEN_SEMICOLON, TOKEN_DOT:
		// Skip punctuation completely - let default editor highlighting handle them
		// Solution A ensures positions remain correct even when skipping tokens
		return -1, 0

	default:
		return -1, 0 // Skip unknown tokens
	}
}

// TokenType constants for semantic highlighting
const (
	SemanticTokenKeyword       = iota // 0: "keyword" (generic fallback)
	SemanticTokenVariable             // 1: "variable"
	SemanticTokenFunction             // 2: "function"
	SemanticTokenMacro                // 3: "macro"
	SemanticTokenPseudoCommand        // 4: "pseudocommand"
	SemanticTokenNumber               // 5: "number"
	SemanticTokenComment              // 6: "comment"
	SemanticTokenString               // 7: "string"
	SemanticTokenOperator             // 8: "operator"
	SemanticTokenMnemonic             // 9: "mnemonic"
	SemanticTokenDirective            // 10: "directive"
	SemanticTokenPreprocessor         // 11: "preprocessor"
	SemanticTokenLabel                // 12: "label"
)

// TokenModifier constants. The LSP transmits modifiers as a bit set indexed by
// the legend, not as an index into it: bit 0 is "declaration", bit 1 is
// "readonly". Numbering them with iota sent 0 (no modifier at all) for
// declarations and 1 ("declaration") for read only symbols.
const (
	SemanticTokenModifierDeclaration = 1 << iota // legend index 0
	SemanticTokenModifierReadonly                // legend index 1
)

// getTokenTypeForSymbol returns the appropriate token type for a symbol
func getTokenTypeForSymbol(kind SymbolKind) int {
	switch kind {
	case Constant:
		return SemanticTokenVariable
	case Variable:
		return SemanticTokenVariable
	case Label:
		return SemanticTokenLabel
	case Function:
		return SemanticTokenFunction
	case Macro:
		return SemanticTokenMacro
	case PseudoCommand:
		return SemanticTokenPseudoCommand
	default:
		return SemanticTokenKeyword
	}
}
