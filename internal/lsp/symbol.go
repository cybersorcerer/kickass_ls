package lsp

import (
	"c64.nvim/internal/log"
	"fmt"
	"strings"
)

// SymbolKind definiert die Art eines Symbols (Konstante, Variable, etc.).
type SymbolKind int

const (
	UnknownSymbol SymbolKind = iota
	Constant
	Variable
	Label
	Function
	Macro
	Namespace
)

// String gibt eine lesbare Repräsentation des SymbolKind zurück.
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

// Position repräsentiert eine Stelle im Code (Zeile und Spalte).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Symbol repräsentiert ein einzelnes Symbol im Code.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	Value    string   // z.B. der Wert einer Konstante
	Position Position // Die Position der Definition
	Scope    *Scope   // Der Geltungsbereich, in dem das Symbol definiert ist
}

// Scope repräsentiert einen Geltungsbereich (z.B. eine Datei, ein Namespace oder eine Funktion).
type Scope struct {
	Name        string
	Parent      *Scope
	Children    []*Scope
	Symbols     map[string]*Symbol
	Range       Range // Der Bereich, den dieser Scope im Dokument abdeckt
	Uri         string
}

// Range repräsentiert einen Bereich im Code (Anfang und Ende).
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}


// NewRootScope erstellt einen neuen Wurzel-Scope für ein Dokument.
func NewRootScope(uri string) *Scope {
	return &Scope{
		Name:     "root",
		Symbols:  make(map[string]*Symbol),
		Children: make([]*Scope, 0),
		Uri:      uri,
	}
}

// AddSymbol fügt ein Symbol zum Scope hinzu.
func (s *Scope) AddSymbol(symbol *Symbol) error {
	if _, exists := s.Symbols[symbol.Name]; exists {
		return fmt.Errorf("symbol '%s' already defined in this scope", symbol.Name)
	}
	log.Debug("Adding symbol '%s' to scope '%s'", symbol.Name, s.Name)
	s.Symbols[symbol.Name] = symbol
	return nil
}

// AddChildScope fügt einen untergeordneten Scope hinzu.
func (s *Scope) AddChildScope(child *Scope) {
	child.Parent = s
	s.Children = append(s.Children, child)
}

// FindSymbol sucht nach einem Symbol, beginnend im aktuellen Scope und dann rekursiv in den Eltern-Scopes.
// Behandelt auch qualifizierte Namen (z.B. namespace.symbol).
func (s *Scope) FindSymbol(name string) (*Symbol, bool) {
	parts := strings.Split(name, ".")

	if len(parts) > 1 {
		// Qualifizierter Name
		namespaceName := parts[0]
		symbolName := parts[1]

		// Finde den Namespace-Scope
		if nsScope := s.FindNamespace(normalizeLabel(namespaceName)); nsScope != nil {
			// Suche das Symbol innerhalb des Namespace-Scopes
			if symbol, ok := nsScope.Symbols[normalizeLabel(symbolName)]; ok {
				return symbol, true
			}
		}
	} else {
		// Nicht qualifizierter Name
		if symbol, ok := s.Symbols[normalizeLabel(name)]; ok {
			return symbol, true
		}
		if s.Parent != nil {
			return s.Parent.FindSymbol(name)
		}
	}

	return nil, false
}


// FindNamespace sucht nach einem Namespace-Scope mit dem gegebenen Namen.
func (s *Scope) FindNamespace(name string) *Scope {
	for _, child := range s.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// FindAllVisibleSymbols sammelt alle Symbole, die von einem bestimmten Punkt im Code aus sichtbar sind.
func (s *Scope) FindAllVisibleSymbols(lineNumber int) []*Symbol {
	var visibleSymbols []*Symbol

	// Finde den innersten Scope, der die aktuelle Zeilennummer umschließt
	currentScope := s.findInnermostScope(lineNumber)

	// Sammle Symbole vom aktuellen Scope nach oben bis zum Root
	for scope := currentScope; scope != nil; scope = scope.Parent {
		for _, symbol := range scope.Symbols {
			visibleSymbols = append(visibleSymbols, symbol)
		}
	}

	return visibleSymbols
}

// findInnermostScope findet den spezifischsten Scope für eine gegebene Zeilennummer.
func (s *Scope) findInnermostScope(lineNumber int) *Scope {
	for _, child := range s.Children {
		if lineNumber >= child.Range.Start.Line && lineNumber <= child.Range.End.Line {
			return child.findInnermostScope(lineNumber) // Rekursiv im passenden Kind-Scope suchen
		}
	}
	return s // Kein passenderer Kind-Scope gefunden, also ist dies der innerste
}