package lsp

import (
	"strings"
)

// ParseDocument analysiert den gesamten Text eines Dokuments und baut einen Symbolbaum auf.
func ParseDocument(uri string, text string) *Scope {
	root := NewRootScope(uri)
	currentScope := root
	lines := strings.Split(text, "\n")

	// Set the initial range for the root scope
	root.Range.Start.Line = 0

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Ignoriere leere Zeilen und Kommentare
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		// Scope-Management
		if strings.HasSuffix(trimmedLine, "{") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 1 && parts[0] == ".namespace" {
				namespaceName := parts[1]
				namespaceSymbol := &Symbol{
					Name:  namespaceName,
					Kind:  Namespace,
					Position: Position{
						Line:      i,
						Character: strings.Index(line, namespaceName),
					},
					Scope: currentScope,
				}
				currentScope.AddSymbol(namespaceSymbol)

				newScope := &Scope{
					Name:    namespaceName,
					Symbols: make(map[string]*Symbol),
					Children: make([]*Scope, 0),
					Uri:     uri,
					Parent:  currentScope, // Set parent explicitly
					Range:   Range{Start: Position{Line: i, Character: 0}}, // Set start of new scope
				}
				currentScope.AddChildScope(newScope)
				currentScope = newScope
			}
			continue
		}

		if trimmedLine == "}" {
			if currentScope.Parent != nil {
				currentScope.Range.End.Line = i // Set end of current scope
				currentScope = currentScope.Parent
			}
			continue
		}

		parts := strings.Fields(trimmedLine)
		if len(parts) == 0 {
			continue
		}

		// --- Direktiven- und Symbolerkennung ---
		firstWord := parts[0]
		lowerFirstWord := strings.ToLower(firstWord)

		// .const, .var, .label
		if (lowerFirstWord == ".const" || lowerFirstWord == ".var" || lowerFirstWord == ".label") && len(parts) >= 2 {
			var symbolName, symbolValue string
			var kind SymbolKind

			if lowerFirstWord == ".label" {
				kind = Label
				symbolName = strings.TrimSuffix(parts[1], ":")
			} else {
				if len(parts) >= 4 && parts[2] == "=" {
					symbolName = parts[1]
					symbolValue = strings.Join(parts[3:], " ")
					if lowerFirstWord == ".const" {
						kind = Constant
					} else {
						kind = Variable
					}
				} else {
					continue
				}
			}

			symbol := &Symbol{
				Name:  symbolName,
				Kind:  kind,
				Value: symbolValue,
				Position: Position{
					Line:      i,
					Character: strings.Index(line, symbolName),
				},
				Scope: currentScope,
			}
			currentScope.AddSymbol(symbol)
			continue
		}

		// Label-Definition
		if strings.HasSuffix(firstWord, ":") {
			labelName := strings.TrimSuffix(firstWord, ":")
			symbol := &Symbol{
				Name: labelName,
				Kind: Label,
				Position: Position{
					Line:      i,
					Character: strings.Index(line, labelName),
				},
				Scope: currentScope,
			}
			currentScope.AddSymbol(symbol)
			continue
		}

	}

	// Set the end line for the root scope (last line of the document)
	root.Range.End.Line = len(lines) - 1

	return root
}
