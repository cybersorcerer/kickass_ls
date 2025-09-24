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
	Name       string
	Kind       SymbolKind
	Value      string   // e.g., the value of a constant
	Position   Position // The position of the definition
	Scope      *Scope   // The scope in which the symbol is defined
	Params     []string // For functions and macros
	Signature  string   // For functions and macros
	UsageCount int      // To track references
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

// Reference represents a reference to a symbol in the code
type Reference struct {
	Position Position `json:"position"`
	Uri      string   `json:"uri"`
}

// FindAllReferences finds all references to a symbol in the document
func (s *Scope) FindAllReferences(symbolName, documentText, uri string) []map[string]interface{} {
	references := []map[string]interface{}{}
	
	// Normalize the symbol name
	normalizedName := normalizeLabel(symbolName)
	
	// Split document into lines for searching
	lines := strings.Split(documentText, "\n")
	
	// Search through all lines
	for lineNum, line := range lines {
		// Find all occurrences of the symbol in this line
		references = append(references, findReferencesInLine(line, lineNum, normalizedName, uri)...)
	}
	
	return references
}

// findReferencesInLine finds all references to a symbol in a single line
func findReferencesInLine(line string, lineNum int, symbolName, uri string) []map[string]interface{} {
	references := []map[string]interface{}{}
	
	// Convert line to lowercase for case-insensitive search, but preserve original positions
	lowerLine := strings.ToLower(line)
	lowerSymbol := strings.ToLower(symbolName)
	
	// Find comment positions to exclude them from search
	commentStart := findCommentStart(line)
	
	// Find all occurrences
	searchIndex := 0
	for {
		index := strings.Index(lowerLine[searchIndex:], lowerSymbol)
		if index == -1 {
			break
		}
		
		actualIndex := searchIndex + index
		
		// Skip if this occurrence is in a comment
		if commentStart != -1 && actualIndex >= commentStart {
			searchIndex = actualIndex + 1
			continue
		}
		
		// Check if this is a complete word (not part of another identifier)
		if isCompleteWord(line, actualIndex, len(symbolName)) {
			reference := map[string]interface{}{
				"uri": uri,
				"range": map[string]interface{}{
					"start": map[string]interface{}{
						"line":      lineNum,
						"character": actualIndex,
					},
					"end": map[string]interface{}{
						"line":      lineNum,
						"character": actualIndex + len(symbolName),
					},
				},
			}
			references = append(references, reference)
		}
		
		searchIndex = actualIndex + 1
	}
	
	return references
}

// isCompleteWord checks if the found text is a complete word and not part of another identifier
func isCompleteWord(line string, startPos, length int) bool {
	endPos := startPos + length
	
	// Check character before (if exists)
	if startPos > 0 {
		charBefore := line[startPos-1]
		if isWordChar(charBefore) {
			return false
		}
	}
	
	// Check character after (if exists)
	if endPos < len(line) {
		charAfter := line[endPos]
		if isWordChar(charAfter) {
			return false
		}
	}
	
	return true
}

// isWordChar checks if a character is part of a word (identifier)
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.'
}

// normalizeLabel removes trailing colon from labels
func normalizeLabel(label string) string {
	return strings.TrimSuffix(label, ":")
}

// findCommentStart finds the position where a comment starts in a line
// Returns -1 if no comment is found
// This function handles strings properly to avoid false positives
func findCommentStart(line string) int {
	inString := false
	stringChar := byte(0)
	
	for i := 0; i < len(line); i++ {
		c := line[i]
		
		// Handle string literals
		if !inString {
			if c == '"' || c == '\'' {
				inString = true
				stringChar = c
				continue
			}
		} else {
			// We're in a string, check for end of string
			if c == stringChar {
				// Check if it's escaped (simple check)
				if i > 0 && line[i-1] != '\\' {
					inString = false
					stringChar = 0
				}
			}
			continue
		}
		
		// If not in string, check for comments
		if !inString {
			// Check for C-style comments (//)
			if c == '/' && i+1 < len(line) && line[i+1] == '/' {
				return i
			}
			
			// Check for assembly-style comments (;)
			if c == ';' {
				return i
			}
		}
	}
	
	return -1
}
