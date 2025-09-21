package lsp

import (
	"sort"
	"strings"

	"c64.nvim/internal/log"
)

var (
	tokenTypes = map[string]uint32{
		"keyword":  0,
		"variable": 1,
		"function": 2,
		"macro":    3,
		"number":   4,
		"comment":  5,
		"string":   6,
		"operator": 7,
	}
	tokenModifiers = map[string]uint32{
		"declaration": 1 << 0,
		"readonly":    1 << 1,
	}
	KickAssemblerDirectives = map[string]bool{
		".const":         true,
		".var":           true,
		".word":          true,
		".byte":          true,
		".namespace":     true,
		".function":      true,
		".macro":         true,
		".label":         true,
		".pseudocommand": true,
		".if":            true,
		".for":           true,
		".while":         true,
		".return":        true,
		"#import":        true,
		"#include":       true,
	}
	allOpcodes = make(map[string]bool)
)

type semanticToken struct {
	line      uint32
	startChar uint32
	length    uint32
	tokenType uint32
	tokenMods uint32
}

func generateSemanticTokens(uri string, text string) []uint32 {
	log.Debug("Generating semantic tokens for %s", uri)

	// Populate allOpcodes map
	for _, m := range mnemonics {
		allOpcodes[strings.ToUpper(m.Mnemonic)] = true
	}

	// Use the new AST-based parser.
	l := NewLexer(text)
	p := NewParser(l)
	program := p.ParseProgram()

	allTokens := buildTokensFromAST(program)

	// Sort tokens by start character index, as regex can be unordered.
	sort.SliceStable(allTokens, func(i, j int) bool {
		if allTokens[i].line != allTokens[j].line {
			return allTokens[i].line < allTokens[j].line
		}
		return allTokens[i].startChar < allTokens[j].startChar
	})

	// Build the response.

	data := make([]uint32, 0, len(allTokens)*5)

	var prevLine, prevChar uint32
	for _, token := range allTokens {
		deltaLine := token.line - prevLine
		deltaChar := token.startChar
		if deltaLine == 0 {
			deltaChar = token.startChar - prevChar
		}

		data = append(data, deltaLine, deltaChar, token.length, token.tokenType, token.tokenMods)

		prevLine = token.line
		prevChar = token.startChar
	}

	return data
}

func buildTokensFromAST(node Node) []semanticToken {
	tokens := []semanticToken{}

	// Add a nil check to prevent panics on incomplete AST nodes from the parser.
	if node == nil {
		return tokens
	}

	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Statements {
			// Defensive check: Ensure statement from the program is not nil before processing.
			if stmt == nil {
				continue
			}
			tokens = append(tokens, buildTokensFromAST(stmt)...)
		}
	case *InstructionStatement:
		tok := n.Token
		tokens = append(tokens, semanticToken{
			line:      uint32(tok.Line - 1),
			startChar: uint32(tok.Column - 1),
			length:    uint32(len(tok.Literal)),
			tokenType: tokenTypes["keyword"],
		})
		if n.Operand != nil {
			tokens = append(tokens, buildTokensFromAST(n.Operand)...)
		}
	case *LabelStatement:
		tok := n.Token
		tokens = append(tokens, semanticToken{
			line:      uint32(tok.Line - 1),
			startChar: uint32(tok.Column - 1),
			length:    uint32(len(tok.Literal)),
			tokenType: tokenTypes["function"], // Use a distinct color for labels
			tokenMods: tokenModifiers["declaration"],
		})
	case *DirectiveStatement:
		tok := n.Token
		tokens = append(tokens, semanticToken{
			line:      uint32(tok.Line - 1),
			startChar: uint32(tok.Column - 1),
			length:    uint32(len(tok.Literal)),
			tokenType: tokenTypes["macro"],
		})
		if n.Name != nil {
			tokens = append(tokens, buildTokensFromAST(n.Name)...)
		}
		if n.Value != nil {
			tokens = append(tokens, buildTokensFromAST(n.Value)...)
		}
	case *Identifier:
		tok := n.Token
		tokens = append(tokens, semanticToken{
			line:      uint32(tok.Line - 1),
			startChar: uint32(tok.Column - 1),
			length:    uint32(len(tok.Literal)),
			tokenType: tokenTypes["variable"],
		})
	case *IntegerLiteral:
		tok := n.Token
		tokens = append(tokens, semanticToken{
			line:      uint32(tok.Line - 1),
			startChar: uint32(tok.Column - 1),
			length:    uint32(len(tok.Literal)),
			tokenType: tokenTypes["number"],
		})
	}

	return tokens
}
