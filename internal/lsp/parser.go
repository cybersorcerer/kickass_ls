package lsp

import (
	"fmt"
	"strconv"
	"strings"

	"c64.nvim/internal/log"
)

// ParseDocument is the new entry point for parsing a document.
// It uses the new lexer and parser to build an AST, then converts
// that AST into the old Scope/Symbol structure for compatibility.
func ParseDocument(uri string, text string) (*Scope, []ParseError) {
	l := NewLexer(text)
	p := NewParser(l)
	program := p.ParseProgram()

	// Convert the new AST to the old Scope/Symbol structure.
	scope := buildScopeFromAST(program, uri)
	return scope, p.Errors()
}

// buildScopeFromAST walks the AST and builds the old Scope/Symbol tree.
func buildScopeFromAST(program *Program, uri string) *Scope {
	rootScope := NewRootScope(uri)
	buildScope(program.Statements, rootScope)
	return rootScope
}

func buildScope(statements []Statement, currentScope *Scope) {
	for _, statement := range statements {
		// Add a nil check to prevent panics on incomplete AST nodes from the parser.
		if statement == nil {
			log.Debug("buildScope: Encountered a nil statement, skipping.")
			continue
		}

		switch stmt := statement.(type) {
		case *InstructionStatement:
			// TODO: Create symbols from operand expressions if needed.
			_ = stmt // Placeholder
		case *LabelStatement:
			symbol := &Symbol{
				Name: stmt.Name.Value,
				Kind: Label,
				Position: Position{
					Line:      stmt.Token.Line,
					Character: stmt.Token.Column,
				},
				Scope: currentScope,
			}
			if err := currentScope.AddSymbol(symbol); err != nil {
				log.Warn("Failed to add symbol: %v", err)
			}
		case *DirectiveStatement:
			var kind SymbolKind
			var params []string
			var signature string

			// Handle directives that create symbols (.const, .var)
			if stmt.Value != nil {
				switch stmt.Token.Literal {
				case ".const":
					kind = Constant
				case ".var":
					kind = Variable
				default:
					kind = UnknownSymbol
				}
			} else if stmt.Block != nil {
				switch stmt.Token.Literal {
				case ".function":
					kind = Function
				case ".macro", ".pseudocommand":
					kind = Macro
				case ".var":
					kind = Variable
				default:
					kind = UnknownSymbol
				}

				if kind != UnknownSymbol && stmt.Name != nil {
					for _, p := range stmt.Parameters {
						params = append(params, p.Value)
					}
					signature = fmt.Sprintf("%s(%s)", stmt.Name.Value, strings.Join(params, ", "))
				}
			} else if strings.EqualFold(stmt.Token.Literal, ".label") {
				kind = Label
			}

			if kind != UnknownSymbol && stmt.Name != nil {
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
						Line:      stmt.Name.Token.Line,
						Character: stmt.Name.Token.Column,
					},
					Scope: currentScope,
				}
				if err := currentScope.AddSymbol(symbol); err != nil {
					log.Warn("Failed to add symbol: %v", err)
				}
			}

			// Handle directives that create scopes (.namespace, .function)
			if stmt.Block != nil && stmt.Name != nil {
				log.Debug("buildScope: Creating new scope for directive '%s' with name '%s'", stmt.Token.Literal, stmt.Name.Value)
				if stmt.Block == nil {
					log.Error("buildScope: stmt.Block is nil for '%s', this should not happen inside this if-block. Skipping scope creation.", stmt.Name.Value)
					continue
				}

				newScope := &Scope{
					Name:     stmt.Name.Value,
					Parent:   currentScope,
					Children: make([]*Scope, 0),
					Symbols:  make(map[string]*Symbol),
					Uri:      currentScope.Uri,
					Range: Range{
						Start: Position{Line: stmt.Name.Token.Line, Character: stmt.Name.Token.Column},
						// Set the end of the scope using the closing brace token from the block.
						// Set the end of the scope using the closing brace token from the block
						// to ensure the range is always valid.
						End: Position{Line: stmt.Block.EndToken.Line, Character: stmt.Block.EndToken.Column + 1},
					},
				}
				currentScope.AddChildScope(newScope)
				buildScope(stmt.Block.Statements, newScope) // Recursive call for the new scope
			}
		default:
			log.Debug("buildScope: Encountered unknown statement type: %T", statement)
		}
	}
}

