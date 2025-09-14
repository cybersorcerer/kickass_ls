package lsp

import (
	"c64.nvim/internal/log"
	"regexp"
	"sort"
	"strings"
)

var (
	tokenTypes = map[string]uint32{
		"keyword":   0,
		"variable":  1,
		"function":  2,
		"macro":     3,
		"number":    4,
		"comment":   5,
		"string":    6,
		"operator":  7,
	}
	tokenModifiers = map[string]uint32{
		"declaration": 1 << 0,
		"readonly":    1 << 1,
	}
	KickAssemblerDirectives = map[string]bool{
		".const":        true,
		".var":          true,
		".word":         true,
		".byte":         true,
		".namespace":    true,
		".function":     true,
		".macro":        true,
		".label":        true,
		".pseudocommand": true,
		".if":           true,
		".for":          true,
		".while":        true,
		".return":       true,
		"#import":       true,
		"#include":      true,
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

func parseLineForTokens(line string, lineNum uint32, symbolTree *Scope) []semanticToken {
	var tokens []semanticToken
	log.Debug("Parsing line %d: '%s'", lineNum, line)

	// 1. Comments
	commentIdx := strings.Index(line, ";")
	if commentIdx == -1 {
		commentIdx = strings.Index(line, "//")
	}

	codePart := line
	if commentIdx != -1 {
		commentText := line[commentIdx:]
		tokens = append(tokens, semanticToken{lineNum, uint32(commentIdx), uint32(len(commentText)), tokenTypes["comment"], 0})
		codePart = line[:commentIdx]
	}

	// 2. String Literals
	re := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	stringLiterals := re.FindAllStringIndex(codePart, -1)

	nonStringCodePart := []rune(codePart)

	for _, match := range stringLiterals {
		start, end := match[0], match[1]
		log.Debug("Found String: '%s'", codePart[start:end])
		tokens = append(tokens, semanticToken{lineNum, uint32(start), uint32(end - start), tokenTypes["string"], 0})
		for i := start; i < end; i++ {
			nonStringCodePart[i] = ' '
		}
	}

	codePart = string(nonStringCodePart)

	parts := strings.Fields(codePart)

	for _, part := range parts {
		startChar := strings.Index(codePart, part)
		if startChar == -1 {
			continue
		}

		lowerPart := strings.ToLower(part)
		upperPart := strings.ToUpper(part)

		if _, isDirective := KickAssemblerDirectives[lowerPart]; isDirective {
			log.Debug("Found Directive: '%s'", part)
			tokens = append(tokens, semanticToken{lineNum, uint32(startChar), uint32(len(part)), tokenTypes["macro"], 0})
		} else if _, isOpcode := allOpcodes[upperPart]; isOpcode {
			log.Debug("Found Opcode: '%s'", part)
			tokens = append(tokens, semanticToken{lineNum, uint32(startChar), uint32(len(part)), tokenTypes["keyword"], 0})
		} else if strings.HasPrefix(part, "#") || strings.HasPrefix(part, "$") || strings.HasPrefix(part, "%") || (part[0] >= '0' && part[0] <= '9') {
			log.Debug("Found Number: '%s'", part)
			tokens = append(tokens, semanticToken{lineNum, uint32(startChar), uint32(len(part)), tokenTypes["number"], 0})
		} else if symbol, found := symbolTree.FindSymbol(normalizeLabel(part)); found {
			log.Debug("Found Symbol: '%s' (Kind: %s)", part, symbol.Kind)
			var mod uint32 = 0
			if symbol.Kind == Constant {
				mod = tokenModifiers["readonly"]
			}

			tokenType := tokenTypes["variable"]
			if symbol.Kind == Function {
				tokenType = tokenTypes["function"]
			} else if symbol.Kind == Macro {
				tokenType = tokenTypes["macro"]
			} else if strings.HasSuffix(part, ":") {
				mod |= tokenModifiers["declaration"]
			}

			tokens = append(tokens, semanticToken{lineNum, uint32(startChar), uint32(len(part)), tokenType, mod})
		} else if len(part) == 1 && strings.ContainsAny(part, ":=+-*/(),<>") {
			log.Debug("Found Operator: '%s'", part)
			tokens = append(tokens, semanticToken{lineNum, uint32(startChar), uint32(len(part)), tokenTypes["operator"], 0})
		} else {
			log.Debug("Symbol not found: '%s'", part)
		}
	}

	return tokens
}

func generateSemanticTokens(uri string, text string) []uint32 {
	log.Debug("Generating semantic tokens for %s", uri)

	// Populate allOpcodes map
	for _, m := range mnemonics {
		allOpcodes[strings.ToUpper(m.Mnemonic)] = true
	}

	// We need the symbol tree to identify symbols
	symbolTree := ParseDocument(uri, text)

	allTokens := []semanticToken{}

	lines := strings.Split(text, "\n")

	for i, line := range lines {
		lineTokens := parseLineForTokens(line, uint32(i), symbolTree)
		allTokens = append(allTokens, lineTokens...)
	}

	// Sort tokens by start character index, as regex can be unordered
	sort.SliceStable(allTokens, func(i, j int) bool {
		if allTokens[i].line != allTokens[j].line {
			return allTokens[i].line < allTokens[j].line
		}
		return allTokens[i].startChar < allTokens[j].startChar
	})

	// Build the response

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
