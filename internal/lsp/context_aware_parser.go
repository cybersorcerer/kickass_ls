package lsp

import (
	"fmt"
	"strconv"
	"strings"

	log "c64.nvim/internal/log"
)

// Context-Aware Parser for 6510/C64/Kick Assembler
// Uses ContextAwareLexer and produces Enhanced AST

// ContextAwareParser represents the new context-aware parser
type ContextAwareParser struct {
	lexer        *ContextAwareLexer
	currentToken *ContextToken
	peekToken    *ContextToken
	diagnostics  []Diagnostic
	processorCtx *ProcessorContext
	debugMode    bool
}

// NewContextAwareParser creates a new context-aware parser instance
func NewContextAwareParser(lexer *ContextAwareLexer, processorCtx *ProcessorContext) *ContextAwareParser {
	p := &ContextAwareParser{
		lexer:        lexer,
		diagnostics:  []Diagnostic{},
		processorCtx: processorCtx,
		debugMode:    IsParserDebugModeEnabled(),
	}

	// Read first two tokens
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances the parser to the next token
func (p *ContextAwareParser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()

	if p.debugMode && p.currentToken != nil {
		log.Debug("Parser: Token %s '%s' at Line %d, Col %d (State: %s)",
			p.currentToken.Type.String(),
			p.currentToken.Literal,
			p.currentToken.Line,
			p.currentToken.Column,
			p.currentToken.Context.State.String())
	}
}

// ParseProgram is the entry point for parsing
func (p *ContextAwareParser) ParseProgram() *Program {
	program := &Program{
		Statements: []Statement{},
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Starting program parsing")
	}

	for p.currentToken.Type != TOKEN_EOF {
		if p.currentToken == nil {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed %d statements", len(program.Statements))
	}

	return program
}

// parseStatement parses a single statement based on current token type and context
func (p *ContextAwareParser) parseStatement() Statement {
	if p.currentToken == nil {
		return nil
	}

	// Skip comments
	if p.currentToken.Type == TOKEN_COMMENT {
		return nil
	}

	switch p.currentToken.Type {
	case TOKEN_LABEL:
		return p.parseLabelStatement()

	case TOKEN_MNEMONIC_STD, TOKEN_MNEMONIC_CTRL, TOKEN_MNEMONIC_ILL:
		return p.parseInstructionStatement()

	case TOKEN_DIRECTIVE_PC:
		return p.parseProgramCounterDirective()

	case TOKEN_DIRECTIVE_KICK_PRE, TOKEN_DIRECTIVE_KICK_FLOW,
		TOKEN_DIRECTIVE_KICK_ASM, TOKEN_DIRECTIVE_KICK_DATA, TOKEN_DIRECTIVE_KICK_TEXT:
		return p.parseDirectiveStatement()

	case TOKEN_IDENTIFIER:
		// Check if this is a macro/function call (followed by '(')
		if p.peekToken != nil && p.peekToken.Type == TOKEN_LPAREN {
			return p.parseMacroOrFunctionCallStatement()
		}
		// Check if this is a pseudocommand call (followed by expressions with ':')
		// Pseudocommands have syntax like: mov #$05 : $80
		if p.peekToken != nil && p.peekToken.Type != TOKEN_EOF && p.peekToken.Type != TOKEN_COMMENT {
			// Check if there's content on the same line (could be pseudocommand arguments)
			if p.peekToken.Line == p.currentToken.Line {
				return p.parsePseudocommandCallStatement()
			}
		}
		// Otherwise, it's an unknown statement
		if p.debugMode {
			log.Debug("ContextAwareParser: Unknown identifier statement '%s' at Line %d",
				p.currentToken.Literal, p.currentToken.Line)
		}
		return nil

	default:
		// Unknown statement
		if p.debugMode {
			log.Debug("ContextAwareParser: Unknown statement type %s at Line %d",
				p.currentToken.Type.String(), p.currentToken.Line)
		}
		return nil
	}
}

// parseLabelStatement parses a label definition
func (p *ContextAwareParser) parseLabelStatement() *LabelStatement {
	stmt := &LabelStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
	}

	// Extract label name (remove trailing ':')
	labelName := strings.TrimSuffix(p.currentToken.Literal, ":")

	stmt.Name = &Identifier{
		Token: stmt.Token,
		Value: labelName,
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed label '%s' at Line %d", labelName, stmt.Token.Line)
	}

	return stmt
}

// parseInstructionStatement parses a 6510 instruction with addressing mode
func (p *ContextAwareParser) parseInstructionStatement() *InstructionStatement {
	stmt := &InstructionStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
	}

	// Store mnemonic info in token literal (for now, until we have enhanced AST)
	// The mnemonic is already in Token.Literal

	// Check if there's an operand by peeking at next token
	// Don't advance if next token is EOF or statement terminator
	if p.peekToken.Type != TOKEN_EOF && !p.isNextTokenStatementTerminator() {
		// Parse operand if present
		p.nextToken()
		operand := p.parseExpression(LOWEST)

		// Check for indexed addressing (,X or ,Y)
		if p.peekToken.Type == TOKEN_COMMA {
			p.nextToken() // move to comma
			p.nextToken() // move to potential X or Y

			if p.currentToken.Type == TOKEN_IDENTIFIER {
				indexReg := strings.ToUpper(p.currentToken.Literal)
				if indexReg == "X" || indexReg == "Y" {
					// Create InfixExpression for indexed addressing
					stmt.Operand = &InfixExpression{
						Token: Token{
							Type:    TOKEN_COMMA,
							Literal: ",",
							Line:    p.currentToken.Line,
							Column:  p.currentToken.Column,
						},
						Left:     operand,
						Operator: ",",
						Right: &Identifier{
							Token: Token{
								Type:    TOKEN_IDENTIFIER,
								Literal: indexReg,
								Line:    p.currentToken.Line,
								Column:  p.currentToken.Column,
							},
							Value: indexReg,
						},
					}
				} else {
					stmt.Operand = operand
				}
			} else {
				stmt.Operand = operand
			}
		} else {
			stmt.Operand = operand
		}
	}

	// TODO: Validate addressing mode using metadata from ContextToken
	// For now, we use the existing InstructionStatement structure

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed instruction '%s' at Line %d", stmt.Token.Literal, stmt.Token.Line)
	}

	return stmt
}

