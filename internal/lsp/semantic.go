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
		log.Debug("generateSemanticTokens: No symbol tree found, creating basic tokens")
	}
	
	// Create lexer to tokenize the text
	lexer := NewLexer(text)
	tokens := []int{}
	prevLine := 0
	prevChar := 0
	
	for {
		token := lexer.NextToken()
		if token.Type == TOKEN_EOF {
			break
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
			continue // Skip this token
		}
		
		// Calculate relative position (LSP semantic tokens format)
		deltaLine := line - prevLine
		deltaChar := char
		if deltaLine == 0 {
			deltaChar = char - prevChar
		}
		
		// Add token: [deltaLine, deltaChar, length, tokenType, modifiers]
		tokens = append(tokens, deltaLine, deltaChar, len(token.Literal), tokenType, modifiers)
		
		prevLine = line
		prevChar = char
	}
	
	log.Debug("generateSemanticTokens: Generated %d tokens", len(tokens)/5)
	return tokens
}

// getSemanticTokenType determines the semantic token type for a given token
func getSemanticTokenType(token Token, tree *Scope) (int, int) {
	switch token.Type {
	case TOKEN_MNEMONIC_STD, TOKEN_MNEMONIC_CTRL, TOKEN_MNEMONIC_ILL, TOKEN_MNEMONIC_65C02:
		return SemanticTokenKeyword, 0 // Opcodes as keywords
		
	case TOKEN_DIRECTIVE_PC, TOKEN_DIRECTIVE_KICK_PRE, TOKEN_DIRECTIVE_KICK_FLOW,
		 TOKEN_DIRECTIVE_KICK_ASM, TOKEN_DIRECTIVE_KICK_DATA, TOKEN_DIRECTIVE_KICK_TEXT:
		return SemanticTokenKeyword, SemanticTokenModifierDeclaration // Directives

	case TOKEN_ELSE:
		return SemanticTokenKeyword, 0 // else keyword for .if directives
		
	case TOKEN_NUMBER_HEX, TOKEN_NUMBER_BIN, TOKEN_NUMBER_DEC, TOKEN_NUMBER_OCT:
		return SemanticTokenNumber, 0 // Numbers
		
	case TOKEN_COMMENT:
		return SemanticTokenComment, 0 // Comments
		
	case TOKEN_STRING:
		return SemanticTokenString, 0 // Strings
		
	case TOKEN_LABEL:
		return SemanticTokenFunction, 0 // Labels (consistent with symbol-based labels)
		
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
		 TOKEN_ASTERISK, TOKEN_SLASH, TOKEN_EQUAL:
		return SemanticTokenOperator, 0 // Operators

	default:
		return -1, 0 // Skip this token
	}
}

// TokenType constants for semantic highlighting
const (
	SemanticTokenKeyword = iota        // 0: "keyword"
	SemanticTokenVariable              // 1: "variable"
	SemanticTokenFunction              // 2: "function"
	SemanticTokenMacro                 // 3: "macro"
	SemanticTokenPseudoCommand         // 4: "pseudocommand"
	SemanticTokenNumber                // 5: "number"
	SemanticTokenComment               // 6: "comment"
	SemanticTokenString                // 7: "string"
	SemanticTokenOperator              // 8: "operator"
)

// TokenModifier constants
const (
	SemanticTokenModifierDeclaration = iota
	SemanticTokenModifierReadonly
)

// encodeSemanticToken encodes a semantic token for LSP
func encodeSemanticToken(line, char, length, tokenType, modifiers int) []int {
	return []int{line, char, length, tokenType, modifiers}
}

// getTokenTypeForSymbol returns the appropriate token type for a symbol
func getTokenTypeForSymbol(kind SymbolKind) int {
	switch kind {
	case Constant:
		return SemanticTokenVariable
	case Variable:
		return SemanticTokenVariable
	case Label:
		return SemanticTokenFunction
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
