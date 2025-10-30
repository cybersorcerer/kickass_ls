package lsp

import (
	"regexp"
	"strings"

	log "c64.nvim/internal/log"
)

// FormattingConfig holds configuration for document formatting
type FormattingConfig struct {
	Enabled           bool // Master switch
	IndentSize        int  // Spaces per indent level (default: 4)
	UseSpaces         bool // true = spaces, false = tabs
	AlignComments     bool // Align end-of-line comments
	AlignInstructions bool // Align instructions in columns
	LabelColumn       int  // Column for labels (0 = start of line)
	InstructionColumn int  // Column for instructions (default: 4)
	OperandColumn     int  // Column for operands (0 = auto-calculate)
	CommentColumn     int  // Column for comments (0 = auto-calculate)
}

// DefaultFormattingConfig returns the default formatting configuration
func DefaultFormattingConfig() FormattingConfig {
	return FormattingConfig{
		Enabled:           true,
		IndentSize:        4,
		UseSpaces:         true,
		AlignComments:     false,
		AlignInstructions: false,
		LabelColumn:       0,
		InstructionColumn: 4,
		OperandColumn:     0,
		CommentColumn:     0,
	}
}

// FormattedLine represents a parsed line with its components
type FormattedLine struct {
	OriginalLine  string // Original line text
	IndentLevel   int    // Nesting level (0, 1, 2, ...)
	Label         string // "start:", "!loop+:", etc.
	Instruction   string // "lda", ".macro", ".function", etc.
	Operands      string // "#$01", "param", "(ptr),y", etc.
	Comment       string // "// comment" or "; comment"
	IsBlank       bool   // Empty line
	IsCommentOnly bool   // Line with only comment
	IsDirective   bool   // .macro, .function, .namespace, etc.
	BlockStart    bool   // Line ends with '{'
	BlockEnd      bool   // Line starts with '}'
}

// Formatter handles document formatting
type Formatter struct {
	config FormattingConfig
	lines  []FormattedLine
}

// Regular expressions for parsing lines
var (
	// Label patterns
	labelRegex = regexp.MustCompile(`^(\s*)([a-zA-Z_][a-zA-Z0-9_]*|\![a-zA-Z_][a-zA-Z0-9_]*[\+\-]?)\s*:`)

	// Comment patterns
	lineCommentRegex  = regexp.MustCompile(`//.*$`)
	blockCommentRegex = regexp.MustCompile(`/\*.*?\*/`)
	semicolonCommentRegex = regexp.MustCompile(`;.*$`)

	// Block-starting directives
	blockDirectivesRegex = regexp.MustCompile(`^\s*\.(macro|function|namespace|pseudocommand|for|while|if)\b`)

	// Directive patterns (any directive starting with .)
	directiveRegex = regexp.MustCompile(`^\s*\.([a-zA-Z_][a-zA-Z0-9_]*)`)
)

// FormatDocument formats an entire document
func FormatDocument(text string, config FormattingConfig) (string, error) {
	if !config.Enabled {
		return text, nil
	}

	formatter := &Formatter{
		config: config,
		lines:  make([]FormattedLine, 0),
	}

	// Phase 1: Parse lines into FormattedLine structures
	formatter.phase1ParseLines(text)

	// Phase 2: Determine indent levels
	formatter.phase2DetermineIndent()

	// Phase 3: (Optional) Align columns
	if config.AlignInstructions || config.AlignComments {
		formatter.phase3AlignColumns()
	}

	// Phase 4: Reconstruct formatted text
	return formatter.phase4Reconstruct(), nil
}

// FormatRange formats a range of lines in a document
func FormatRange(text string, startLine, endLine int, config FormattingConfig) (string, error) {
	if !config.Enabled {
		return text, nil
	}

	lines := strings.Split(text, "\n")
	if startLine < 0 || endLine >= len(lines) || startLine > endLine {
		return text, nil
	}

	// Extract the range
	rangeText := strings.Join(lines[startLine:endLine+1], "\n")

	// Format the range
	formattedRange, err := FormatDocument(rangeText, config)
	if err != nil {
		return text, err
	}

	// Reconstruct the full document
	result := make([]string, 0, len(lines))
	result = append(result, lines[:startLine]...)
	result = append(result, strings.Split(formattedRange, "\n")...)
	result = append(result, lines[endLine+1:]...)

	return strings.Join(result, "\n"), nil
}

