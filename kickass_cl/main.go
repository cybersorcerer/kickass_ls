package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		testSuite = flag.String("suite", "", "Path to test suite JSON file")
		serverPath = flag.String("server", "../6510lsp_server", "Path to LSP server executable")
		serverArgs = flag.String("args", "", "Additional server arguments")
		rootPath = flag.String("root", ".", "Root path for test files")
		outputFile = flag.String("output", "", "Save test results to JSON file")
		verbose = flag.Bool("verbose", false, "Verbose output")
		interactive = flag.Bool("interactive", false, "Interactive mode for manual testing")
	)
	flag.Parse()

	if *interactive {
		runInteractiveMode(*serverPath, *serverArgs, *rootPath, *verbose)
		return
	}

	// Check if a file path is provided as positional argument (quick test mode)
	if flag.NArg() > 0 {
		filePath := flag.Arg(0)
		// Special case: if file is "completion-test", run completion test
		if filePath == "completion-test" {
			runCompletionTest(*serverPath, *verbose)
			return
		}
		// Special case: if file is "completion-at", run completion at position test
		if filePath == "completion-at" {
			if flag.NArg() < 4 {
				fmt.Println("Usage: test-client completion-at <file> <line> <char>")
				fmt.Println("Example: test-client completion-at test.asm 5 8")
				os.Exit(1)
			}
			file := flag.Arg(1)
			var line, char int
			fmt.Sscanf(flag.Arg(2), "%d", &line)
			fmt.Sscanf(flag.Arg(3), "%d", &char)
			runCompletionAtPosition(*serverPath, file, line, char, *verbose)
			return
		}
		runQuickTest(*serverPath, filePath, *verbose)
		return
	}

	if *testSuite == "" {
		fmt.Println("Usage:")
		fmt.Println("  test-client <file.asm>                    - Quick test a single file")
		fmt.Println("  test-client -suite <test-suite.json>     - Run a test suite")
		fmt.Println("  test-client -interactive                 - Interactive mode")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  test-client test.asm")
		fmt.Println("  test-client comprehensive-test.asm -verbose")
		fmt.Println("  test-client -suite basic-completion.json")
		fmt.Println("  test-client -interactive -server ./kickass_ls")
		os.Exit(1)
	}

	// Run test suite
	runner := NewTestRunner()
	err := runner.RunTestSuite(*testSuite)

	// Save results if requested
	if *outputFile != "" {
		if saveErr := runner.SaveResults(*outputFile); saveErr != nil {
			fmt.Printf("Failed to save results: %v\n", saveErr)
		} else {
			fmt.Printf("Results saved to %s\n", *outputFile)
		}
	}

	if err != nil {
		fmt.Printf("Test suite failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All tests passed!")
}