// parseProgramCounterDirective parses *= directive
func (p *ContextAwareParser) parseProgramCounterDirective() *DirectiveStatement {
	stmt := &DirectiveStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: "*=",
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: "*=",
		},
	}

	// Parse address expression
	p.nextToken()
	if p.currentToken.Type != TOKEN_EOF {
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed program counter directive at Line %d", stmt.Token.Line)
	}

	return stmt
}

// parseDirectiveStatement parses a Kick Assembler directive
func (p *ContextAwareParser) parseDirectiveStatement() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)

	// Special handling for new directive types (v0.9.7)
	switch directiveName {
	case ".encoding":
		return p.parseEncodingDirective()
	case ".define", ".undef":
		return p.parseDefineDirective()
	case ".import":
		return p.parseImportDirective()
	case ".function":
		return p.parseFunctionDirective()
	case ".macro":
		return p.parseMacroDirective()
	case ".pseudocommand":
		return p.parsePseudocommandDirective()
	case ".namespace":
		return p.parseNamespaceDirective()
	case ".enum":
		return p.parseEnumDirective()
	}

	// Special handling for data directives with comma-separated values
	if isDataDirective(directiveName) {
		return p.parseDataDirective()
	}

	// Special handling for .for loops
	if directiveName == ".for" {
		return p.parseForDirective()
	}

	// Special handling for .if/.else
	if directiveName == ".if" {
		return p.parseConditionalDirective()
	}

	// Special handling for named directives (.const, .var, .label)
	if isNamedDirective(directiveName) {
		return p.parseNamedDirective()
	}

	// Generic directive parsing
	stmt := &DirectiveStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: directiveName,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: directiveName,
		},
	}

	// Parse directive value/expression
	p.nextToken()
	if p.currentToken.Type != TOKEN_EOF && !p.isStatementTerminator() {
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed directive '%s' at Line %d", directiveName, stmt.Token.Line)
	}

	return stmt
}

// parseDataDirective parses data directives (.byte, .word, .text) with comma-separated values
func (p *ContextAwareParser) parseDataDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: directiveName,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: directiveName,
		},
	}

	// Parse comma-separated values
	values := []Expression{}
	p.nextToken()

	// Parse first value
	if p.currentToken.Type != TOKEN_EOF && !p.isStatementTerminator() {
		firstValue := p.parseExpression(LOWEST)
		if firstValue != nil {
			values = append(values, firstValue)
		}

		// Parse remaining comma-separated values
		for p.peekToken.Type == TOKEN_COMMA {
			p.nextToken() // consume current value
			p.nextToken() // consume comma

			if p.currentToken.Type == TOKEN_EOF || p.isStatementTerminator() {
				break
			}

			value := p.parseExpression(LOWEST)
			if value != nil {
				values = append(values, value)
			}
		}
	}

	// Store values as ArrayExpression
	if len(values) > 0 {
		stmt.Value = &ArrayExpression{
			Token:    stmt.Token,
			Elements: values,
		}
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed data directive '%s' with %d values at Line %d",
			directiveName, len(values), stmt.Token.Line)
	}

	return stmt
}