// phase1ParseLines parses the input text into FormattedLine structures
func (f *Formatter) phase1ParseLines(text string) {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		formatted := FormattedLine{
			OriginalLine: line,
		}

		// Check if blank line
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			formatted.IsBlank = true
			f.lines = append(f.lines, formatted)
			continue
		}

		// Check for block end (line starting with })
		if strings.HasPrefix(trimmed, "}") {
			formatted.BlockEnd = true
			// Rest of line after }
			rest := strings.TrimSpace(trimmed[1:])
			if rest != "" {
				formatted.Comment = rest
			}
			f.lines = append(f.lines, formatted)
			continue
		}

		// Extract comment (if any) - we'll do this first to avoid confusing it with other elements
		workingLine := line
		comment := ""

		// Check for line comment (//)
		if idx := strings.Index(workingLine, "//"); idx >= 0 {
			comment = workingLine[idx:]
			workingLine = workingLine[:idx]
		} else if idx := strings.Index(workingLine, ";"); idx >= 0 {
			// Check for semicolon comment
			comment = workingLine[idx:]
			workingLine = workingLine[:idx]
		}

		formatted.Comment = strings.TrimSpace(comment)
		workingLine = strings.TrimSpace(workingLine)

		// Check if comment-only line
		if workingLine == "" && formatted.Comment != "" {
			formatted.IsCommentOnly = true
			f.lines = append(f.lines, formatted)
			continue
		}

		// Check for label at start of line
		if matches := labelRegex.FindStringSubmatch(workingLine); matches != nil {
			formatted.Label = matches[2] + ":" // Include the colon
			workingLine = strings.TrimSpace(workingLine[len(matches[0]):])
		}

		// Check if this is a directive
		if strings.HasPrefix(workingLine, ".") {
			formatted.IsDirective = true
		}

		// Check for block start (line ending with {)
		if strings.HasSuffix(workingLine, "{") {
			formatted.BlockStart = true
			workingLine = strings.TrimSpace(workingLine[:len(workingLine)-1])
		}

		// Split remaining into instruction and operands
		if workingLine != "" {
			parts := strings.Fields(workingLine)
			if len(parts) > 0 {
				formatted.Instruction = parts[0]
				if len(parts) > 1 {
					// Join the rest as operands (preserve spaces in operands)
					// Find where the instruction ends in the original working line
					instrEnd := strings.Index(workingLine, parts[0]) + len(parts[0])
					formatted.Operands = strings.TrimSpace(workingLine[instrEnd:])
				}
			}
		}

		f.lines = append(f.lines, formatted)
	}

	log.Debug("Formatter Phase 1: Parsed %d lines", len(f.lines))
}

// phase2DetermineIndent calculates indent levels for each line
func (f *Formatter) phase2DetermineIndent() {
	currentLevel := 0

	for i := range f.lines {
		line := &f.lines[i]

		// Block end decreases indent level BEFORE applying to this line
		if line.BlockEnd {
			currentLevel--
			if currentLevel < 0 {
				currentLevel = 0
			}
		}

		// Apply current indent level
		line.IndentLevel = currentLevel

		// Labels always at level 0 (unless inside a block)
		// This is a design decision - we can make it configurable
		if line.Label != "" && !line.IsDirective && currentLevel == 0 {
			line.IndentLevel = 0
		}

		// Block start increases indent level AFTER applying to this line
		if line.BlockStart {
			currentLevel++
		}
	}

	log.Debug("Formatter Phase 2: Determined indent levels")
}

// phase3AlignColumns aligns instructions and comments in columns
func (f *Formatter) phase3AlignColumns() {
	// Calculate maximum widths for alignment
	maxLabelWidth := 0
	maxInstructionWidth := 0
	maxOperandWidth := 0

	for _, line := range f.lines {
		if line.IsBlank || line.IsCommentOnly {
			continue
		}

		labelWidth := len(line.Label)
		if labelWidth > maxLabelWidth {
			maxLabelWidth = labelWidth
		}

		instrWidth := len(line.Instruction)
		if instrWidth > maxInstructionWidth {
			maxInstructionWidth = instrWidth
		}

		operandWidth := len(line.Operands)
		if operandWidth > maxOperandWidth {
			maxOperandWidth = operandWidth
		}
	}

	log.Debug("Formatter Phase 3: Max widths - Label: %d, Instruction: %d, Operand: %d",
		maxLabelWidth, maxInstructionWidth, maxOperandWidth)

	// Store these for reconstruction
	// For now, we'll use them in phase4, but we could store them in the struct if needed
}

// phase4Reconstruct builds the final formatted text
func (f *Formatter) phase4Reconstruct() string {
	var result strings.Builder

	indentStr := f.makeIndentString(1)

	for i, line := range f.lines {
		// Blank lines - preserve as-is
		if line.IsBlank {
			result.WriteString("\n")
			continue
		}

		// Comment-only lines
		if line.IsCommentOnly {
			// Indent comment-only lines
			for j := 0; j < line.IndentLevel; j++ {
				result.WriteString(indentStr)
			}
			result.WriteString(line.Comment)
			result.WriteString("\n")
			continue
		}

		// Block end lines
		if line.BlockEnd {
			// Indent the closing brace
			for j := 0; j < line.IndentLevel; j++ {
				result.WriteString(indentStr)
			}
			result.WriteString("}")
			if line.Comment != "" {
				result.WriteString(" ")
				result.WriteString(line.Comment)
			}
			result.WriteString("\n")
			continue
		}

		// Regular lines with label/instruction/operands
		lineStr := ""

		// Calculate effective indent level
		// Instructions without labels get +1 indent for readability
		effectiveIndent := line.IndentLevel
		if line.Label == "" && line.Instruction != "" && !line.IsDirective {
			effectiveIndent++
		}

		// Add base indentation
		for j := 0; j < effectiveIndent; j++ {
			lineStr += indentStr
		}

		// Add label (if present)
		if line.Label != "" {
			lineStr += line.Label
			// If there's an instruction after the label, add spacing
			if line.Instruction != "" {
				lineStr += " "
			}
		}

		// Add instruction
		if line.Instruction != "" {
			lineStr += line.Instruction

			// Add operands
			if line.Operands != "" {
				lineStr += " " + line.Operands
			}

			// Handle block start
			if line.BlockStart {
				lineStr += " {"
			}
		}

		// Add comment
		if line.Comment != "" {
			if lineStr != "" {
				lineStr += "  " // Two spaces before comment
			}
			lineStr += line.Comment
		}

		result.WriteString(lineStr)

		// Add newline except for last line (to avoid adding extra newline at end)
		if i < len(f.lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// makeIndentString creates an indent string based on config
func (f *Formatter) makeIndentString(level int) string {
	if f.config.UseSpaces {
		return strings.Repeat(" ", f.config.IndentSize*level)
	}
	return strings.Repeat("\t", level)
}
