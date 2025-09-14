package lsp

import (
	"sort"
)

func generateDocumentSymbols(uri string) []DocumentSymbol {
	symbolStore.RLock()
	rootScope, ok := symbolStore.trees[uri]
	symbolStore.RUnlock()

	if !ok {
		return nil
	}

	return convertScopeToDocumentSymbols(rootScope)
}

func convertScopeToDocumentSymbols(scope *Scope) []DocumentSymbol {
	var symbols []DocumentSymbol

	// Convert symbols in the current scope
	for _, symbol := range scope.Symbols {
		// Skip adding namespaces here as they are handled as child scopes
		if symbol.Kind == Namespace {
			continue
		}

		docSymbol := DocumentSymbol{
			Name:   symbol.Name,
			Kind:   toDocumentSymbolKind(symbol.Kind),
			Detail: symbol.Value,
			Range: Range{
				Start: Position{Line: symbol.Position.Line, Character: 0},
				End:   Position{Line: symbol.Position.Line, Character: 80}, // Default to whole line
			},
			SelectionRange: Range{
				Start: Position{Line: symbol.Position.Line, Character: symbol.Position.Character},
				End:   Position{Line: symbol.Position.Line, Character: symbol.Position.Character + len(symbol.Name)},
			},
		}
		symbols = append(symbols, docSymbol)
	}

	// Convert child scopes (namespaces)
	for _, childScope := range scope.Children {
		namespaceSymbol := DocumentSymbol{
			Name:   childScope.Name,
			Kind:   3, // Namespace
			Range:  childScope.Range,
			SelectionRange: Range{
				Start: childScope.Range.Start,
				End:   Position{Line: childScope.Range.Start.Line, Character: childScope.Range.Start.Character + len(childScope.Name)},
			},
			Children: convertScopeToDocumentSymbols(childScope),
		}
		symbols = append(symbols, namespaceSymbol)
	}

	// Sort symbols by line number for a clean outline
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].SelectionRange.Start.Line < symbols[j].SelectionRange.Start.Line
	})

	return symbols
}

func toDocumentSymbolKind(kind SymbolKind) float64 {
	switch kind {
	case Constant:
		return 14 // Constant
	case Variable:
		return 13 // Variable
	case Label:
		return 8 // Field
	case Function:
		return 12 // Function
	case Macro:
		return 12 // Function
	case Namespace:
		return 3 // Namespace
	default:
		return 20 // Key
	}
}