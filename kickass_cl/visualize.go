package main

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes for terminal output
const (
	ColorReset     = "\033[0m"
	ColorRed       = "\033[31m"
	ColorGreen     = "\033[32m"
	ColorYellow    = "\033[33m"
	ColorBlue      = "\033[34m"
	ColorMagenta   = "\033[35m"
	ColorCyan      = "\033[36m"
	ColorWhite     = "\033[37m"
	ColorOrange    = "\033[38;5;208m"
	ColorPurple    = "\033[38;5;93m"
	ColorLightBlue = "\033[38;5;117m"
	ColorGray      = "\033[90m"
	ColorBold      = "\033[1m"
)

// Token type to color mapping (matching your LSP server's semantic token types)
var tokenTypeColors = map[int]string{
	0:  ColorCyan,      // keyword
	1:  ColorCyan,      // variable
	2:  ColorYellow,    // function
	3:  ColorYellow,    // macro
	4:  ColorYellow,    // pseudocommand
	5:  ColorOrange,    // number
	6:  ColorGray,      // comment
	7:  ColorGreen,     // string
	8:  ColorWhite,     // operator
	9:  ColorMagenta,   // mnemonic
	10: ColorPurple,    // directive
	11: ColorLightBlue, // preprocessor
	12: ColorBlue,      // label
}

var tokenTypeNames = []string{
	"keyword",       // 0
	"variable",      // 1
	"function",      // 2
	"macro",         // 3
	"pseudocommand", // 4
	"number",        // 5
	"comment",       // 6
	"string",        // 7
	"operator",      // 8
	"mnemonic",      // 9
	"directive",     // 10
	"preprocessor",  // 11
	"label",         // 12
}

// DecodedToken represents a decoded semantic token with position and type
type DecodedToken struct {
	Line       int
	StartChar  int
	Length     int
	TokenType  int
	Modifiers  int
	TypeName   string
	ColorCode  string
}

// DecodeSemanticTokens decodes the relative-encoded semantic tokens into absolute positions
func DecodeSemanticTokens(data []int) []DecodedToken {
	if len(data)%5 != 0 {
		fmt.Fprintf(os.Stderr, "Warning: semantic tokens data length %d is not multiple of 5\n", len(data))
		return nil
	}

	tokens := make([]DecodedToken, 0, len(data)/5)
	currentLine := 0
	currentChar := 0

	for i := 0; i < len(data); i += 5 {
		deltaLine := data[i]
		deltaChar := data[i+1]
		length := data[i+2]
		tokenType := data[i+3]
		modifiers := data[i+4]

		// Update absolute position
		if deltaLine > 0 {
			currentLine += deltaLine
			currentChar = deltaChar
		} else {
			currentChar += deltaChar
		}

		typeName := "unknown"
		if tokenType >= 0 && tokenType < len(tokenTypeNames) {
			typeName = tokenTypeNames[tokenType]
		}

		colorCode := tokenTypeColors[tokenType]
		if colorCode == "" {
			colorCode = ColorWhite
		}

		tokens = append(tokens, DecodedToken{
			Line:      currentLine,
			StartChar: currentChar,
			Length:    length,
			TokenType: tokenType,
			Modifiers: modifiers,
			TypeName:  typeName,
			ColorCode: colorCode,
		})
	}

	return tokens
}

// VisualizeSemanticTokens prints the file content with colored semantic tokens
func VisualizeSemanticTokens(content string, tokens []DecodedToken) {
	lines := strings.Split(content, "\n")

	// Group tokens by line for efficient rendering
	tokensByLine := make(map[int][]DecodedToken)
	for _, token := range tokens {
		tokensByLine[token.Line] = append(tokensByLine[token.Line], token)
	}

	// Print each line with colored tokens
	for lineNum, lineContent := range lines {
		lineTokens := tokensByLine[lineNum]

		// Sort tokens by start position
		// (they should already be sorted, but just to be safe)
		for i := 0; i < len(lineTokens)-1; i++ {
			for j := i + 1; j < len(lineTokens); j++ {
				if lineTokens[j].StartChar < lineTokens[i].StartChar {
					lineTokens[i], lineTokens[j] = lineTokens[j], lineTokens[i]
				}
			}
		}

		// Print line number
		fmt.Printf("%s%4d:%s ", ColorGray, lineNum+1, ColorReset)

		// Print line with colored tokens
		if len(lineTokens) == 0 {
			// No tokens on this line, print as-is
			fmt.Println(lineContent)
		} else {
			// Print with colors
			pos := 0
			for _, token := range lineTokens {
				// Print text before token (uncolored)
				if token.StartChar > pos {
					fmt.Print(lineContent[pos:token.StartChar])
				}

				// Print token with color
				endPos := token.StartChar + token.Length
				if endPos > len(lineContent) {
					endPos = len(lineContent)
				}
				fmt.Printf("%s%s%s", token.ColorCode, lineContent[token.StartChar:endPos], ColorReset)
				pos = endPos
			}

			// Print remaining text after last token
			if pos < len(lineContent) {
				fmt.Println(lineContent[pos:])
			} else {
				fmt.Println()
			}
		}
	}
}

