// internal/lsp/semantic.go
package lsp

import (
	"c64.nvim/internal/log"
	"unicode/utf16"
)

// generateSemanticTokens creates semantic tokens for syntax highlighting
func generateSemanticTokens(uri string, text string) []int {
	log.Debug("generateSemanticTokens: Generating tokens for URI: %s", uri)

	// Get the symbol tree for context
	symbolStore.RLock()
	tree, exists := symbolStore.trees[uri]
	symbolStore.RUnlock()

	if !exists {
		log.Debug("generateSemanticTokens: No symbol tree found, parsing document now")
		// Parse document to get symbol tree
		tree, _ = ParseDocument(uri, text)
		// Store it for future use
		symbolStore.Lock()
		symbolStore.trees[uri] = tree
		symbolStore.Unlock()
	}
	
	// Create context-aware lexer to tokenize the text
	lexer := NewContextAwareLexer(text, globalProcessorContext)
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

		// Debug: log tokens around problematic areas
		// Lines 80-108 (all enums), lines 246-249 (pseudocommand calls), line 232 (clearScreen macro call)
		if tokenCount <= 50 || (line >= 79 && line <= 108) || (line >= 231 && line <= 233) || (line >= 246 && line <= 249) {
			log.Debug("SemanticToken[%d]: '%s' L%d:C%d delta(%d,%d) len=%d type=%d tokenType=%d | prevL=%d prevC=%d | RAW=[%d,%d,%d,%d,%d]",
				tokenCount, token.Literal, line+1, char+1, deltaLine, deltaChar, tokenLength, tokenType, token.Type, lastEmittedLine+1, lastEmittedChar+1,
				deltaLine, deltaChar, tokenLength, tokenType, modifiers)
		}

		// Update last emitted position (start of THIS token)
		lastEmittedLine = line
		lastEmittedChar = char
	}
	
	log.Debug("generateSemanticTokens: Generated %d tokens", len(tokens)/5)

	// Debug: Output raw token values for enum section (tokens 875-900 = values 4375-4500)
	if len(tokens) >= 900 {
		log.Debug("Raw LSP tokens for enum section (values 875-900): %v", tokens[875:900])
	}

	return tokens
}

// getSemanticTokenType determines the semantic token type for a given token
func getSemanticTokenType(token Token, tree *Scope) (int, int) {
	switch token.Type {
	case TOKEN_MNEMONIC_STD, TOKEN_MNEMONIC_CTRL, TOKEN_MNEMONIC_ILL, TOKEN_MNEMONIC_65C02:
		return SemanticTokenKeyword, 0 // Opcodes as keywords
		
	case TOKEN_DIRECTIVE_PC, TOKEN_DIRECTIVE_KICK_FLOW,
		 TOKEN_DIRECTIVE_KICK_ASM, TOKEN_DIRECTIVE_KICK_DATA, TOKEN_DIRECTIVE_KICK_TEXT:
		return SemanticTokenKeyword, SemanticTokenModifierDeclaration // Directives

	case TOKEN_DIRECTIVE_KICK_PRE:
		// Preprocessor directives (#import, #define, etc.) get macro highlighting
		if len(token.Literal) > 0 && token.Literal[0] == '#' {
			return SemanticTokenMacro, 0
		}
		return SemanticTokenKeyword, SemanticTokenModifierDeclaration // Other pre directives

	case TOKEN_ELSE:
		return SemanticTokenKeyword, 0 // else keyword for .if directives
		
	case TOKEN_NUMBER_HEX, TOKEN_NUMBER_BIN, TOKEN_NUMBER_DEC, TOKEN_NUMBER_OCT:
		return SemanticTokenNumber, 0 // Numbers
		
	case TOKEN_COMMENT, TOKEN_COMMENT_BLOCK:
		return SemanticTokenComment, 0 // Comments (line and block)
		
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

// utf16Length returns the length of a string in UTF-16 code units
func utf16Length(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// utf8ToUTF16Offset converts a UTF-8 byte offset on a line to UTF-16 code unit offset
func utf8ToUTF16Offset(text string, targetLine int, byteOffset int) int {
	// Find the start of the target line
	currentLine := 0
	lineStart := 0

	for i := 0; i < len(text); i++ {
		if currentLine == targetLine {
			// We're on the target line, extract substring up to byteOffset
			lineEnd := lineStart
			for lineEnd < len(text) && text[lineEnd] != '\n' {
				lineEnd++
			}

			// Calculate the byte position on this line
			targetPos := lineStart + byteOffset
			if targetPos > lineEnd {
				targetPos = lineEnd
			}

			// Convert the substring from line start to target position to UTF-16
			substring := text[lineStart:targetPos]
			return utf16Length(substring)
		}

		if text[i] == '\n' {
			currentLine++
			lineStart = i + 1
		}
	}

	// Last line without newline
	if currentLine == targetLine {
		targetPos := lineStart + byteOffset
		if targetPos > len(text) {
			targetPos = len(text)
		}
		substring := text[lineStart:targetPos]
		return utf16Length(substring)
	}

	return 0
}