// parseNamedDirective parses directives with named identifiers (.const, .var, .label)
// Format: .const constant_name = value
func (p *ContextAwareParser) parseNamedDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Next token should be the identifier name (e.g., magic_number, counter)
	p.nextToken()

	if p.currentToken.Type == TOKEN_IDENTIFIER {
		// Set the name to the identifier, not the directive
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}

		// Check for '=' sign
		if p.peekToken.Type == TOKEN_EQUAL {
			p.nextToken() // consume identifier
			p.nextToken() // consume '='

			// Parse the value expression
			if p.currentToken.Type != TOKEN_EOF && !p.isStatementTerminator() {
				stmt.Value = p.parseExpression(LOWEST)
			}
		}
	} else {
		// No identifier found, use directive name as fallback
		stmt.Name = &Identifier{
			Token: Token{
				Type:    directiveToken.Type,
				Literal: directiveName,
				Line:    directiveToken.Line,
				Column:  directiveToken.Column,
			},
			Value: directiveName,
		}
	}

	if p.debugMode {
		if stmt.Name != nil {
			log.Debug("ContextAwareParser: Parsed named directive '%s' with name '%s' at Line %d",
				directiveName, stmt.Name.Value, stmt.Token.Line)
		}
	}

	return stmt
}

// parseForDirective parses .for loop directive
func (p *ContextAwareParser) parseForDirective() *DirectiveStatement {
	stmt := &DirectiveStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    TOKEN_DIRECTIVE_KICK_FLOW,
				Literal: ".for",
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: ".for",
		},
	}

	// Parse (var i = 0; i < 3; i++)
	// For now, skip the entire parameter list until we find the closing )
	p.nextToken()
	if p.currentToken.Type == TOKEN_LPAREN {
		parenDepth := 1
		p.nextToken() // skip (

		// Skip all tokens until we find the matching )
		for parenDepth > 0 && p.currentToken.Type != TOKEN_EOF {
			if p.currentToken.Type == TOKEN_LPAREN {
				parenDepth++
			} else if p.currentToken.Type == TOKEN_RPAREN {
				parenDepth--
			}
			if parenDepth > 0 {
				p.nextToken()
			}
		}
		// Now currentToken should be the closing )
	}

	// Parse block { }
	if p.peekToken.Type == TOKEN_LBRACE {
		p.nextToken() // move to {
		stmt.Block = p.parseBlockStatement()
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .for directive at Line %d", stmt.Token.Line)
	}

	return stmt
}

// parseConditionalDirective parses .if/.else directive
func (p *ContextAwareParser) parseConditionalDirective() *DirectiveStatement {
	stmt := &DirectiveStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    TOKEN_DIRECTIVE_KICK_FLOW,
				Literal: ".if",
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: ".if",
		},
	}

	// Parse condition
	p.nextToken()
	if p.currentToken.Type == TOKEN_LPAREN {
		stmt.Value = p.parseExpression(LOWEST)
	}

	// Parse then block
	if p.peekToken.Type == TOKEN_LBRACE {
		p.nextToken()
		stmt.Block = p.parseBlockStatement()
	}

	// TODO: Parse else block if present

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .if directive at Line %d", stmt.Token.Line)
	}

	return stmt
}

// parseBlockStatement parses a block { ... }
func (p *ContextAwareParser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Statements: []Statement{},
	}

	p.nextToken() // skip {

	for p.currentToken.Type != TOKEN_RBRACE && p.currentToken.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// Helper functions

// isStatementTerminator checks if current token terminates a statement
func (p *ContextAwareParser) isStatementTerminator() bool {
	return p.currentToken.Type == TOKEN_EOF ||
		p.currentToken.Type == TOKEN_LABEL ||
		p.currentToken.Type == TOKEN_MNEMONIC_STD ||
		p.currentToken.Type == TOKEN_MNEMONIC_CTRL ||
		p.currentToken.Type == TOKEN_MNEMONIC_ILL ||
		p.currentToken.Type == TOKEN_DIRECTIVE_PC ||
		strings.HasPrefix(p.currentToken.Literal, ".")
}

func (p *ContextAwareParser) isNextTokenStatementTerminator() bool {
	return p.peekToken.Type == TOKEN_EOF ||
		p.peekToken.Type == TOKEN_LABEL ||
		p.peekToken.Type == TOKEN_MNEMONIC_STD ||
		p.peekToken.Type == TOKEN_MNEMONIC_CTRL ||
		p.peekToken.Type == TOKEN_MNEMONIC_ILL ||
		p.peekToken.Type == TOKEN_DIRECTIVE_PC ||
		strings.HasPrefix(p.peekToken.Literal, ".")
}

// isDataDirective checks if a directive is a data directive
func isDataDirective(directive string) bool {
	dataDirectives := []string{".byte", ".word", ".dword", ".text", ".fill", ".align"}
	for _, d := range dataDirectives {
		if directive == d {
			return true
		}
	}
	return false
}

// isNamedDirective checks if a directive requires a name identifier (.const, .var, .label)
func isNamedDirective(directive string) bool {
	namedDirectives := []string{".const", ".var", ".label"}
	for _, d := range namedDirectives {
		if directive == d {
			return true
		}
	}
	return false
}

// validateAddressingMode validates if the addressing mode is valid for the mnemonic
func (p *ContextAwareParser) validateAddressingMode(stmt *InstructionStatement) {
	// TODO: Implement addressing mode validation using MnemonicInfo.AddressingModes
	// This will check if the parsed operand matches one of the valid addressing modes
}

// Errors returns all diagnostics collected during parsing
func (p *ContextAwareParser) Errors() []Diagnostic {
	return p.diagnostics
}

