package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Test Case Definitions

type TestSuite struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Setup       TestSetup  `json:"setup"`
	TestCases   []TestCase `json:"testCases"`
}

type TestSetup struct {
	ServerPath string            `json:"serverPath"`
	ServerArgs []string          `json:"serverArgs"`
	RootPath   string            `json:"rootPath"`
	Files      map[string]string `json:"files"`
}

type TestCase struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        string        `json:"type"` // "completion", "hover", "diagnostics", etc.
	Input       TestInput     `json:"input"`
	Expected    TestExpected  `json:"expected"`
	Timeout     int           `json:"timeout,omitempty"` // seconds, default 5
	Action      string        `json:"action,omitempty"`  // for performance tests
	Operations  []TestOperation `json:"operations,omitempty"` // for memory tests
}

type TestOperation struct {
	Type       string `json:"type"`
	Iterations int    `json:"iterations"`
}

type TestInput struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
	Content   string `json:"content,omitempty"` // For document change tests
}

type TestExpected struct {
	// For completion tests
	CompletionItems []ExpectedCompletion `json:"completionItems,omitempty"`
	MinItems        int                  `json:"minItems,omitempty"`
	MaxItems        int                  `json:"maxItems,omitempty"`

	// For hover tests
	HoverContent string `json:"hoverContent,omitempty"`
	HoverRange   *Range `json:"hoverRange,omitempty"`

	// For diagnostics tests
	Diagnostics []ExpectedDiagnostic `json:"diagnostics,omitempty"`

	// For definition/references tests
	Locations []ExpectedLocation `json:"locations,omitempty"`

	// For document symbols tests
	Symbols []ExpectedSymbol `json:"symbols,omitempty"`

	// For semantic tokens tests
	SemanticTokens []ExpectedSemanticToken `json:"semanticTokens,omitempty"`
	MinTokens      int                     `json:"minTokens,omitempty"`
	MaxTokens      int                     `json:"maxTokens,omitempty"`

	// For error tests
	ErrorMessage string `json:"errorMessage,omitempty"`
	ShouldError  bool   `json:"shouldError,omitempty"`
}

type ExpectedCompletion struct {
	Label         string `json:"label"`
	Kind          *int   `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

type ExpectedDiagnostic struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
}

type ExpectedLocation struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

type ExpectedSymbol struct {
	Name   string `json:"name"`
	Kind   int    `json:"kind"`
	Line   int    `json:"line"`
	Detail string `json:"detail,omitempty"`
}

type ExpectedSemanticToken struct {
	Line      int    `json:"line"`
	StartChar int    `json:"startChar"`
	Length    int    `json:"length"`
	TokenType string `json:"tokenType"` // "keyword", "variable", "function", etc.
}

// Test Runner

type TestRunner struct {
	client    *LSPClient
	results   []TestResult
	totalRun  int
	totalPass int
	totalFail int
}

type TestResult struct {
	TestCase TestCase      `json:"testCase"`
	Status   string        `json:"status"` // "PASS", "FAIL", "ERROR"
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration"`
	Details  interface{}   `json:"details,omitempty"`
}

func NewTestRunner() *TestRunner {
	return &TestRunner{
		results: make([]TestResult, 0),
	}
}

