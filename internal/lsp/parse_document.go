package lsp

import (
	log "c64.nvim/internal/log"
)

// ParseDocument parses an assembly document and returns the symbol scope and diagnostics
func ParseDocument(uri string, text string) (*Scope, []Diagnostic) {
	var program *Program
	var parserDiagnostics []Diagnostic

	// Use Context-Aware Lexer and Parser
	log.Debug("ParseDocument: Using Context-Aware Lexer and Parser")

	processorCtx := GetProcessorContext()
	if processorCtx == nil {
		log.Error("ParseDocument: ProcessorContext is nil")
		return NewRootScope(uri), []Diagnostic{{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 0},
			},
			Severity: 1, // Error
			Source:   "parser",
			Message:  "Internal error: ProcessorContext not initialized",
		}}
	}

	// Create context-aware lexer and parser
	lexer := NewContextAwareLexer(text, processorCtx)
	parser := NewContextAwareParser(lexer, processorCtx)
	program = parser.ParseProgram()
	parserDiagnostics = parser.Errors()

	// Pass 1: Build the symbol table from the AST
	scope, definitionDiagnostics := buildScopeFromAST(program, uri)

	// Pass 2: Perform semantic analysis (e.g., find symbol usages)
	analyzer := NewSemanticAnalyzer(scope, text)
	semanticDiagnostics := analyzer.Analyze(program)

	// Combine all diagnostics
	allDiagnostics := append(parserDiagnostics, definitionDiagnostics...)
	allDiagnostics = append(allDiagnostics, semanticDiagnostics...)

	return scope, allDiagnostics
}