// addError adds a diagnostic error
func (p *ContextAwareParser) addError(message string, line, column int) {
	diagnostic := Diagnostic{
		Severity: SeverityError,
		Range: Range{
			Start: Position{Line: line - 1, Character: column - 1},
			End:   Position{Line: line - 1, Character: column + 10},
		},
		Message: message,
		Source:  "context-aware-parser",
	}
	p.diagnostics = append(p.diagnostics, diagnostic)
}

// Expression parsing methods adapted from old parser

// parseExpression parses expressions with operator precedence
func (p *ContextAwareParser) parseExpression(precedence int) Expression {
	if p.currentToken == nil || p.currentToken.Type == TOKEN_COMMENT {
		return nil
	}

	// Parse prefix expression
	var leftExp Expression

	switch p.currentToken.Type {
	case TOKEN_IDENTIFIER:
		leftExp = p.parseIdentifier()
	case TOKEN_NUMBER_DEC, TOKEN_NUMBER_HEX, TOKEN_NUMBER_BIN, TOKEN_NUMBER_OCT:
		leftExp = p.parseIntegerLiteral()
	case TOKEN_STRING:
		leftExp = p.parseStringLiteral()
	case TOKEN_HASH, TOKEN_MINUS, TOKEN_PLUS, TOKEN_LESS, TOKEN_GREATER, TOKEN_DOT, TOKEN_AT:
		leftExp = p.parsePrefixExpression()
	case TOKEN_LPAREN:
		leftExp = p.parseGroupedExpression()
	case TOKEN_BUILTIN_MATH_FUNC, TOKEN_BUILTIN_STRING_FUNC, TOKEN_BUILTIN_FILE_FUNC, TOKEN_BUILTIN_3D_FUNC:
		leftExp = p.parseBuiltinFunction()
	case TOKEN_BUILTIN_MATH_CONST, TOKEN_BUILTIN_COLOR_CONST:
		leftExp = p.parseBuiltinConstant()
	case TOKEN_ILLEGAL:
		// Provide context-aware error message for illegal tokens
		if p.debugMode {
			log.Debug("ContextAwareParser: TOKEN_ILLEGAL encountered - Literal='%s', Line=%d, Column=%d",
				p.currentToken.Literal, p.currentToken.Line, p.currentToken.Column)
		}
		var message string
		if strings.HasPrefix(p.currentToken.Literal, "$") || strings.HasPrefix(p.currentToken.Literal, "#$") {
			// This is an invalid hex number like $NE, $GG, etc.
			message = fmt.Sprintf("Invalid hex value '%s' - hex values must only contain digits 0-9 and letters A-F", p.currentToken.Literal)
		} else {
			message = fmt.Sprintf("Illegal character sequence '%s'", p.currentToken.Literal)
		}
		p.addError(message, p.currentToken.Line, p.currentToken.Column)
		return nil
	default:
		p.addError(fmt.Sprintf("Unexpected token '%s' in expression", p.currentToken.Literal),
			p.currentToken.Line, p.currentToken.Column)
		return nil
	}

	// Parse infix expressions
	for p.peekToken != nil && p.peekToken.Type != TOKEN_EOF && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case TOKEN_PLUS, TOKEN_MINUS, TOKEN_SLASH, TOKEN_ASTERISK, TOKEN_EQUAL, TOKEN_DOT:
			p.nextToken()
			leftExp = p.parseInfixExpression(leftExp)
		case TOKEN_LPAREN:
			p.nextToken()
			leftExp = p.parseCallExpression(leftExp)
		default:
			return leftExp
		}
	}

	return leftExp
}

// parseIdentifier parses an identifier
func (p *ContextAwareParser) parseIdentifier() Expression {
	return &Identifier{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Value: p.currentToken.Literal,
	}
}

// parseIntegerLiteral parses numeric literals
func (p *ContextAwareParser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
	}

	var val int64
	var err error
	literal := strings.TrimPrefix(p.currentToken.Literal, "#")

	switch p.currentToken.Type {
	case TOKEN_NUMBER_DEC:
		if strings.Contains(literal, ".") {
			floatVal, err := parseFloat(literal)
			if err == nil {
				val = int64(floatVal)
			}
		} else {
			val, err = parseInt(literal, 10)
		}
	case TOKEN_NUMBER_HEX:
		literal = strings.TrimPrefix(literal, "$")
		val, err = parseInt(literal, 16)
	case TOKEN_NUMBER_BIN:
		literal = strings.TrimPrefix(literal, "%")
		val, err = parseInt(literal, 2)
	case TOKEN_NUMBER_OCT:
		literal = strings.TrimPrefix(literal, "&")
		val, err = parseInt(literal, 8)
	}

	if err != nil {
		p.addError(fmt.Sprintf("Could not parse %s as integer", p.currentToken.Literal),
			p.currentToken.Line, p.currentToken.Column)
		return nil
	}

	lit.Value = val
	return lit
}

// parseStringLiteral parses string literals
func (p *ContextAwareParser) parseStringLiteral() Expression {
	// Remove quotes
	value := p.currentToken.Literal
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	return &StringLiteral{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Value: value,
	}
}

