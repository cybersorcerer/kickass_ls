package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "c64.nvim/internal/log"
	lsp "c64.nvim/internal/lsp"
)

func main() {
	// Log file management is handled by internal/log package

	// Parse command line flags
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stderr)
	debug := flag.Bool("debug", false, "Enable debug logging")
	warnUnused := flag.Bool("warn-unused-labels", false, "Enable warnings for unused labels")
	testFile := flag.String("test", "", "Test mode: analyze file and output diagnostics")
	testCompletion := flag.String("test-completion", "", "Test completion at file:line:char")
	testHover := flag.String("test-hover", "", "Test hover at file:line:char")
	testSignature := flag.String("test-signature", "", "Test signature help at file:line:char")
	testSymbols := flag.String("test-symbols", "", "Test symbol listing for file")
	testReferences := flag.String("test-references", "", "Test find references at file:line:char")
	testGotoDef := flag.String("test-goto-definition", "", "Test go-to-definition at file:line:char")

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		log.Warn("Invalid command line argument: %v. Valid flags are: --debug, --warn-unused-labels, --test, --test-completion, --test-hover, --test-signature, --test-symbols, --test-references, --test-goto-definition", err)
	}

	// Set log level
	if *debug {
		log.SetLevel(log.DEBUG)
	} else {
		log.SetLevel(log.INFO)
	}

	lsp.SetWarnUnusedLabels(*warnUnused)

	// Initialize logger
	if err := log.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(11)
	}

	log.Info("6510 Language Server started.")

	// Get executable directory for config files
	exePath, err := os.Executable()
	if err != nil {
		log.Error("Failed to get executable path: %v", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	// Set paths for configuration files
	mnemonicPath := filepath.Join(exeDir, "mnemonic.json")
	kickassDir := exeDir

	// Check if test mode is requested
	if *testFile != "" {
		runTestMode(*testFile, mnemonicPath, kickassDir)
		return
	}

	// Check for LSP feature testing modes
	if *testCompletion != "" {
		runCompletionTest(*testCompletion, mnemonicPath, kickassDir)
		return
	}

	if *testHover != "" {
		runHoverTest(*testHover, mnemonicPath, kickassDir)
		return
	}

	if *testSignature != "" {
		runSignatureTest(*testSignature, mnemonicPath, kickassDir)
		return
	}

	if *testSymbols != "" {
		runSymbolsTest(*testSymbols, mnemonicPath, kickassDir)
		return
	}

	if *testReferences != "" {
		runReferencesTest(*testReferences, mnemonicPath, kickassDir)
		return
	}

	if *testGotoDef != "" {
		runGotoDefinitionTest(*testGotoDef, mnemonicPath, kickassDir)
		return
	}

	// Start LSP server
	lsp.Start()
}

// runTestMode analyzes a single file and outputs diagnostics
func runTestMode(filename, mnemonicPath, kickassDir string) {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		os.Exit(3)
	}

	// Initialize the lexer path (needed for mnemonic loading)
	lsp.SetKickassJSONPath(kickassDir)

	// Load configuration files (same as LSP mode)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document and run semantic analysis
	text := string(content)
	_, allDiagnostics := lsp.ParseDocument(filename, text)

	// Output results
	if len(allDiagnostics) == 0 {
		fmt.Println("No issues found.")
		os.Exit(0)
	}

	// Count errors and warnings
	errorCount := 0
	warningCount := 0
	for _, diag := range allDiagnostics {
		if diag.Severity == lsp.SeverityError {
			errorCount++
		} else if diag.Severity == lsp.SeverityWarning {
			warningCount++
		}
	}

	// Output diagnostics in text format
	for _, diag := range allDiagnostics {
		severity := "info"
		switch diag.Severity {
		case lsp.SeverityError:
			severity = "error"
		case lsp.SeverityWarning:
			severity = "warning"
		case lsp.SeverityHint:
			severity = "hint"
		}
		fmt.Printf("%s:%d:%d: %s: %s\n", filename, diag.Range.Start.Line+1, diag.Range.Start.Character+1, severity, diag.Message)
	}

	// Print summary
	fmt.Printf("\nSummary: %d warnings, %d errors\n", warningCount, errorCount)

	// Exit with appropriate code
	if errorCount > 0 {
		os.Exit(2)
	} else if warningCount > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// initTestMode initializes LSP components for test modes using config directory
func initTestMode() error {
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "6510lsp")

	// Initialize LSP components with config directory paths
	lsp.SetKickassJSONPath(filepath.Join(configDir, "kickass.json"))
	lsp.SetMnemonicJSONPath(filepath.Join(configDir, "mnemonic.json"))

	err = lsp.LoadMnemonics(filepath.Join(configDir, "mnemonic.json"))
	if err != nil {
		return fmt.Errorf("error loading mnemonics: %v", err)
	}

	_, err = lsp.LoadKickassDirectives(configDir)
	if err != nil {
		return fmt.Errorf("error loading kickass directives: %v", err)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(configDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		return fmt.Errorf("error loading built-ins: %v", err)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Load C64 memory map data
	err = lsp.LoadC64MemoryMap(filepath.Join(configDir, "c64memory.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load C64 memory map: %v\n", err)
	}

	return nil
}

// parseFilePosition parses file:line:char format
func parseFilePosition(filePos string) (file string, line int, char int, err error) {
	parts := strings.Split(filePos, ":")
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid format, expected file:line:char")
	}

	file = parts[0]

	line, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid line number: %s", parts[1])
	}

	char, err = strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid character number: %s", parts[2])
	}

	// Convert to 0-based indexing (LSP uses 0-based)
	line--
	char--

	return file, line, char, nil
}