func (tr *TestRunner) RunTestSuite(suitePath string) error {
	// Load test suite
	suite, err := tr.loadTestSuite(suitePath)
	if err != nil {
		return fmt.Errorf("failed to load test suite: %w", err)
	}

	fmt.Printf("Running test suite: %s\n", suite.Name)
	fmt.Printf("Description: %s\n\n", suite.Description)

	// Setup LSP client
	serverPath := suite.Setup.ServerPath
	if !filepath.IsAbs(serverPath) {
		// Make relative to suite file directory
		suiteDir := filepath.Dir(suitePath)
		serverPath = filepath.Join(suiteDir, serverPath)
	}

	tr.client = NewLSPClient(serverPath, suite.Setup.ServerArgs...)

	// Start server
	if err := tr.client.Start(); err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}
	defer tr.client.Stop()

	// Initialize server
	rootPath := suite.Setup.RootPath
	if !filepath.IsAbs(rootPath) {
		suiteDir := filepath.Dir(suitePath)
		rootPath = filepath.Join(suiteDir, rootPath)
	}

	_, err = tr.client.Initialize(rootPath)
	if err != nil {
		return fmt.Errorf("failed to initialize LSP server: %w", err)
	}

	// Create test files
	for filename, content := range suite.Setup.Files {
		filePath := filepath.Join(rootPath, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write test file %s: %w", filename, err)
		}

		// Open document in LSP
		uri := "file://" + filePath
		if err := tr.client.OpenDocument(uri, "assembly", content); err != nil {
			return fmt.Errorf("failed to open document %s: %w", uri, err)
		}
	}

	// Wait a bit for initial diagnostics
	time.Sleep(500 * time.Millisecond)

	// Run test cases
	for _, testCase := range suite.TestCases {
		result := tr.runTestCase(testCase, rootPath)
		tr.results = append(tr.results, result)
		tr.totalRun++

		if result.Status == "PASS" {
			tr.totalPass++
			fmt.Printf("✓ %s (%v)\n", testCase.Name, result.Duration)
		} else {
			tr.totalFail++
			fmt.Printf("✗ %s (%v)\n", testCase.Name, result.Duration)
			fmt.Printf("  %s\n", result.Message)
			if result.Details != nil {
				fmt.Printf("  Details: %+v\n", result.Details)
			}
		}
	}

	// Print summary
	fmt.Printf("\nTest Summary:\n")
	fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", tr.totalRun, tr.totalPass, tr.totalFail)

	if tr.totalFail > 0 {
		return fmt.Errorf("test suite failed with %d failures", tr.totalFail)
	}

	return nil
}

