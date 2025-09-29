package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		log.Warn("Invalid command line argument: %v. Valid flags are: --debug, --warn-unused-labels, --test", err)
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

	// Start LSP server
	lsp.Start(mnemonicPath, kickassDir)
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