// runCompletionTest tests completion at a specific position
func runCompletionTest(filePos, mnemonicPath, kickassDir string) {
	file, line, char, err := parseFilePosition(filePos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing position: %v\n", err)
		os.Exit(3)
	}

	// Read the file
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		os.Exit(3)
	}

	// Initialize LSP components
	lsp.SetKickassJSONPath(kickassDir)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(file, text)

	// Get the line content for completion context
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		fmt.Fprintf(os.Stderr, "Line %d is out of range (file has %d lines)\n", line+1, len(lines))
		os.Exit(3)
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		fmt.Fprintf(os.Stderr, "Character %d is out of range for line %d\n", char+1, line+1)
		os.Exit(3)
	}

	// Get completion context
	isOperand, wordToComplete := lsp.GetCompletionContext(currentLine, char)

	// Generate completions
	completions := lsp.GenerateCompletions(scope, line, isOperand, wordToComplete)

	// Output results
	fmt.Printf("Completion at %s:%d:%d:\n", file, line+1, char+1)
	if len(completions) == 0 {
		fmt.Println("No completions available")
		os.Exit(0)
	}

	// Categorize and count completions
	categories := make(map[string]int)
	for _, completion := range completions {
		if kind, ok := completion["kind"].(float64); ok {
			switch int(kind) {
			case 1: // Text
				categories["text"]++
			case 3: // Function
				categories["function"]++
			case 6: // Variable
				categories["variable"]++
			case 13: // Enum
				categories["constant"]++
			case 14: // Instruction
				categories["instruction"]++
			case 21: // Constant
				categories["constant"]++
			default:
				categories["other"]++
			}
		}

		if label, ok := completion["label"].(string); ok {
			if detail, ok := completion["detail"].(string); ok {
				fmt.Printf("- %s (%s)\n", label, detail)
			} else {
				fmt.Printf("- %s\n", label)
			}
		}
	}

	// Print summary
	fmt.Printf("\nSummary: %d items", len(completions))
	if len(categories) > 0 {
		fmt.Printf(" (")
		first := true
		for category, count := range categories {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%d %s", count, category)
			first = false
		}
		fmt.Printf(")")
	}
	fmt.Println()

	os.Exit(0)
}

