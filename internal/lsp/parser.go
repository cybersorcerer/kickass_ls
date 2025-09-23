package lsp

import (
	"fmt"
	"strconv"
	"strings"

	"c64.nvim/internal/log"
)

// ParseDocument is the entry point for parsing a document.
func ParseDocument(uri string, text string) (*Scope, []Diagnostic) {
	l := NewLexer(text)
	p := NewParser(l)
	program := p.ParseProgram()

	// Convert the AST to the Scope/Symbol structure
	scope, semanticDiagnostics := buildScopeFromAST(program, uri)

	// Combine parser diagnostics (e.g., syntax errors) with semantic diagnostics
	allDiagnostics := append(p.Errors(), semanticDiagnostics...)

	return scope, allDiagnostics
}

// scopeBuilder holds the state for the scope construction phase
type scopeBuilder struct {
	diagnostics []Diagnostic
}

// buildScopeFromAST walks the AST and builds the Scope/Symbol tree.
func buildScopeFromAST(program *Program, uri string) (*Scope, []Diagnostic) {
	rootScope := NewRootScope(uri)
	if program == nil {
		log.Warn("buildScopeFromAST: program is nil")
		return rootScope, []Diagnostic{}
	}

	builder := &scopeBuilder{diagnostics: []Diagnostic{}}
	builder.buildScope(program.Statements, rootScope)
	return rootScope, builder.diagnostics
}

func (sb *scopeBuilder) buildScope(statements []Statement, currentScope *Scope) {
	if statements == nil {
		log.Debug("buildScope: statements is nil")
		return
	}

	for _, statement := range statements {
		if statement == nil {
			log.Debug("buildScope: Encountered a nil statement, skipping.")
			continue
		}

		switch stmt := statement.(type) {
		case *InstructionStatement:
			sb.validateInstruction(stmt)
		case *LabelStatement:
			if stmt.Name == nil {
				log.Debug("buildScope: LabelStatement has nil Name, skipping.")
				continue
			}
			if stmt.Token.Line <= 0 || stmt.Token.Column <= 0 {
				log.Debug("buildScope: LabelStatement has invalid token position, using defaults")
				stmt.Token.Line = 1
				stmt.Token.Column = 1
			}
			symbol := &Symbol{
				Name: stmt.Name.Value,
				Kind: Label,
				Position: Position{
					Line:      stmt.Token.Line - 1,
					Character: stmt.Token.Column - 1,
				},
				Scope: currentScope,
			}
			if err := currentScope.AddSymbol(symbol); err != nil {
				diagnostic := Diagnostic{
					Severity: SeverityError,
					Range:    Range{Start: symbol.Position, End: Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)}},
					Message:  err.Error(),
					Source:   "parser",
				}
				sb.diagnostics = append(sb.diagnostics, diagnostic)
			}
		case *DirectiveStatement:
			if stmt == nil || stmt.Name == nil {
				log.Debug("buildScope: DirectiveStatement or its Name is nil, skipping.")
				continue
			}
			if stmt.Name.Token.Type == 0 {
				log.Debug("buildScope: DirectiveStatement has invalid token, creating default")
				stmt.Name.Token = Token{Type: TOKEN_IDENTIFIER, Literal: stmt.Name.Value, Line: 1, Column: 1}
			}
			if stmt.Name.Token.Line <= 0 || stmt.Name.Token.Column <= 0 {
				log.Debug("buildScope: DirectiveStatement has invalid token position, using defaults")
				stmt.Name.Token.Line = 1
				stmt.Name.Token.Column = 1
			}

			var kind SymbolKind
			var params []string
			var signature string

			if stmt.Value != nil {
				switch strings.ToLower(stmt.Token.Literal) {
				case ".const":
					kind = Constant
				case ".var":
					kind = Variable
				default:
					kind = UnknownSymbol
				}
			} else if stmt.Block != nil {
				switch strings.ToLower(stmt.Token.Literal) {
				case ".function":
					kind = Function
				case ".macro", ".pseudocommand":
					kind = Macro
				case ".namespace":
					kind = Namespace
				default:
					kind = UnknownSymbol
				}
				if kind != UnknownSymbol && stmt.Parameters != nil {
					for _, p := range stmt.Parameters {
						if p != nil {
							params = append(params, p.Value)
						}
					}
					signature = fmt.Sprintf("%s(%s)", stmt.Name.Value, strings.Join(params, ", "))
				}
			} else if strings.EqualFold(stmt.Token.Literal, ".label") {
				kind = Label
			}

			if kind != UnknownSymbol {
				log.Debug("buildScope: Creating symbol for directive '%s' with name '%s'", stmt.Token.Literal, stmt.Name.Value)
				value := ""
				if stmt.Value != nil {
					value = stmt.Value.TokenLiteral()
				}
				symbol := &Symbol{
					Name:      stmt.Name.Value,
					Kind:      kind,
					Value:     value,
					Params:    params,
					Signature: signature,
					Position: Position{
						Line:      stmt.Name.Token.Line - 1,
						Character: stmt.Name.Token.Column - 1,
					},
					Scope: currentScope,
				}
				if err := currentScope.AddSymbol(symbol); err != nil {
					diagnostic := Diagnostic{
						Severity: SeverityError,
						Range:    Range{Start: symbol.Position, End: Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)}},
						Message:  err.Error(),
						Source:   "parser",
					}
					sb.diagnostics = append(sb.diagnostics, diagnostic)
				}
			}

			if stmt.Block != nil && stmt.Name != nil && strings.ToLower(stmt.Token.Literal) == ".namespace" {
				log.Debug("buildScope: Creating new scope for directive '%s' with name '%s'", stmt.Token.Literal, stmt.Name.Value)
				if stmt.Block.Statements == nil {
					log.Debug("buildScope: Block.Statements is nil, initializing empty slice")
					stmt.Block.Statements = []Statement{}
				}
				endLine := stmt.Name.Token.Line - 1
				endChar := stmt.Name.Token.Column - 1
				if stmt.Block.EndToken.Type != TOKEN_EOF {
					endLine = stmt.Block.EndToken.Line - 1
					endChar = stmt.Block.EndToken.Column
				}
				newScope := &Scope{
					Name:     stmt.Name.Value,
					Parent:   currentScope,
					Children: make([]*Scope, 0),
					Symbols:  make(map[string]*Symbol),
					Uri:      currentScope.Uri,
					Range: Range{
						Start: Position{Line: stmt.Name.Token.Line - 1, Character: stmt.Name.Token.Column - 1},
						End:   Position{Line: endLine, Character: endChar},
					},
				}
				currentScope.AddChildScope(newScope)
				sb.buildScope(stmt.Block.Statements, newScope)
			}
		default:
			log.Debug("buildScope: Encountered unknown statement type: %T", statement)
		}
	}
}

