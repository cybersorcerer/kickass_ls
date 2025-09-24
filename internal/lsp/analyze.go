package lsp

import (
	"fmt"
	"strings"
)

// SemanticAnalyzer performs semantic analysis on the AST, after the initial scope has been built.
// This includes tasks like resolving symbols, checking for unused symbols, etc.
type SemanticAnalyzer struct {
	scope         *Scope
	diagnostics   []Diagnostic
	documentLines []string
}

// NewSemanticAnalyzer creates a new analyzer.
func NewSemanticAnalyzer(scope *Scope, text string) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		scope:         scope,
		diagnostics:   []Diagnostic{},
		documentLines: strings.Split(text, "\n"),
	}
}

// Analyze starts the analysis of the program.
func (a *SemanticAnalyzer) Analyze(program *Program) []Diagnostic {
	if program == nil {
		return a.diagnostics
	}
	a.walkStatements(program.Statements, a.scope)

	// After walking the whole tree, check for unused symbols.
	if warnUnusedLabelsEnabled {
		a.diagnostics = append(a.diagnostics, a.checkForUnusedSymbols(a.scope)...)
	}

	return a.diagnostics
}

func (a *SemanticAnalyzer) walkStatements(statements []Statement, currentScope *Scope) {
	for _, statement := range statements {
		a.walkStatement(statement, currentScope)
	}
}

func (a *SemanticAnalyzer) walkStatement(stmt Statement, currentScope *Scope) {
	if stmt == nil {
		return
	}
	switch node := stmt.(type) {
	case *InstructionStatement:
		if node != nil && node.Operand != nil {
			a.walkExpression(node.Operand, currentScope)
		}
	case *DirectiveStatement:
		if node != nil {
			if node.Value != nil {
				a.walkExpression(node.Value, currentScope)
			}
			if node.Block != nil {
				// Find the child scope that corresponds to this block
				var newScope *Scope
				if node.Name != nil {
					newScope = currentScope.FindNamespace(node.Name.Value)
				}
				if newScope != nil {
					a.walkStatements(node.Block.Statements, newScope)
				} else {
					// Fallback to current scope if a specific child scope isn't found (should not happen for well-formed ASTs)
					a.walkStatements(node.Block.Statements, currentScope)
				}
			}
		}
		// We don't need to walk LabelStatement or others as they don't contain expressions with symbol usages.
	}
}

func (a *SemanticAnalyzer) walkExpression(expr Expression, currentScope *Scope) {
	switch node := expr.(type) {
	case *Identifier:
		// Check if the identifier is in a comment before counting it as a usage.
		lineNum := node.Token.Line - 1
		if lineNum >= 0 && lineNum < len(a.documentLines) {
			line := a.documentLines[lineNum]
			commentStart := findCommentStart(line)
			if commentStart != -1 && (node.Token.Column-1) >= commentStart {
				return // It's in a comment, so don't process it.
			}
		}

		if symbol, found := currentScope.FindSymbol(node.Value); found {
			symbol.UsageCount++
		}
	case *PrefixExpression:
		if node.Right != nil {
			a.walkExpression(node.Right, currentScope)
		}
	case *InfixExpression:
		if node.Left != nil {
			a.walkExpression(node.Left, currentScope)
		}
		if node.Right != nil {
			a.walkExpression(node.Right, currentScope)
		}
	case *GroupedExpression:
		if node.Expression != nil {
			a.walkExpression(node.Expression, currentScope)
		}
	}
}

// checkForUnusedSymbols recursively traverses the scopes and finds symbols with UsageCount == 0.
func (a *SemanticAnalyzer) checkForUnusedSymbols(scope *Scope) []Diagnostic {
	var diagnostics []Diagnostic

	for _, symbol := range scope.Symbols {
		// Only warn for certain kinds of symbols. Namespaces, for example, don't need to be explicitly used.
		switch symbol.Kind {
		case Label, Constant, Variable:
			if symbol.UsageCount == 0 {
				diagnostic := Diagnostic{
					Severity: SeverityWarning,
					Range:    Range{Start: symbol.Position, End: Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)}},
					Message:  fmt.Sprintf("Unused %s '%s'", symbol.Kind.String(), symbol.Name),
					Source:   "analyzer",
				}
				diagnostics = append(diagnostics, diagnostic)
			}
		}
	}

	for _, childScope := range scope.Children {
		diagnostics = append(diagnostics, a.checkForUnusedSymbols(childScope)...)
	}

	return diagnostics
}