// parsePrefixExpression parses prefix expressions like #$00, -1, <addr, >addr
func (p *ContextAwareParser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Operator: p.currentToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseGroupedExpression parses expressions in parentheses
// Also handles 6502 indirect addressing modes: ($80, x) and ($80), y
func (p *ContextAwareParser) parseGroupedExpression() Expression {
	openParenLine := p.currentToken.Line

	p.nextToken() // skip (

	exp := p.parseExpression(LOWEST)

	// Check for Indexed Indirect addressing: ($80, x)
	// If we see a comma here, it could be ($addr, x) addressing mode
	if p.peekToken != nil && p.peekToken.Type == TOKEN_COMMA {
		p.nextToken() // move to comma

		// Check if next token is X or Y register
		if p.peekToken != nil && p.peekToken.Type == TOKEN_IDENTIFIER {
			p.nextToken() // move to identifier
			indexReg := strings.ToUpper(p.currentToken.Literal)

			if indexReg == "X" || indexReg == "Y" {
				// This is Indexed Indirect addressing: ($addr, X)
				// Construct an InfixExpression with the address and register
				indirectExp := &InfixExpression{
					Token: Token{
						Type:    TOKEN_COMMA,
						Literal: ",",
						Line:    p.currentToken.Line,
						Column:  p.currentToken.Column,
					},
					Left:     exp,
					Operator: ",",
					Right: &Identifier{
						Token: Token{
							Type:    TOKEN_IDENTIFIER,
							Literal: indexReg,
							Line:    p.currentToken.Line,
							Column:  p.currentToken.Column,
						},
						Value: indexReg,
					},
				}

				// Now expect the closing )
				if p.peekToken == nil || p.peekToken.Type != TOKEN_RPAREN {
					p.addError("Expected ')' after indexed indirect addressing",
						p.currentToken.Line, p.currentToken.Column)
					return nil
				}

				p.nextToken() // consume )

				if p.debugMode {
					log.Debug("Parser: Parsed Indexed Indirect addressing (addr, %s) at Line %d",
						indexReg, openParenLine)
				}

				return indirectExp
			}
		}

		// Not a valid indexed indirect - report error
		p.addError(fmt.Sprintf("Expected 'X' or 'Y' after ',' in indirect addressing, got '%s'",
			p.currentToken.Literal), p.currentToken.Line, p.currentToken.Column)
		return nil
	}

	// Normal grouped expression or regular indirect: ($addr) or ($addr), y
	if p.peekToken == nil || p.peekToken.Type != TOKEN_RPAREN {
		p.addError("Expected ')' after expression", p.currentToken.Line, p.currentToken.Column)
		return nil
	}

	p.nextToken() // consume )

	return exp
}

// parseInfixExpression parses infix expressions like a + b
func (p *ContextAwareParser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Operator: p.currentToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseCallExpression parses function calls
func (p *ContextAwareParser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Function: function,
	}

	exp.Arguments = p.parseExpressionList(TOKEN_RPAREN)

	return exp
}

// parseBuiltinFunction parses built-in function calls
func (p *ContextAwareParser) parseBuiltinFunction() Expression {
	funcToken := Token{
		Type:    p.currentToken.Type,
		Literal: p.currentToken.Literal,
		Line:    p.currentToken.Line,
		Column:  p.currentToken.Column,
	}

	if p.peekToken == nil || p.peekToken.Type != TOKEN_LPAREN {
		p.addError(fmt.Sprintf("Expected '(' after function '%s'", p.currentToken.Literal),
			p.currentToken.Line, p.currentToken.Column)
		return nil
	}

	p.nextToken() // move to LPAREN

	callExp := &CallExpression{
		Token:    funcToken,
		Function: &Identifier{Token: funcToken, Value: funcToken.Literal},
	}

	callExp.Arguments = p.parseExpressionList(TOKEN_RPAREN)

	return callExp
}

// parseBuiltinConstant parses built-in constants
func (p *ContextAwareParser) parseBuiltinConstant() Expression {
	return &Identifier{
		Token: Token{
			Type:    p.currentToken.Type,
			Literal: p.currentToken.Literal,
			Line:    p.currentToken.Line,
			Column:  p.currentToken.Column,
		},
		Value: p.currentToken.Literal,
	}
}

// parseExpressionList parses a comma-separated list of expressions
func (p *ContextAwareParser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}

	if p.peekToken != nil && p.peekToken.Type == end {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekToken != nil && p.peekToken.Type == TOKEN_COMMA {
		p.nextToken() // move to current expression end
		p.nextToken() // move past comma

		if p.currentToken.Type == end {
			break
		}

		list = append(list, p.parseExpression(LOWEST))
	}

	if p.peekToken != nil && p.peekToken.Type == end {
		p.nextToken()
	}

	return list
}

// Precedence helpers

func (p *ContextAwareParser) peekPrecedence() int {
	if p.peekToken == nil {
		return LOWEST
	}
	return precedences[p.peekToken.Type]
}

func (p *ContextAwareParser) curPrecedence() int {
	if p.currentToken == nil {
		return LOWEST
	}
	return precedences[p.currentToken.Type]
}

// parseEncodingDirective handles .encoding "string"
// Format: .encoding "petscii_upper"
func (p *ContextAwareParser) parseEncodingDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    directiveToken.Type,
				Literal: directiveName,
				Line:    directiveToken.Line,
				Column:  directiveToken.Column,
			},
			Value: directiveName,
		},
	}

	// Move to next token (should be string)
	p.nextToken()

	// Expect string literal
	if p.currentToken.Type == TOKEN_STRING {
		// Remove quotes from string value
		value := p.currentToken.Literal
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		stmt.Value = &StringLiteral{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: value,
		}
	} else {
		// Error: expected string
		p.addError("Expected string literal after .encoding directive", p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .encoding directive with value '%s' at Line %d",
			p.currentToken.Literal, stmt.Token.Line)
	}

	return stmt
}

