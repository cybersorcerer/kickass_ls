package lsp

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

type ExpressionStatement struct {
	Token      Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

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

type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

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

type GroupedExpression struct {
	Token      Token // The '(' token
	Expression Expression
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }

type CallExpression struct {
	Token     Token // The '(' token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// ArrayExpression represents an array of expressions (for comma-separated values)
type ArrayExpression struct {
	Token    Token // The first token
	Elements []Expression
}

func (ae *ArrayExpression) expressionNode()      {}
func (ae *ArrayExpression) TokenLiteral() string { return ae.Token.Literal }

// ProgramCounterExpression represents the program counter (*) in addressing modes
type ProgramCounterExpression struct {
	Token Token // The '*' token
}

func (pc *ProgramCounterExpression) expressionNode()      {}
func (pc *ProgramCounterExpression) TokenLiteral() string { return pc.Token.Literal }