func (sb *scopeBuilder) validateInstruction(stmt *InstructionStatement) {
	mode, err := determineAddressingModeFromAST(stmt.Operand)
	if err != nil {
		diagnostic := Diagnostic{
			Severity: SeverityError,
			Range:    Range{Start: Position{Line: stmt.Token.Line - 1, Character: stmt.Token.Column - 1}, End: Position{Line: stmt.Token.Line - 1, Character: stmt.Token.Column + len(stmt.Token.Literal)}},
			Message:  err.Error(),
			Source:   "parser",
		}
		sb.diagnostics = append(sb.diagnostics, diagnostic)
		return
	}

	if !isAddressingModeValid(stmt.Token.Literal, mode) {
		diagnostic := Diagnostic{
			Severity: SeverityError,
			Range:    Range{Start: Position{Line: stmt.Token.Line - 1, Character: stmt.Token.Column - 1}, End: Position{Line: stmt.Token.Line - 1, Character: stmt.Token.Column + len(stmt.Token.Literal)}},
			Message:  fmt.Sprintf("Invalid addressing mode '%s' for instruction '%s'", mode, strings.ToUpper(stmt.Token.Literal)),
			Source:   "parser",
		}
		sb.diagnostics = append(sb.diagnostics, diagnostic)
	}
}

// determineAddressingModeFromAST determines the addressing mode from an AST expression.
func determineAddressingModeFromAST(expr Expression) (string, error) {
	if expr == nil {
		return "Implied", nil
	}

	switch e := expr.(type) {
	case *PrefixExpression:
		if e.Operator == "#" {
			return "Immediate", nil
		}
	case *Identifier, *IntegerLiteral:
		// This could be zeropage or absolute. For now, we'll treat them as absolute
		// as it's a superset for validation purposes.
		return "Absolute", nil
	}

	// TODO: Implement more complex modes (indirect, indexed)
	return "unknown", fmt.Errorf("unrecognized operand structure")
}

// isAddressingModeValid checks if the given addressing mode is valid for the instruction.
func isAddressingModeValid(mnemonic string, mode string) bool {
	for _, m := range mnemonics {
		if strings.EqualFold(m.Mnemonic, mnemonic) {
			for _, am := range m.AddressingModes {
				if am.AddressingMode == mode {
					return true
				}
				// Allow zeropage operands for absolute addressing modes
				if mode == "Absolute" && (am.AddressingMode == "Zeropage" || am.AddressingMode == "Absolute") {
					return true
				}
			}
			return false
		}
	}
	return false // Mnemonic not found
}