// runHoverTest tests hover information at a specific position
func runHoverTest(filePos, mnemonicPath, kickassDir string) {
	file, line, char, err := parseFilePosition(filePos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing position: %v\n", err)
		os.Exit(3)
	}

	// Read the file
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		os.Exit(3)
	}

	// Initialize LSP components
	err = initTestMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(3)
	}

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(file, text)

	// Get the line content for hover context
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		fmt.Fprintf(os.Stderr, "Line %d is out of range (file has %d lines)\n", line+1, len(lines))
		os.Exit(3)
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		fmt.Fprintf(os.Stderr, "Character %d is out of range for line %d\n", char+1, line+1)
		os.Exit(3)
	}

	// Generate hover information
	hoverContent, found := lsp.GenerateHover(scope, currentLine, char)

	// Output results
	fmt.Printf("Hover at %s:%d:%d:\n", file, line+1, char+1)
	if !found {
		fmt.Println("No hover information available")
		os.Exit(0)
	}

	// Print the markdown content in a readable format
	fmt.Printf("%s\n", hoverContent)
	os.Exit(0)
}

// runSignatureTest tests signature help at a specific position
func runSignatureTest(filePos, mnemonicPath, kickassDir string) {
	file, line, char, err := parseFilePosition(filePos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing position: %v\n", err)
		os.Exit(3)
	}

	// Read the file
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		os.Exit(3)
	}

	// Initialize LSP components
	lsp.SetKickassJSONPath(kickassDir)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(file, text)

	// Get the line content for signature context
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		fmt.Fprintf(os.Stderr, "Line %d is out of range (file has %d lines)\n", line+1, len(lines))
		os.Exit(3)
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		fmt.Fprintf(os.Stderr, "Character %d is out of range for line %d\n", char+1, line+1)
		os.Exit(3)
	}

	// Generate signature help
	signature, found := lsp.GenerateSignatureHelp(scope, currentLine, char)

	// Output results
	fmt.Printf("Signature help at %s:%d:%d:\n", file, line+1, char+1)
	if !found {
		fmt.Println("No signature help available")
		os.Exit(0)
	}

	fmt.Printf("%s\n", signature)
	os.Exit(0)
}

// runSymbolsTest lists all symbols in a file
func runSymbolsTest(filename, mnemonicPath, kickassDir string) {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		os.Exit(3)
	}

	// Initialize LSP components
	lsp.SetKickassJSONPath(kickassDir)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(filename, text)

	// Generate symbol listing
	symbols := lsp.ListSymbols(scope)

	// Output results
	fmt.Printf("Symbols in %s:\n", filename)
	if len(symbols) == 0 {
		fmt.Println("No symbols found")
		os.Exit(0)
	}

	// Group symbols by type
	categories := make(map[string][]map[string]interface{})
	for _, symbol := range symbols {
		symbolType := symbol["type"].(string)
		categories[symbolType] = append(categories[symbolType], symbol)
	}

	// Display symbols by category
	totalSymbols := 0
	for category, syms := range categories {
		if len(syms) > 0 {
			fmt.Printf("\n%s:\n", strings.Title(category))
			for _, sym := range syms {
				name := sym["name"].(string)
				location := sym["location"].(string)
				if detail, ok := sym["detail"].(string); ok && detail != "" {
					fmt.Printf("- %s (%s) %s\n", name, location, detail)
				} else {
					fmt.Printf("- %s (%s)\n", name, location)
				}
			}
			totalSymbols += len(syms)
		}
	}

	// Add built-ins summary
	fmt.Printf("\nBuilt-ins:\n")
	builtinFunctions, builtinConstants = lsp.GetBuiltins()
	if len(builtinFunctions) > 0 {
		fmt.Printf("- %d functions (sin, cos, pow, etc.)\n", len(builtinFunctions))
	}
	if len(builtinConstants) > 0 {
		fmt.Printf("- %d constants (PI, E, color values, etc.)\n", len(builtinConstants))
	}

	fmt.Printf("\nSummary: %d user symbols, %d built-in functions, %d built-in constants\n",
		totalSymbols, len(builtinFunctions), len(builtinConstants))
	os.Exit(0)
}