// --- Abstract Syntax Tree (AST) --- //

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string // for debugging
}

// Statement is a sub-interface for all statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Expression is a sub-interface for all expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// BlockStatement represents a block of statements, e.g., inside { ... }.
type BlockStatement struct {
	Token      Token // The { LBRACE token
	Statements []Statement
	EndToken   Token // The } RBRACE token
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// InstructionStatement represents a CPU instruction (e.g., LDA #$10).
type InstructionStatement struct {
	Token   Token // The mnemonic token (e.g., TOKEN_MNEMONIC_STD)
	Operand Expression
}

func (is *InstructionStatement) statementNode()       {}
func (is *InstructionStatement) TokenLiteral() string { return is.Token.Literal }

// DirectiveStatement represents a directive (e.g., .const MAX = 10 or .namespace GFX { ... }).
type DirectiveStatement struct {
	Token      Token // The directive token (e.g., TOKEN_DIRECTIVE_KICK_DATA)
	Name       *Identifier
	Parameters []*Identifier // For functions, macros, pseudocommands
	Value      Expression
	Block      *BlockStatement
}

func (ds *DirectiveStatement) statementNode()       {}
func (ds *DirectiveStatement) TokenLiteral() string { return ds.Token.Literal }

// LabelStatement represents a label definition (e.g., start:).
type LabelStatement struct {
	Token Token // The TOKEN_LABEL token
	Name  *Identifier
}

func (ls *LabelStatement) statementNode()       {}
func (ls *LabelStatement) TokenLiteral() string { return ls.Token.Literal }

// Identifier represents an identifier used in an expression.
type Identifier struct {
	Token Token // The TOKEN_IDENTIFIER token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// IntegerLiteral represents a numeric literal.
type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// PrefixExpression represents an expression with a prefix operator (e.g., -5, <label).
type PrefixExpression struct {
	Token    Token // The prefix token, e.g., TOKEN_MINUS
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// InfixExpression represents an expression with an infix operator (e.g., 5 + 5).
type InfixExpression struct {
	Token    Token // The operator token, e.g., TOKEN_PLUS
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// --- Parser --- //

// ParseError represents a single error that occurred during parsing.
type ParseError struct {
	Message string
	Line    int
	Column  int
}

// Operator Precedence
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or <X
)

var precedences = map[TokenType]int{
	TOKEN_EQUAL:    EQUALS,
	TOKEN_PLUS:     SUM,
	TOKEN_MINUS:    SUM,
	TOKEN_SLASH:    PRODUCT,
	TOKEN_ASTERISK: PRODUCT,
	TOKEN_LESS:     PREFIX,
	TOKEN_GREATER:  PREFIX,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// Parser holds the state for the parsing process.
type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []ParseError

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

// NewParser creates a new Parser.
func NewParser(l *Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []ParseError{},
	}

	p.prefixParseFns = make(map[TokenType]prefixParseFn)
	p.registerPrefix(TOKEN_IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(TOKEN_NUMBER_DEC, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_HEX, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_BIN, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_NUMBER_OCT, p.parseIntegerLiteral)
	p.registerPrefix(TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_LESS, p.parsePrefixExpression)
	p.registerPrefix(TOKEN_GREATER, p.parsePrefixExpression)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(TOKEN_ASTERISK, p.parseInfixExpression)
	p.registerInfix(TOKEN_EQUAL, p.parseInfixExpression)

	// Read two tokens, so curToken and peekToken are both set.
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

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []ParseError {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %d, got %d instead", t, p.peekToken.Type)
	p.errors = append(p.errors, ParseError{Message: msg, Line: p.peekToken.Line, Column: p.peekToken.Column})
}

// ParseProgram is the main entry point for parsing.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	// The user's provided code snippet goes here.
	for p.curToken.Type != TOKEN_EOF {
		stmt := p.parseStatement()
		// Do not add nil statements to the AST.
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
		// If the current token doesn't start a known statement type,
		// we can try to parse it as an expression statement or simply skip.
		// For now, returning nil is okay as long as the caller (ParseProgram) handles it.
		// Let's try to parse an expression to handle lines that are just values.
		// return p.parseExpressionStatement() // Future enhancement
		return nil // Current behavior is to skip unknown statements.
	}
}

func (p *Parser) parseInstructionStatement() *InstructionStatement {
	stmt := &InstructionStatement{Token: p.curToken}

	// Only parse an operand if the next token is not EOF, not a comment,
	// and not another statement-starting token on a new line.
	// This prevents parsing expressions for instructions like `rts` that have no operands.
	if !p.peekTokenIs(TOKEN_EOF) && p.peekToken.Type != TOKEN_COMMENT && p.curToken.Line == p.peekToken.Line {
		p.nextToken() // Consume the mnemonic
		stmt.Operand = p.parseExpression(LOWEST)
	} else {
		// No operand present. The main loop's p.nextToken() will advance past the mnemonic.
		stmt.Operand = nil
	}

	return stmt
}

func (p *Parser) parseLabelStatement() *LabelStatement {
	stmt := &LabelStatement{Token: p.curToken}
	// The lexer regex for TOKEN_LABEL includes the trailing colon, remove it.
	labelName := strings.TrimSuffix(p.curToken.Literal, ":")
	stmt.Name = &Identifier{Token: p.curToken, Value: labelName}
	return stmt
}

func (p *Parser) parseDirectiveStatement() *DirectiveStatement {
	stmt := &DirectiveStatement{Token: p.curToken}

	if strings.EqualFold(stmt.Token.Literal, ".label") && p.peekTokenIs(TOKEN_LABEL) {
		p.nextToken() // consume .label, curToken is now the TOKEN_LABEL
		stmt.Name = &Identifier{Token: p.curToken, Value: strings.TrimSuffix(p.curToken.Literal, ":")}
	} else {
		if !p.expectPeek(TOKEN_IDENTIFIER) {
			return nil
		}
		stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(TOKEN_EQUAL) {
			p.nextToken() // consume IDENTIFIER or whatever was before =
			p.nextToken() // consume "="
			stmt.Value = p.parseExpression(LOWEST)
		} else if p.peekTokenIs(TOKEN_LPAREN) {
			p.nextToken() // consume IDENTIFIER
			stmt.Parameters = p.parseIdentifierList()
			// After parsing parameters, a block is expected for functions/macros,
			// but we should not assume it's always there for a valid (though incomplete) statement.
			if p.peekTokenIs(TOKEN_LBRACE) {
				p.nextToken() // consume ")" or whatever is before {
				stmt.Block = p.parseBlockStatement()
			}
		} else if p.peekTokenIs(TOKEN_LBRACE) {
			p.nextToken() // consume IDENTIFIER
			stmt.Block = p.parseBlockStatement()
		}
	}

	return stmt
}

func (p *Parser) parseIdentifierList() []*Identifier {
	identifiers := []*Identifier{}

	if !p.expectPeek(TOKEN_LPAREN) {
		return nil
	}

	if p.peekTokenIs(TOKEN_RPAREN) {
		p.nextToken() // consume ')'
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
	block.Statements = []Statement{}

	p.nextToken() // consume {

	for !p.curTokenIs(TOKEN_RBRACE) && !p.curTokenIs(TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	block.EndToken = p.curToken // Store the closing RBRACE token

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

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
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
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// no prefix parse function found
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != TOKEN_EOF && precedence < p.peekPrecedence() {
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
		val, err = strconv.ParseInt(strings.TrimPrefix(p.curToken.Literal, "$"), 16, 64)
	case TOKEN_NUMBER_BIN:
		val, err = strconv.ParseInt(strings.TrimPrefix(p.curToken.Literal, "%"), 2, 64)
	case TOKEN_NUMBER_OCT:
		val, err = strconv.ParseInt(strings.TrimPrefix(p.curToken.Literal, "&"), 8, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, ParseError{Message: msg, Line: p.curToken.Line, Column: p.curToken.Column})
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