// --- Abstract Syntax Tree (AST) --- //

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 && p.Statements[0] != nil {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

type BlockStatement struct {
	Token      Token
	Statements []Statement
	EndToken   Token
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

type InstructionStatement struct {
	Token   Token
	Operand Expression
}

func (is *InstructionStatement) statementNode()       {}
func (is *InstructionStatement) TokenLiteral() string { return is.Token.Literal }

type DirectiveStatement struct {
	Token      Token
	Name       *Identifier
	Parameters []*Identifier
	Value      Expression
	Block      *BlockStatement
}

func (ds *DirectiveStatement) statementNode()       {}
func (ds *DirectiveStatement) TokenLiteral() string { return ds.Token.Literal }

type LabelStatement struct {
	Token Token
	Name  *Identifier
}

func (ls *LabelStatement) statementNode()       {}
func (ls *LabelStatement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

type PrefixExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// --- Parser --- //

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or <X
	MEMBER      // object.member
)

var precedences = map[TokenType]int{
	TOKEN_EQUAL:    EQUALS,
	TOKEN_PLUS:     SUM,
	TOKEN_MINUS:    SUM,
	TOKEN_SLASH:    PRODUCT,
	TOKEN_ASTERISK: PRODUCT,
	TOKEN_LESS:     PREFIX,
	TOKEN_GREATER:  PREFIX,
	TOKEN_DOT:      MEMBER,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	lexer       *Lexer
	curToken    Token
	peekToken   Token
	diagnostics []Diagnostic

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		lexer:       l,
		diagnostics: []Diagnostic{},
	}

	p.prefixParseFns = make(map[TokenType]prefixParseFn)
	p.registerPrefix(TOKEN_IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(TOKEN_NUMBER_DEC, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_HEX, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_BIN, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_OCT, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_HASH, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_LESS, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_GREATER, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_DOT, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(TOKEN_ASTERISK, p.parseInfixExpression)
	p.registerInfix(TOKEN_EQUAL, p.parseInfixExpression)
	p.registerInfix(TOKEN_DOT, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) Errors() []Diagnostic {
	return p.diagnostics
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.diagnostics = append(p.diagnostics, Diagnostic{Message: msg, Range: Range{Start: Position{Line: p.peekToken.Line - 1, Character: p.peekToken.Column - 1}, End: Position{Line: p.peekToken.Line - 1, Character: p.peekToken.Column - 1 + len(p.peekToken.Literal)}}, Severity: SeverityError, Source: "parser"})
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = make([]Statement, 0)

	for p.curToken.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case TOKEN_MNEMONIC_STD, TOKEN_MNEMONIC_CTRL, TOKEN_MNEMONIC_ILL, TOKEN_MNEMONIC_65C02:
		return p.parseInstructionStatement()
	case TOKEN_LABEL:
		return p.parseLabelStatement()
	case TOKEN_DIRECTIVE_PC, TOKEN_DIRECTIVE_KICK_PRE, TOKEN_DIRECTIVE_KICK_FLOW, TOKEN_DIRECTIVE_KICK_ASM, TOKEN_DIRECTIVE_KICK_DATA, TOKEN_DIRECTIVE_KICK_TEXT:
		return p.parseDirectiveStatement()
	default:
		return nil
	}
}

func (p *Parser) parseInstructionStatement() *InstructionStatement {
	stmt := &InstructionStatement{Token: p.curToken}

	if !p.peekTokenIs(TOKEN_EOF) && p.peekToken.Type != TOKEN_COMMENT && p.curToken.Line == p.peekToken.Line {
		p.nextToken()
		stmt.Operand = p.parseExpression(LOWEST)
	} else {
		stmt.Operand = nil
	}

	return stmt
}

func (p *Parser) parseLabelStatement() *LabelStatement {
	stmt := &LabelStatement{Token: p.curToken}
	labelName := strings.TrimSuffix(p.curToken.Literal, ":")
	stmt.Name = &Identifier{Token: p.curToken, Value: labelName}
	return stmt
}

func (p *Parser) parseDirectiveStatement() *DirectiveStatement {
	stmt := &DirectiveStatement{Token: p.curToken}

	if strings.EqualFold(stmt.Token.Literal, ".label") && p.peekTokenIs(TOKEN_LABEL) {
		p.nextToken()
		stmt.Name = &Identifier{Token: p.curToken, Value: strings.TrimSuffix(p.curToken.Literal, ":")}
	} else if strings.EqualFold(stmt.Token.Literal, ".label") && p.peekTokenIs(TOKEN_DOT) {
		p.nextToken()
		if !p.expectPeek(TOKEN_IDENTIFIER) {
			stmt.Name = &Identifier{Token: Token{Type: TOKEN_IDENTIFIER, Literal: "unknown", Line: 1, Column: 1}, Value: "unknown"}
			return stmt
		}
		labelName := "." + p.curToken.Literal
		stmt.Name = &Identifier{Token: p.curToken, Value: labelName}
		if p.peekTokenIs(TOKEN_COLON) {
			p.nextToken()
		}
	} else {
		if !p.expectPeek(TOKEN_IDENTIFIER) {
			stmt.Name = &Identifier{Token: Token{Type: TOKEN_IDENTIFIER, Literal: "unknown", Line: 1, Column: 1}, Value: "unknown"}
			return stmt
		}
		stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(TOKEN_EQUAL) {
			p.nextToken()
			p.nextToken()
			stmt.Value = p.parseExpression(LOWEST)
			if stmt.Value == nil {
				p.diagnostics = append(p.diagnostics, Diagnostic{
					Message: "missing expression after '='",
					Range: Range{
						Start: Position{
							Line:      p.curToken.Line - 1,
							Character: p.curToken.Column - 1,
						},
						End: Position{
							Line:      p.curToken.Line - 1,
							Character: p.curToken.Column,
						},
					},
					Severity: SeverityError,
					Source:   "parser",
				})
			}
		} else if p.peekTokenIs(TOKEN_LPAREN) {
			p.nextToken()
			stmt.Parameters = p.parseIdentifierList()
			if p.peekTokenIs(TOKEN_LBRACE) {
				p.nextToken()
				stmt.Block = p.parseBlockStatement()
			}
		} else if p.peekTokenIs(TOKEN_LBRACE) {
			p.nextToken()
			stmt.Block = p.parseBlockStatement()
		}
	}

	return stmt
}

func (p *Parser) parseIdentifierList() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekTokenIs(TOKEN_RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(TOKEN_RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = make([]Statement, 0)

	p.nextToken()

	for p.curToken.Type != TOKEN_RBRACE && p.curToken.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	block.EndToken = p.curToken

	return block
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseExpression(precedence int) Expression {
	if p.curToken.Type == TOKEN_COMMENT {
		return nil
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.diagnostics = append(p.diagnostics, Diagnostic{
			Message: fmt.Sprintf("no prefix parse function for %s found", p.curToken.Type),
			Range: Range{
				Start: Position{
					Line:      p.curToken.Line - 1,
					Character: p.curToken.Column - 1,
				},
				End: Position{
					Line:      p.curToken.Line - 1,
					Character: p.curToken.Column - 1 + len(p.curToken.Literal),
				},
			},
			Severity: SeverityError,
			Source:   "parser",
		})
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(TOKEN_EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}
	var val int64
	var err error

	switch p.curToken.Type {
	case TOKEN_NUMBER_DEC:
		val, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
	case TOKEN_NUMBER_HEX:
		cleaned := strings.TrimPrefix(p.curToken.Literal, "#")
		cleaned = strings.TrimPrefix(cleaned, "$")
		val, err = strconv.ParseInt(cleaned, 16, 64)
	case TOKEN_NUMBER_BIN:
		cleaned := strings.TrimPrefix(p.curToken.Literal, "#")
		cleaned = strings.TrimPrefix(cleaned, "%")
		val, err = strconv.ParseInt(cleaned, 2, 64)
	case TOKEN_NUMBER_OCT:
		cleaned := strings.TrimPrefix(p.curToken.Literal, "#")
		cleaned = strings.TrimPrefix(cleaned, "&")
		val, err = strconv.ParseInt(cleaned, 8, 64)
	}

	if err != nil {
		//p.diagnostics = append(p.diagnostics, Diagnostic{Message: fmt.Sprintf("could not parse %q as integer", p.curToken.Literal), Range: Range{Start: Position{Line: p.curToken.Line - 1, Character: p.curToken.Column - 1}, End: Position{Line: p.curToken.Line - 1, Character: p.curToken.Column - 1 + len(p.curToken.Literal), Severity: SeverityError, Source: "parser"}})
		p.diagnostics = append(p.diagnostics, Diagnostic{
			Message: fmt.Sprintf("could not parse %q as integer", p.curToken.Literal),
			Range: Range{
				Start: Position{
					Line:      p.curToken.Line - 1,
					Character: p.curToken.Column - 1,
				},
				End: Position{
					Line:      p.curToken.Line - 1,
					Character: p.curToken.Column - 1 + len(p.curToken.Literal),
				},
			},
			Severity: SeverityError,
			Source:   "parser",
		})
		return nil
	}

	lit.Value = val
	return lit
}

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken() // Consume '('
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(TOKEN_RPAREN) {
		return nil
	}
	return exp
}
