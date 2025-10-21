// internal/lsp/document_symbol.go
package lsp

import (
	"c64.nvim/internal/log"
)

// toDocumentSymbolKind converts SymbolKind to LSP DocumentSymbol kind
func toDocumentSymbolKind(kind SymbolKind) float64 {
	switch kind {
	case Constant:
		return 14 // Constant
	case Variable:
		return 13 // Variable  
	case Label:
		return 13 // Variable (labels are named memory addresses/positions)
	case Function:
		return 12 // Function
	case Macro:
		return 12 // Function (closest match for macros)
	case PseudoCommand:
		return 12 // Function (closest match for pseudocommands)
	case Namespace:
		return 3  // Namespace
	default:
		return 1 // File (fallback)
	}
}

// generateDocumentSymbols creates document symbols from the symbol tree
func generateDocumentSymbols(uri string) []DocumentSymbol {
	symbolStore.RLock()
	tree, exists := symbolStore.trees[uri]
	symbolStore.RUnlock()

	if !exists {
		log.Debug("generateDocumentSymbols: No symbol tree found for URI: %s", uri)
		return []DocumentSymbol{}
	}

	return convertScopeToDocumentSymbols(tree)
}

// convertScopeToDocumentSymbols recursively converts a scope to document symbols
func convertScopeToDocumentSymbols(scope *Scope) []DocumentSymbol {
	var symbols []DocumentSymbol

	// Add symbols from current scope
	for _, symbol := range scope.Symbols {
		docSymbol := DocumentSymbol{
			Name: symbol.Name,
			Kind: toDocumentSymbolKind(symbol.Kind),
			Range: Range{
				Start: Position{Line: symbol.Position.Line, Character: symbol.Position.Character},
				End:   Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)},
			},
			SelectionRange: Range{
				Start: Position{Line: symbol.Position.Line, Character: symbol.Position.Character},
				End:   Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)},
			},
			Children: []DocumentSymbol{},
		}

		if symbol.Value != "" {
			docSymbol.Detail = symbol.Value
		} else if symbol.Signature != "" {
			docSymbol.Detail = symbol.Signature
		}

		symbols = append(symbols, docSymbol)
	}

	// Add child scopes as symbols
	for _, child := range scope.Children {
		// Sanitize the range to ensure no negative line numbers (VSCode compatibility)
		endLine := child.Range.End.Line
		if endLine < 0 {
			endLine = 999999 // Use large number instead of -1 for EOF
		}

		childSymbol := DocumentSymbol{
			Name: child.Name,
			Kind: toDocumentSymbolKind(Namespace),
			Range: Range{
				Start: child.Range.Start,
				End:   Position{Line: endLine, Character: child.Range.End.Character},
			},
			SelectionRange: Range{
				Start: child.Range.Start,
				End:   Position{Line: child.Range.Start.Line, Character: child.Range.Start.Character + len(child.Name)},
			},
			Children: convertScopeToDocumentSymbols(child),
		}
		symbols = append(symbols, childSymbol)
	}

	return symbols
}