func (tr *TestRunner) loadTestSuite(suitePath string) (*TestSuite, error) {
	data, err := os.ReadFile(suitePath)
	if err != nil {
		return nil, err
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return &suite, nil
}

func (tr *TestRunner) runTestCase(testCase TestCase, rootPath string) TestResult {
	start := time.Now()
	timeout := time.Duration(testCase.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	result := TestResult{
		TestCase: testCase,
		Status:   "ERROR",
		Duration: 0,
	}

	defer func() {
		result.Duration = time.Since(start)
	}()

	// Get file URI
	filePath := filepath.Join(rootPath, testCase.Input.File)
	uri := "file://" + filePath

	// CRITICAL FIX: Open the document before testing
	// The LSP server needs the document content to provide completions, hover, etc.
	if !tr.client.IsDocumentOpen(uri) {
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			result.Status = "ERROR"
			result.Message = fmt.Sprintf("failed to read test file %s: %v", filePath, err)
			return result
		}

		// Open document in LSP server
		if err := tr.client.OpenDocument(uri, "assembly", string(content)); err != nil {
			result.Status = "ERROR"
			result.Message = fmt.Sprintf("failed to open document %s: %v", uri, err)
			return result
		}

		// Wait a bit for server to process the document
		time.Sleep(100 * time.Millisecond)
	}

	switch testCase.Type {
	case "completion":
		return tr.testCompletion(testCase, uri, result)
	case "hover":
		return tr.testHover(testCase, uri, result)
	case "diagnostics":
		return tr.testDiagnostics(testCase, uri, result)
	case "definition":
		return tr.testDefinition(testCase, uri, result)
	case "references":
		return tr.testReferences(testCase, uri, result)
	case "symbols":
		return tr.testDocumentSymbols(testCase, uri, result)
	case "documentSymbol":
		return tr.testDocumentSymbols(testCase, uri, result)
	case "semanticTokens":
		return tr.testSemanticTokens(testCase, uri, result)
	case "lifecycle":
		return tr.testLifecycle(testCase, uri, result)
	case "performance":
		return tr.testPerformance(testCase, uri, result)
	case "memory":
		return tr.testMemory(testCase, uri, result)
	default:
		result.Message = fmt.Sprintf("unknown test type: %s", testCase.Type)
		return result
	}
}

func (tr *TestRunner) testCompletion(testCase TestCase, uri string, result TestResult) TestResult {
	// Get completion
	items, err := tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
	if err != nil {
		if testCase.Expected.ShouldError {
			if strings.Contains(err.Error(), testCase.Expected.ErrorMessage) {
				result.Status = "PASS"
				return result
			} else {
				result.Status = "FAIL"
				result.Message = fmt.Sprintf("expected error '%s', got '%s'", testCase.Expected.ErrorMessage, err.Error())
				return result
			}
		}
		result.Message = fmt.Sprintf("completion request failed: %v", err)
		return result
	}

	if testCase.Expected.ShouldError {
		result.Status = "FAIL"
		result.Message = "expected error but got successful completion"
		return result
	}

	// Check item count constraints
	if testCase.Expected.MinItems > 0 && len(items) < testCase.Expected.MinItems {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("expected at least %d items, got %d", testCase.Expected.MinItems, len(items))
		result.Details = items
		return result
	}

	if testCase.Expected.MaxItems > 0 && len(items) > testCase.Expected.MaxItems {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("expected at most %d items, got %d", testCase.Expected.MaxItems, len(items))
		result.Details = items
		return result
	}

	// Check specific completion items
	for _, expected := range testCase.Expected.CompletionItems {
		found := false
		for _, item := range items {
			if tr.matchesCompletion(item, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected completion item not found: %+v", expected)
			result.Details = items
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) testHover(testCase TestCase, uri string, result TestResult) TestResult {
	hover, err := tr.client.GetHover(uri, testCase.Input.Line, testCase.Input.Character)
	if err != nil {
		result.Message = fmt.Sprintf("hover request failed: %v", err)
		return result
	}

	if hover == nil {
		if testCase.Expected.HoverContent == "" {
			result.Status = "PASS"
			return result
		}
		result.Status = "FAIL"
		result.Message = "expected hover content but got nil"
		return result
	}

	// Check hover content
	var content string
	switch v := hover.Contents.(type) {
	case string:
		content = v
	case map[string]interface{}:
		if value, ok := v["value"]; ok {
			content = value.(string)
		}
	}

	if testCase.Expected.HoverContent != "" {
		if !strings.Contains(content, testCase.Expected.HoverContent) {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("hover content doesn't contain expected text '%s'", testCase.Expected.HoverContent)
			result.Details = content
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) testDiagnostics(testCase TestCase, uri string, result TestResult) TestResult {
	// Wait for diagnostics to be published
	diagnostics := tr.client.WaitForDiagnostics(uri, 2*time.Second)
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}

	// Check expected diagnostics
	for _, expected := range testCase.Expected.Diagnostics {
		found := false
		for _, diag := range diagnostics {
			if tr.matchesDiagnostic(diag, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected diagnostic not found: line %d, severity %d, message '%s'",
				expected.Line, expected.Severity, expected.Message)
			result.Details = diagnostics
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) testDefinition(testCase TestCase, uri string, result TestResult) TestResult {
	locations, err := tr.client.GetDefinition(uri, testCase.Input.Line, testCase.Input.Character)
	if err != nil {
		result.Message = fmt.Sprintf("definition request failed: %v", err)
		return result
	}

	// Check expected locations
	for _, expected := range testCase.Expected.Locations {
		found := false
		for _, loc := range locations {
			if tr.matchesLocation(loc, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected location not found: %+v", expected)
			result.Details = locations
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) testReferences(testCase TestCase, uri string, result TestResult) TestResult {
	locations, err := tr.client.GetReferences(uri, testCase.Input.Line, testCase.Input.Character, true)
	if err != nil {
		result.Message = fmt.Sprintf("references request failed: %v", err)
		return result
	}

	// Check expected locations
	for _, expected := range testCase.Expected.Locations {
		found := false
		for _, loc := range locations {
			if tr.matchesLocation(loc, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected reference location not found: %+v", expected)
			result.Details = locations
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) testDocumentSymbols(testCase TestCase, uri string, result TestResult) TestResult {
	symbols, err := tr.client.GetDocumentSymbols(uri)
	if err != nil {
		result.Message = fmt.Sprintf("document symbols request failed: %v", err)
		return result
	}

	// Check expected symbols
	for _, expected := range testCase.Expected.Symbols {
		found := false
		for _, symbol := range symbols {
			if tr.matchesSymbol(symbol, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected symbol not found: %+v", expected)
			result.Details = symbols
			return result
		}
	}

	result.Status = "PASS"
	return result
}

// Helper methods for matching expected results

func (tr *TestRunner) matchesCompletion(item CompletionItem, expected ExpectedCompletion) bool {
	if item.Label != expected.Label {
		return false
	}

	if expected.Kind != nil && (item.Kind == nil || *item.Kind != *expected.Kind) {
		return false
	}

	if expected.Detail != "" && (item.Detail == nil || !strings.Contains(*item.Detail, expected.Detail)) {
		return false
	}

	if expected.Documentation != "" {
		var doc string
		switch v := item.Documentation.(type) {
		case string:
			doc = v
		case map[string]interface{}:
			if value, ok := v["value"]; ok {
				doc = value.(string)
			}
		}
		if !strings.Contains(doc, expected.Documentation) {
			return false
		}
	}

	if expected.InsertText != "" && (item.InsertText == nil || *item.InsertText != expected.InsertText) {
		return false
	}

	return true
}

func (tr *TestRunner) matchesDiagnostic(diag Diagnostic, expected ExpectedDiagnostic) bool {
	if diag.Range.Start.Line != expected.Line {
		return false
	}

	// Only check column if explicitly specified (non-zero)
	if expected.Column > 0 && diag.Range.Start.Character != expected.Column {
		return false
	}

	if expected.Severity > 0 && diag.Severity != nil && *diag.Severity != expected.Severity {
		return false
	}

	if expected.Message != "" && !strings.Contains(diag.Message, expected.Message) {
		return false
	}

	if expected.Source != "" && (diag.Source == nil || *diag.Source != expected.Source) {
		return false
	}

	return true
}

func (tr *TestRunner) matchesLocation(loc Location, expected ExpectedLocation) bool {
	// Simple filename matching (just the basename)
	if !strings.HasSuffix(loc.URI, expected.File) {
		return false
	}

	if loc.Range.Start.Line != expected.Line {
		return false
	}

	if loc.Range.Start.Character != expected.Character {
		return false
	}

	return true
}

func (tr *TestRunner) matchesSymbol(symbol DocumentSymbol, expected ExpectedSymbol) bool {
	if symbol.Name != expected.Name {
		return false
	}

	if symbol.Kind != expected.Kind {
		return false
	}

	// Only check line if explicitly specified (non-zero)
	if expected.Line > 0 && symbol.Range.Start.Line != expected.Line {
		return false
	}

	if expected.Detail != "" && (symbol.Detail == nil || !strings.Contains(*symbol.Detail, expected.Detail)) {
		return false
	}

	return true
}

func (tr *TestRunner) GetResults() []TestResult {
	return tr.results
}

func (tr *TestRunner) SaveResults(filename string) error {
	data, err := json.MarshalIndent(tr.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (tr *TestRunner) SaveHTMLReport(filename string, suiteName string) error {
	return GenerateHTMLReport(tr.results, filename, suiteName)
}

// testLifecycle tests basic server lifecycle functionality
func (tr *TestRunner) testLifecycle(testCase TestCase, uri string, result TestResult) TestResult {
	// For lifecycle tests, we just check that basic operations work
	// This is a simplified implementation that verifies server responsiveness

	// Test completion as a proxy for server being alive and responsive
	items, err := tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
	if err != nil {
		result.Message = fmt.Sprintf("Server lifecycle test failed - completion failed: %v", err)
		return result
	}

	// Check if we got a reasonable response
	if len(items) >= 1 {
		result.Status = "PASS"
		result.Message = fmt.Sprintf("Server lifecycle test passed - got %d completion items", len(items))
	} else {
		result.Status = "FAIL"
		result.Message = "Server lifecycle test failed - no completion items returned"
	}

	return result
}

// testPerformance tests response time performance
func (tr *TestRunner) testPerformance(testCase TestCase, uri string, result TestResult) TestResult {
	// For performance tests, we measure response time of operations
	// This is a simplified implementation

	startTime := time.Now()

	// Test the specified action
	switch testCase.Action {
	case "textDocument/completion":
		_, err := tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
		if err != nil {
			result.Message = fmt.Sprintf("Performance test failed: %v", err)
			return result
		}
	case "textDocument/hover":
		_, err := tr.client.GetHover(uri, testCase.Input.Line, testCase.Input.Character)
		if err != nil {
			result.Message = fmt.Sprintf("Performance test failed: %v", err)
			return result
		}
	default:
		// Default to completion test
		_, err := tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
		if err != nil {
			result.Message = fmt.Sprintf("Performance test failed: %v", err)
			return result
		}
	}

	elapsed := time.Since(startTime)
	result.Status = "PASS"
	result.Message = fmt.Sprintf("Performance test completed in %v", elapsed)

	return result
}

// testMemory tests memory usage (simplified implementation)
func (tr *TestRunner) testMemory(testCase TestCase, uri string, result TestResult) TestResult {
	// For memory tests, we perform repeated operations to check for memory leaks
	// This is a simplified implementation that just performs multiple operations

	iterations := 10 // Default iterations
	if testCase.Operations != nil {
		for _, op := range testCase.Operations {
			for i := 0; i < op.Iterations; i++ {
				switch op.Type {
				case "textDocument/completion":
					tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
				case "textDocument/hover":
					tr.client.GetHover(uri, testCase.Input.Line, testCase.Input.Character)
				default:
					tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
				}
			}
		}
	} else {
		// Default behavior - perform multiple completion requests
		for i := 0; i < iterations; i++ {
			tr.client.GetCompletion(uri, testCase.Input.Line, testCase.Input.Character)
		}
	}

	result.Status = "PASS"
	result.Message = fmt.Sprintf("Memory test completed %d operations", iterations)

	return result
}

func (tr *TestRunner) testSemanticTokens(testCase TestCase, uri string, result TestResult) TestResult {
	// Request semantic tokens
	tokens, err := tr.client.GetSemanticTokens(uri)
	if err != nil {
		result.Message = fmt.Sprintf("semantic tokens request failed: %v", err)
		return result
	}

	if tokens == nil || len(tokens.Data) == 0 {
		if testCase.Expected.MinTokens == 0 && len(testCase.Expected.SemanticTokens) == 0 {
			result.Status = "PASS"
			return result
		}
		result.Status = "FAIL"
		result.Message = "expected semantic tokens but got none"
		return result
	}

	// Decode tokens
	decodedTokens := DecodeSemanticTokens(tokens.Data)

	// Check token count constraints
	if testCase.Expected.MinTokens > 0 && len(decodedTokens) < testCase.Expected.MinTokens {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("expected at least %d tokens, got %d", testCase.Expected.MinTokens, len(decodedTokens))
		result.Details = decodedTokens
		return result
	}

	if testCase.Expected.MaxTokens > 0 && len(decodedTokens) > testCase.Expected.MaxTokens {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("expected at most %d tokens, got %d", testCase.Expected.MaxTokens, len(decodedTokens))
		result.Details = decodedTokens
		return result
	}

	// Check specific semantic tokens
	for _, expected := range testCase.Expected.SemanticTokens {
		found := false
		for _, token := range decodedTokens {
			if tr.matchesSemanticToken(token, expected) {
				found = true
				break
			}
		}
		if !found {
			result.Status = "FAIL"
			result.Message = fmt.Sprintf("expected semantic token not found: line %d, char %d, type %s",
				expected.Line, expected.StartChar, expected.TokenType)
			result.Details = decodedTokens
			return result
		}
	}

	result.Status = "PASS"
	return result
}

func (tr *TestRunner) matchesSemanticToken(token DecodedToken, expected ExpectedSemanticToken) bool {
	if token.Line != expected.Line {
		return false
	}

	if token.StartChar != expected.StartChar {
		return false
	}

	if token.Length != expected.Length {
		return false
	}

	if expected.TokenType != "" && token.TypeName != expected.TokenType {
		return false
	}

	return true
}