// runReferencesTest finds all references to a symbol at a specific position
func runReferencesTest(filePos, mnemonicPath, kickassDir string) {
	file, line, char, err := parseFilePosition(filePos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing position: %v\n", err)
		os.Exit(3)
	}

	// Read the file
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		os.Exit(3)
	}

	// Initialize LSP components
	lsp.SetKickassJSONPath(kickassDir)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(file, text)

	// Get the line content for reference context
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		fmt.Fprintf(os.Stderr, "Line %d is out of range (file has %d lines)\n", line+1, len(lines))
		os.Exit(3)
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		fmt.Fprintf(os.Stderr, "Character %d is out of range for line %d\n", char+1, line+1)
		os.Exit(3)
	}

	// Find references
	references, symbolName := lsp.FindReferences(scope, currentLine, char, line)

	// Output results
	fmt.Printf("References for '%s' at %s:%d:%d:\n", symbolName, file, line+1, char+1)
	if len(references) == 0 {
		fmt.Println("No references found")
		os.Exit(0)
	}

	definitionCount := 0
	usageCount := 0
	for _, ref := range references {
		refType := ref["type"].(string)
		location := ref["location"].(string)
		fmt.Printf("- %s (%s)\n", location, refType)
		if refType == "definition" {
			definitionCount++
		} else {
			usageCount++
		}
	}

	fmt.Printf("\nSummary: %d references (%d definition, %d usages)\n",
		len(references), definitionCount, usageCount)
	os.Exit(0)
}

// runGotoDefinitionTest finds the definition of a symbol at a specific position
func runGotoDefinitionTest(filePos, mnemonicPath, kickassDir string) {
	file, line, char, err := parseFilePosition(filePos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing position: %v\n", err)
		os.Exit(3)
	}

	// Read the file
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
		os.Exit(3)
	}

	// Initialize LSP components
	lsp.SetKickassJSONPath(kickassDir)
	err = lsp.LoadMnemonics(mnemonicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading mnemonics: %v\n", err)
		os.Exit(3)
	}

	_, err = lsp.LoadKickassDirectives(kickassDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kickass directives: %v\n", err)
		os.Exit(3)
	}

	// Load built-in functions and constants
	kickassJSONPath := filepath.Join(kickassDir, "kickass.json")
	builtinFunctions, builtinConstants, err := lsp.LoadBuiltins(kickassJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading built-ins: %v\n", err)
		os.Exit(3)
	}
	lsp.SetBuiltins(builtinFunctions, builtinConstants)

	// Parse the document to get symbol tree
	text := string(content)
	scope, _ := lsp.ParseDocument(file, text)

	// Get the line content for definition context
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		fmt.Fprintf(os.Stderr, "Line %d is out of range (file has %d lines)\n", line+1, len(lines))
		os.Exit(3)
	}

	currentLine := lines[line]
	if char > len(currentLine) {
		fmt.Fprintf(os.Stderr, "Character %d is out of range for line %d\n", char+1, line+1)
		os.Exit(3)
	}

	// Find definition
	definition, symbolName, found := lsp.GotoDefinition(scope, currentLine, char)

	// Output results
	fmt.Printf("Go-to-definition at %s:%d:%d:\n", file, line+1, char+1)
	if !found {
		fmt.Println("No definition found")
		os.Exit(0)
	}

	fmt.Printf("Symbol: '%s'\n", symbolName)
	if definition["type"].(string) == "built-in" {
		fmt.Printf("Built-in %s - no source location\n", definition["kind"].(string))
	} else {
		fmt.Printf("Definition: %s\n", definition["location"].(string))
		fmt.Printf("Type: %s\n", definition["type"].(string))
	}

	os.Exit(0)
}
