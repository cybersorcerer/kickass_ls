package lsp

import (
	"fmt"
	"strings"

	"c64.nvim/internal/log"
)

// SymbolKind defines the type of a symbol (constant, variable, etc.).
type SymbolKind int

const (
	UnknownSymbol SymbolKind = iota
	Constant
	Variable // Note: In Kick Assembler, .var can be reassigned.
	Label
	Function
	Macro
	Namespace
)

// String returns a human-readable representation of the SymbolKind.
func (sk SymbolKind) String() string {
	switch sk {
	case Constant:
		return "constant"
	case Variable:
		return "variable"
	case Label:
		return "label"
	case Function:
		return "function"
	case Macro:
		return "macro"
	case Namespace:
		return "namespace"
	default:
		return "unknown"
	}
}

// Position represents a location in the code (line and character).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Symbol represents a single symbol in the code.
type Symbol struct {
	Name      string
	Kind      SymbolKind
	Value     string   // e.g., the value of a constant
	Position  Position // The position of the definition
	Scope     *Scope   // The scope in which the symbol is defined
	Params    []string // For functions and macros
	Signature string   // For functions and macros
}

// Scope represents a scope (e.g., a file, a namespace, or a function).
type Scope struct {
	Name     string
	Parent   *Scope
	Children []*Scope
	Symbols  map[string]*Symbol
	Range    Range // The range this scope covers in the document
	Uri      string
}

// Range represents a range in the code (start and end).
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// NewRootScope creates a new root scope for a document.
func NewRootScope(uri string) *Scope {
	return &Scope{
		Name:     "root",
		Symbols:  make(map[string]*Symbol),
		Children: make([]*Scope, 0),
		Uri:      uri,
	}
}

// AddSymbol adds a symbol to the scope.
func (s *Scope) AddSymbol(symbol *Symbol) error {
	if _, exists := s.Symbols[symbol.Name]; exists {
		return fmt.Errorf("symbol '%s' already defined in this scope", symbol.Name)
	}
	log.Debug("Adding symbol '%s' to scope '%s'", symbol.Name, s.Name)
	s.Symbols[symbol.Name] = symbol
	return nil
}

// AddChildScope adds a child scope.
func (s *Scope) AddChildScope(child *Scope) {
	child.Parent = s
	s.Children = append(s.Children, child)
}

// FindSymbol searches for a symbol, starting in the current scope and then recursively in parent scopes.
// It also handles qualified names (e.g., namespace.symbol).
func (s *Scope) FindSymbol(name string) (*Symbol, bool) {
	parts := strings.Split(name, ".")

	if len(parts) > 1 {
		// Qualified name
		namespaceName := parts[0]
		symbolName := parts[1]

		// Find the namespace scope
		if nsScope := s.FindNamespace(normalizeLabel(namespaceName)); nsScope != nil {
			// Search for the symbol within the namespace scope
			if symbol, ok := nsScope.Symbols[normalizeLabel(symbolName)]; ok {
				return symbol, true
			}
		}
	} else {
		// Unqualified name
		if symbol, ok := s.Symbols[normalizeLabel(name)]; ok {
			return symbol, true
		}
		if s.Parent != nil {
			return s.Parent.FindSymbol(name)
		}
	}

	return nil, false
}

// FindNamespace searches for a namespace scope with the given name.
func (s *Scope) FindNamespace(name string) *Scope {
	for _, child := range s.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// FindAllVisibleSymbols collects all symbols that are visible from a specific point in the code.
func (s *Scope) FindAllVisibleSymbols(lineNumber int) []*Symbol {
	var visibleSymbols []*Symbol

	// Find the innermost scope that encloses the current line number
	currentScope := s.findInnermostScope(lineNumber)

	// Collect symbols from the current scope up to the root
	for scope := currentScope; scope != nil; scope = scope.Parent {
		for _, symbol := range scope.Symbols {
			visibleSymbols = append(visibleSymbols, symbol)
		}
	}

	return visibleSymbols
}

// findInnermostScope finds the most specific scope for a given line number.
func (s *Scope) findInnermostScope(lineNumber int) *Scope {
	for _, child := range s.Children {
		// Defensive check: Ensure the child scope has a valid range before checking containment.
		// A zero-value End.Line indicates an incompletely parsed scope.
		if child.Range.End.Line == 0 && child.Range.Start.Line == 0 {
			continue
		}
		if lineNumber >= child.Range.Start.Line && lineNumber <= child.Range.End.Line {
			return child.findInnermostScope(lineNumber) // Recurse into the matching child scope
		}
	}
	return s // No more specific child scope found, so this is the innermost one
}
