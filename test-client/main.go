package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	if *testSuite == "" {
		fmt.Println("Usage: test-client -suite <test-suite.json>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  test-client -suite basic-completion.json")
		fmt.Println("  test-client -suite diagnostics.json -verbose")
		fmt.Println("  test-client -interactive -server ./6510lsp_server")
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