func runQuickTest(serverPath, filePath string, verbose bool) {
	fmt.Printf("Testing file: %s\n", filePath)
	fmt.Println("=====================================")

	// Create client
	client := NewLSPClient(serverPath)

	// Start server
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer client.Stop()

	// Test file
	diagnostics, err := client.TestFile(filePath)
	if err != nil {
		fmt.Printf("❌ Test failed: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if len(diagnostics) == 0 {
		fmt.Println("✅ No diagnostics - file is clean!")
		return
	}

	fmt.Printf("\n📋 Diagnostics (%d):\n", len(diagnostics))
	fmt.Println("-------------------------------------")

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, diag := range diagnostics {
		severity := "Info"
		icon := "ℹ️"
		if diag.Severity != nil {
			switch *diag.Severity {
			case 1:
				severity = "Error"
				icon = "❌"
				errorCount++
			case 2:
				severity = "Warning"
				icon = "⚠️"
				warningCount++
			case 3:
				severity = "Info"
				icon = "ℹ️"
				infoCount++
			case 4:
				severity = "Hint"
				icon = "💡"
				infoCount++
			}
		}

		fmt.Printf("%s Line %d:%d [%s] %s\n",
			icon,
			diag.Range.Start.Line+1,
			diag.Range.Start.Character+1,
			severity,
			diag.Message)

		if verbose && diag.Source != nil {
			fmt.Printf("   Source: %s\n", *diag.Source)
		}
	}

	fmt.Println("-------------------------------------")
	fmt.Printf("Summary: %d errors, %d warnings, %d info/hints\n",
		errorCount, warningCount, infoCount)

	if errorCount > 0 {
		os.Exit(1)
	}
}

func runCompletionTest(serverPath string, verbose bool) {
	fmt.Println("=== Completion Test: Testing '.' character ===")
	fmt.Println()

	// Create client
	client := NewLSPClient(serverPath)

	// Start server
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer client.Stop()

	// Initialize
	rootPath, _ := os.Getwd()
	_, err := client.Initialize(rootPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	// Create test content with just a dot
	content := "*=$800\n."
	uri := "file:///test_completion_dot.asm"

	// Open document
	err = client.OpenDocument(uri, "kickasm", content)
	if err != nil {
		fmt.Printf("❌ Failed to open document: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Document content:\n%s\n", content)
	fmt.Println()

	// Wait a bit for document to be parsed
	fmt.Println("Waiting for document to be parsed...")
	// Give server time to process
	time.Sleep(200 * time.Millisecond)

	// Request completion at line 1, character 1 (right after the '.')
	fmt.Println("Requesting completion at line=1, char=1 (after '.')")
	completions, err := client.GetCompletion(uri, 1, 1)
	if err != nil {
		fmt.Printf("❌ Failed to get completion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== Got %d completion items ===\n\n", len(completions))

	// Analyze completion results
	hasDirectives := false
	hasMnemonics := false
	directiveCount := 0
	mnemonicCount := 0

	fmt.Println("First 30 items:")
	maxShow := 30
	if len(completions) < maxShow {
		maxShow = len(completions)
	}

	for i := 0; i < maxShow; i++ {
		item := completions[i]
		fmt.Printf("%3d. %-20s %-30s [kind=%v]\n", i+1, item.Label, item.Detail, item.Kind)

		// Analyze
		if len(item.Label) > 0 && item.Label[0] == '.' {
			hasDirectives = true
			directiveCount++
		}
		label := item.Label
		if label == "lda" || label == "sta" || label == "jmp" || label == "nop" {
			hasMnemonics = true
			mnemonicCount++
		}
	}

	if len(completions) > maxShow {
		fmt.Printf("\n... and %d more items\n", len(completions)-maxShow)
	}

	// Final analysis
	fmt.Println("\n=== Analysis ===")
	if hasDirectives {
		fmt.Printf("✅ Has %d directives starting with '.'\n", directiveCount)
	} else {
		fmt.Println("❌ NO directives found! (WRONG)")
	}

	if hasMnemonics {
		fmt.Printf("❌ Has %d mnemonics (WRONG - should NOT suggest mnemonics after '.')\n", mnemonicCount)
	} else {
		fmt.Println("✅ No mnemonics found (correct)")
	}

	if !hasDirectives || hasMnemonics {
		fmt.Println("\n❌ COMPLETION TEST FAILED")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ COMPLETION TEST PASSED")
	}
}

func runCompletionAtPosition(serverPath, filePath string, line, char int, verbose bool) {
	fmt.Printf("=== Completion Test at Position ===\n")
	fmt.Printf("File: %s\n", filePath)
	fmt.Printf("Position: Line %d, Char %d\n\n", line, char)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %v\n", err)
		os.Exit(1)
	}

	// Create client
	client := NewLSPClient(serverPath)

	// Start server
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer client.Stop()

	// Initialize
	rootPath, _ := os.Getwd()
	_, err = client.Initialize(rootPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	// Get absolute path for URI
	absPath, _ := filepath.Abs(filePath)
	uri := "file://" + absPath

	// Open document
	err = client.OpenDocument(uri, "kickasm", string(content))
	if err != nil {
		fmt.Printf("❌ Failed to open document: %v\n", err)
		os.Exit(1)
	}

	// Wait for document to be parsed
	time.Sleep(200 * time.Millisecond)

	// Request completion at specified position
	completions, err := client.GetCompletion(uri, line, char)
	if err != nil {
		fmt.Printf("❌ Failed to get completion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Got %d completion items ===\n\n", len(completions))

	// Show all completions
	for i, item := range completions {
		detail := ""
		if item.Detail != nil {
			detail = *item.Detail
		}
		fmt.Printf("%3d. %-25s %s\n", i+1, item.Label, detail)
	}
}

func runInteractiveMode(serverPath, serverArgs, rootPath string, verbose bool) {
	fmt.Println("LSP Test Client - Interactive Mode")
	fmt.Println("==================================")

	// Create a simple test client
	client := NewLSPClient(serverPath)

	// Start server
	if err := client.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer client.Stop()

	// Initialize
	fmt.Println("Initializing LSP server...")
	result, err := client.Initialize(rootPath)
	if err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Server capabilities: %+v\n", result.Capabilities)
	}

	// Interactive commands
	fmt.Println("\nServer initialized. Available commands:")
	fmt.Println("  open <file>           - Open a document")
	fmt.Println("  completion <line> <char> - Get completion at position")
	fmt.Println("  hover <line> <char>      - Get hover info at position")
	fmt.Println("  diagnostics             - Show current diagnostics")
	fmt.Println("  symbols                 - Show document symbols")
	fmt.Println("  quit                    - Exit")
	fmt.Println()

	var currentFile string
	var currentURI string

	for {
		fmt.Print("> ")
		var cmd string
		n, err := fmt.Scanln(&cmd)
		if err != nil || n == 0 {
			// If we can't read from stdin (e.g., running in background), exit
			fmt.Println("No input available, exiting interactive mode")
			return
		}

		switch cmd {
		case "quit", "exit", "q":
			return

		case "open":
			var filename string
			fmt.Print("File path: ")
			fmt.Scanln(&filename)

			if !filepath.IsAbs(filename) {
				filename = filepath.Join(rootPath, filename)
			}

			content, err := os.ReadFile(filename)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				continue
			}

			currentFile = filename
			currentURI = "file://" + filename

			err = client.OpenDocument(currentURI, "assembly", string(content))
			if err != nil {
				fmt.Printf("Error opening document: %v\n", err)
				continue
			}

			fmt.Printf("Opened: %s\n", filename)

		case "completion":
			if currentURI == "" {
				fmt.Println("No document open. Use 'open <file>' first.")
				continue
			}

			var line, char int
			fmt.Print("Line: ")
			fmt.Scanln(&line)
			fmt.Print("Character: ")
			fmt.Scanln(&char)

			items, err := client.GetCompletion(currentURI, line, char)
			if err != nil {
				fmt.Printf("Error getting completion: %v\n", err)
				continue
			}

			fmt.Printf("Completion items (%d):\n", len(items))
			for i, item := range items {
				if i >= 20 { // Limit output
					fmt.Printf("... and %d more items\n", len(items)-i)
					break
				}
				detail := ""
				if item.Detail != nil {
					detail = " - " + *item.Detail
				}
				fmt.Printf("  %s%s\n", item.Label, detail)
			}

		case "hover":
			if currentURI == "" {
				fmt.Println("No document open. Use 'open <file>' first.")
				continue
			}

			var line, char int
			fmt.Print("Line: ")
			fmt.Scanln(&line)
			fmt.Print("Character: ")
			fmt.Scanln(&char)

			hover, err := client.GetHover(currentURI, line, char)
			if err != nil {
				fmt.Printf("Error getting hover: %v\n", err)
				continue
			}

			if hover == nil {
				fmt.Println("No hover information available")
				continue
			}

			var content string
			switch v := hover.Contents.(type) {
			case string:
				content = v
			case map[string]interface{}:
				if value, ok := v["value"]; ok {
					content = value.(string)
				}
			}

			fmt.Printf("Hover content:\n%s\n", content)

		case "diagnostics":
			if currentURI == "" {
				fmt.Println("No document open. Use 'open <file>' first.")
				continue
			}

			diagnostics := client.GetDiagnostics(currentURI)
			if len(diagnostics) == 0 {
				fmt.Println("No diagnostics")
				continue
			}

			fmt.Printf("Diagnostics (%d):\n", len(diagnostics))
			for _, diag := range diagnostics {
				severity := "Info"
				if diag.Severity != nil {
					switch *diag.Severity {
					case 1:
						severity = "Error"
					case 2:
						severity = "Warning"
					case 3:
						severity = "Info"
					case 4:
						severity = "Hint"
					}
				}
				fmt.Printf("  %s:%d:%d [%s] %s\n",
					filepath.Base(currentFile),
					diag.Range.Start.Line+1,
					diag.Range.Start.Character+1,
					severity,
					diag.Message)
			}

		case "symbols":
			if currentURI == "" {
				fmt.Println("No document open. Use 'open <file>' first.")
				continue
			}

			symbols, err := client.GetDocumentSymbols(currentURI)
			if err != nil {
				fmt.Printf("Error getting symbols: %v\n", err)
				continue
			}

			fmt.Printf("Document symbols (%d):\n", len(symbols))
			for _, symbol := range symbols {
				detail := ""
				if symbol.Detail != nil {
					detail = " - " + *symbol.Detail
				}
				fmt.Printf("  %s (line %d)%s\n", symbol.Name, symbol.Range.Start.Line+1, detail)
			}

		default:
			fmt.Println("Unknown command. Available: open, completion, hover, diagnostics, symbols, quit")
		}
	}
}