// PrintSemanticTokensSummary prints a summary of semantic tokens
func PrintSemanticTokensSummary(tokens []DecodedToken) {
	fmt.Printf("\n%s=== Semantic Tokens Summary ===%s\n", ColorBold, ColorReset)
	fmt.Printf("Total tokens: %d\n\n", len(tokens))

	// Count by type
	typeCounts := make(map[string]int)
	for _, token := range tokens {
		typeCounts[token.TypeName]++
	}

	fmt.Println("Token counts by type:")
	for i, typeName := range tokenTypeNames {
		if count, ok := typeCounts[typeName]; ok && count > 0 {
			color := tokenTypeColors[i]
			fmt.Printf("  %s%-15s%s: %d\n", color, typeName, ColorReset, count)
		}
	}
}

// PrintTokenDetails prints detailed information about tokens at a specific line
func PrintTokenDetails(tokens []DecodedToken, line int, content string) {
	lines := strings.Split(content, "\n")
	if line < 0 || line >= len(lines) {
		fmt.Printf("Line %d out of range\n", line)
		return
	}

	lineContent := lines[line]
	lineTokens := make([]DecodedToken, 0)
	for _, token := range tokens {
		if token.Line == line {
			lineTokens = append(lineTokens, token)
		}
	}

	fmt.Printf("\n%s=== Line %d Details ===%s\n", ColorBold, line+1, ColorReset)
	fmt.Printf("Content: %s\n\n", lineContent)

	if len(lineTokens) == 0 {
		fmt.Println("No tokens on this line")
		return
	}

	fmt.Printf("Tokens (%d):\n", len(lineTokens))
	for i, token := range lineTokens {
		endPos := token.StartChar + token.Length
		if endPos > len(lineContent) {
			endPos = len(lineContent)
		}
		text := lineContent[token.StartChar:endPos]

		fmt.Printf("%3d. %sChar %2d-%2d%s [Type=%s%-12s%s] %s\"%s\"%s\n",
			i+1,
			ColorGray, token.StartChar, endPos, ColorReset,
			token.ColorCode, token.TypeName, ColorReset,
			token.ColorCode, text, ColorReset)
	}

	// Print visual indicator
	fmt.Println("\nVisual:")
	fmt.Printf("     %s\n", lineContent)
	fmt.Print("     ")
	for pos := 0; pos < len(lineContent); pos++ {
		found := false
		for _, token := range lineTokens {
			if pos >= token.StartChar && pos < token.StartChar+token.Length {
				fmt.Print("^")
				found = true
				break
			}
		}
		if !found {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}

// CompareSemanticTokens compares expected vs actual tokens and prints differences
func CompareSemanticTokens(expected, actual []DecodedToken, content string) bool {
	if len(expected) != len(actual) {
		fmt.Printf("%sToken count mismatch: expected %d, got %d%s\n",
			ColorRed, len(expected), len(actual), ColorReset)
		return false
	}

	allMatch := true
	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		act := actual[i]

		if exp.Line != act.Line || exp.StartChar != act.StartChar ||
			exp.Length != act.Length || exp.TokenType != act.TokenType {

			allMatch = false
			lines := strings.Split(content, "\n")
			lineContent := ""
			if exp.Line < len(lines) {
				lineContent = lines[exp.Line]
			}

			fmt.Printf("%s❌ Token %d mismatch:%s\n", ColorRed, i+1, ColorReset)
			fmt.Printf("   Line %d: %s\n", exp.Line+1, lineContent)
			fmt.Printf("   Expected: Line=%d, Char=%d, Length=%d, Type=%s%s%s\n",
				exp.Line, exp.StartChar, exp.Length,
				exp.ColorCode, exp.TypeName, ColorReset)
			fmt.Printf("   Actual:   Line=%d, Char=%d, Length=%d, Type=%s%s%s\n",
				act.Line, act.StartChar, act.Length,
				act.ColorCode, act.TypeName, ColorReset)
		}
	}

	if allMatch {
		fmt.Printf("%s✅ All tokens match!%s\n", ColorGreen, ColorReset)
	}

	return allMatch
}