// parseDefineDirective handles .define and .undef (symbol-only, no value)
// Format: .define DEBUG
func (p *ContextAwareParser) parseDefineDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be identifier)
	p.nextToken()

	// Expect identifier (the symbol name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		// No value for .define/.undef - just the symbol name
		stmt.Value = nil
	} else {
		// Error: expected identifier
		p.addError(fmt.Sprintf("Expected identifier after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		// Use directive name as fallback
		stmt.Name = &Identifier{
			Token: Token{
				Type:    directiveToken.Type,
				Literal: directiveName,
				Line:    directiveToken.Line,
				Column:  directiveToken.Column,
			},
			Value: directiveName,
		}
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed %s directive for symbol '%s' at Line %d",
			directiveName, stmt.Name.Value, stmt.Token.Line)
	}

	return stmt
}

// parseImportDirective handles .import source "file.asm"
// Format: .import source "file.asm"
//         .import binary "data.bin"
//         .import c64 "music.sid"
func (p *ContextAwareParser) parseImportDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
		Name: &Identifier{
			Token: Token{
				Type:    directiveToken.Type,
				Literal: directiveName,
				Line:    directiveToken.Line,
				Column:  directiveToken.Column,
			},
			Value: directiveName,
		},
	}

	// Move to next token (should be keyword: source, binary, c64)
	p.nextToken()

	var importType string
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		importType = strings.ToLower(p.currentToken.Literal)
		// Valid import types
		if importType != "source" && importType != "binary" && importType != "c64" {
			p.addError(fmt.Sprintf("Invalid import type '%s', expected 'source', 'binary', or 'c64'", importType), p.currentToken.Line, p.currentToken.Column)
		}
		p.nextToken() // Move to string
	}

	// Expect string literal (filename)
	if p.currentToken.Type == TOKEN_STRING {
		// Create an array expression to store both import type and filename
		elements := []Expression{}

		// Add import type as first element (if present)
		if importType != "" {
			elements = append(elements, &Identifier{
				Token: Token{
					Type:    TOKEN_IDENTIFIER,
					Literal: importType,
					Line:    p.currentToken.Line,
					Column:  p.currentToken.Column,
				},
				Value: importType,
			})
		}

		// Add filename as second element
		// Remove quotes from string value
		value := p.currentToken.Literal
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		elements = append(elements, &StringLiteral{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: value,
		})

		stmt.Value = &ArrayExpression{
			Token: Token{
				Type:    directiveToken.Type,
				Literal: directiveToken.Literal,
				Line:    directiveToken.Line,
				Column:  directiveToken.Column,
			},
			Elements: elements,
		}
	} else {
		// Error: expected string
		p.addError("Expected string literal (filename) in .import directive", p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .import directive at Line %d", stmt.Token.Line)
	}

	return stmt
}

// parseFunctionDirective handles .function name(params) { ... }
// Format: .function getColor(x) { .return Colors.RED }
func (p *ContextAwareParser) parseFunctionDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be function name)
	p.nextToken()

	// Expect identifier (function name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		p.nextToken() // Move to parameter list or block
	} else {
		p.addError(fmt.Sprintf("Expected function name after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		return stmt
	}

	// Parse parameter list if present: (param1, param2, ...)
	if p.currentToken.Type == TOKEN_LPAREN {
		stmt.Parameters = p.parseParameterList()
		// parseParameterList leaves us at the closing ), so move to next token
		p.nextToken()
	}

	// Parse block if present: { ... }
	if p.currentToken.Type == TOKEN_LBRACE {
		stmt.Block = p.parseBlockStatement()
	} else {
		p.addError(fmt.Sprintf("Expected '{' after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .function directive '%s' at Line %d", stmt.Name.Value, stmt.Token.Line)
	}

	return stmt
}

// parseMacroDirective handles .macro name(params) { ... }
// Format: .macro clearScreen() { lda #$20 }
func (p *ContextAwareParser) parseMacroDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be macro name)
	p.nextToken()

	// Expect identifier (macro name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		p.nextToken() // Move to parameter list or block
	} else {
		p.addError(fmt.Sprintf("Expected macro name after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		return stmt
	}

	// Parse parameter list if present: (param1, param2, ...)
	if p.currentToken.Type == TOKEN_LPAREN {
		stmt.Parameters = p.parseParameterList()
		// parseParameterList leaves us at the closing ), so move to next token
		p.nextToken()
	}

	// Parse block if present: { ... }
	if p.currentToken.Type == TOKEN_LBRACE {
		stmt.Block = p.parseBlockStatement()
	} else {
		p.addError(fmt.Sprintf("Expected '{' after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .macro directive '%s' at Line %d", stmt.Name.Value, stmt.Token.Line)
	}

	return stmt
}

// parsePseudocommandDirective handles .pseudocommand name param1 : param2 : param3 { ... }
// Format: .pseudocommand mov src : dst { lda src; sta dst }
// Pseudocommands use colon-separated parameters instead of parenthesized comma-separated
func (p *ContextAwareParser) parsePseudocommandDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be pseudocommand name)
	p.nextToken()

	// Expect identifier (pseudocommand name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		p.nextToken() // Move to parameter list or block
	} else {
		p.addError(fmt.Sprintf("Expected pseudocommand name after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		return stmt
	}

	// Parse colon-separated parameter list: param1 : param2 : param3
	// Stop when we hit '{'
	if p.currentToken.Type != TOKEN_LBRACE {
		stmt.Parameters = p.parseColonSeparatedParameterList()
	}

	// Parse block if present: { ... }
	if p.currentToken.Type == TOKEN_LBRACE {
		stmt.Block = p.parseBlockStatement()
	} else {
		p.addError(fmt.Sprintf("Expected '{' after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .pseudocommand directive '%s' with %d parameters at Line %d",
			stmt.Name.Value, len(stmt.Parameters), stmt.Token.Line)
	}

	return stmt
}

// parseNamespaceDirective handles .namespace name { ... }
// Format: .namespace utils { ... }
// Namespaces have no parameters, just a name and a block
func (p *ContextAwareParser) parseNamespaceDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be namespace name)
	p.nextToken()

	// Expect identifier (namespace name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		p.nextToken() // Move to block
	} else {
		p.addError(fmt.Sprintf("Expected namespace name after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		return stmt
	}

	// Parse block: { ... }
	if p.currentToken.Type == TOKEN_LBRACE {
		stmt.Block = p.parseBlockStatement()
	} else {
		p.addError(fmt.Sprintf("Expected '{' after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .namespace directive '%s' at Line %d", stmt.Name.Value, stmt.Token.Line)
	}

	return stmt
}

// parseEnumDirective handles .enum name { members }
// Format: .enum Colors { BLACK = 0, WHITE = 1 }
func (p *ContextAwareParser) parseEnumDirective() *DirectiveStatement {
	directiveName := strings.ToLower(p.currentToken.Literal)
	directiveToken := p.currentToken

	stmt := &DirectiveStatement{
		Token: Token{
			Type:    directiveToken.Type,
			Literal: directiveToken.Literal,
			Line:    directiveToken.Line,
			Column:  directiveToken.Column,
		},
	}

	// Move to next token (should be enum name)
	p.nextToken()

	// Expect identifier (enum name)
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		stmt.Name = &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		}
		p.nextToken() // Move to block
	} else {
		p.addError(fmt.Sprintf("Expected enum name after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
		return stmt
	}

	// Parse block: { ... }
	if p.currentToken.Type == TOKEN_LBRACE {
		stmt.Block = p.parseBlockStatement()
	} else {
		p.addError(fmt.Sprintf("Expected '{' after %s directive", directiveName), p.currentToken.Line, p.currentToken.Column)
	}

	if p.debugMode {
		log.Debug("ContextAwareParser: Parsed .enum directive '%s' at Line %d", stmt.Name.Value, stmt.Token.Line)
	}

	return stmt
}

// parseParameterList parses (param1, param2, ...)
// Returns the parameter list and leaves the parser at the closing )
func (p *ContextAwareParser) parseParameterList() []*Identifier {
	params := []*Identifier{}

	p.nextToken() // Skip opening (

	// Empty parameter list
	if p.currentToken.Type == TOKEN_RPAREN {
		return params
	}

	// Parse first parameter
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		params = append(params, &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		})
		p.nextToken()
	}

	// Parse remaining parameters: , param2, param3, ...
	for p.currentToken.Type == TOKEN_COMMA {
		p.nextToken() // Skip comma

		if p.currentToken.Type == TOKEN_IDENTIFIER {
			params = append(params, &Identifier{
				Token: Token{
					Type:    p.currentToken.Type,
					Literal: p.currentToken.Literal,
					Line:    p.currentToken.Line,
					Column:  p.currentToken.Column,
				},
				Value: p.currentToken.Literal,
			})
			p.nextToken()
		} else {
			p.addError("Expected parameter name after ','", p.currentToken.Line, p.currentToken.Column)
			break
		}
	}

	// Expect closing )
	if p.currentToken.Type != TOKEN_RPAREN {
		p.addError("Expected ')' after parameter list", p.currentToken.Line, p.currentToken.Column)
	}

	return params
}

// parseColonSeparatedParameterList parses param1 : param2 : param3
// Used by .pseudocommand directive
// Returns the parameter list and leaves the parser at the '{'
func (p *ContextAwareParser) parseColonSeparatedParameterList() []*Identifier {
	params := []*Identifier{}

	// Empty parameter list (directly followed by '{')
	if p.currentToken.Type == TOKEN_LBRACE {
		return params
	}

	// Parse first parameter
	if p.currentToken.Type == TOKEN_IDENTIFIER {
		params = append(params, &Identifier{
			Token: Token{
				Type:    p.currentToken.Type,
				Literal: p.currentToken.Literal,
				Line:    p.currentToken.Line,
				Column:  p.currentToken.Column,
			},
			Value: p.currentToken.Literal,
		})
		p.nextToken()
	}

	// Parse remaining parameters: : param2 : param3 ...
	for p.currentToken.Type == TOKEN_COLON {
		p.nextToken() // Skip colon

		if p.currentToken.Type == TOKEN_IDENTIFIER {
			params = append(params, &Identifier{
				Token: Token{
					Type:    p.currentToken.Type,
					Literal: p.currentToken.Literal,
					Line:    p.currentToken.Line,
					Column:  p.currentToken.Column,
				},
				Value: p.currentToken.Literal,
			})
			p.nextToken()
		} else {
			p.addError("Expected parameter name after ':'", p.currentToken.Line, p.currentToken.Column)
			break
		}
	}

	// currentToken should now be at '{' or error
	if p.currentToken.Type != TOKEN_LBRACE {
		p.addError("Expected '{' after parameter list", p.currentToken.Line, p.currentToken.Column)
	}

	return params
}

// parseMacroOrFunctionCallStatement parses macro/function calls as statements
// Format: macroName(arg1, arg2, ...)
func (p *ContextAwareParser) parseMacroOrFunctionCallStatement() Statement {
	// Current token is the identifier (macro/function name)
	nameToken := Token{
		Type:    p.currentToken.Type,
		Literal: p.currentToken.Literal,
		Line:    p.currentToken.Line,
		Column:  p.currentToken.Column,
	}

	functionIdent := &Identifier{
		Token: nameToken,
		Value: p.currentToken.Literal,
	}

	// Move to LPAREN
	p.nextToken()

	// Parse as CallExpression
	callExpr := &CallExpression{
		Token:    nameToken,
		Function: functionIdent,
	}

	// Parse arguments
	callExpr.Arguments = p.parseExpressionList(TOKEN_RPAREN)

	// Wrap in ExpressionStatement
	stmt := &ExpressionStatement{
		Token:      nameToken,
		Expression: callExpr,
	}

	// Always log macro/function calls for debugging
	log.Debug("ContextAwareParser: Parsed macro/function call '%s' with %d arguments at Line %d",
		functionIdent.Value, len(callExpr.Arguments), nameToken.Line)

	return stmt
}

// parsePseudocommandCallStatement parses pseudocommand calls as statements
// Format: pseudocmdName arg1 : arg2 : arg3
// Pseudocommands use colon-separated arguments instead of comma-separated
func (p *ContextAwareParser) parsePseudocommandCallStatement() Statement {
	// Current token is the identifier (pseudocommand name)
	nameToken := Token{
		Type:    p.currentToken.Type,
		Literal: p.currentToken.Literal,
		Line:    p.currentToken.Line,
		Column:  p.currentToken.Column,
	}

	pseudocmdIdent := &Identifier{
		Token: nameToken,
		Value: p.currentToken.Literal,
	}

	// Move to first argument
	p.nextToken()

	// Parse colon-separated arguments
	arguments := []Expression{}

	// Parse first argument
	if p.currentToken.Type != TOKEN_EOF && p.currentToken.Type != TOKEN_COMMENT {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			arguments = append(arguments, arg)
		}
	}

	// Parse remaining arguments separated by ':'
	for p.peekToken != nil && p.peekToken.Type == TOKEN_COLON {
		p.nextToken() // Move to COLON
		p.nextToken() // Move to next argument

		if p.currentToken.Type != TOKEN_EOF && p.currentToken.Type != TOKEN_COMMENT {
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				arguments = append(arguments, arg)
			}
		}
	}

	// Create CallExpression for pseudocommand
	callExpr := &CallExpression{
		Token:     nameToken,
		Function:  pseudocmdIdent,
		Arguments: arguments,
	}

	// Wrap in ExpressionStatement
	stmt := &ExpressionStatement{
		Token:      nameToken,
		Expression: callExpr,
	}

	log.Debug("ContextAwareParser: Parsed pseudocommand call '%s' with %d arguments at Line %d",
		pseudocmdIdent.Value, len(arguments), nameToken.Line)

	return stmt
}

// Helper functions for parsing

func parseInt(s string, base int) (int64, error) {
	return strconv.ParseInt(s, base, 64)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
