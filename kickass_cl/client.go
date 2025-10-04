package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LSPClient struct {
	serverPath string
	serverArgs []string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser

	// Request/Response handling
	nextID      int
	pendingReqs map[int]chan *Message
	reqMutex    sync.Mutex

	// Document state
	documents map[string]*DocumentState
	docMutex  sync.RWMutex

	// Diagnostics handling
	diagnostics map[string][]Diagnostic
	diagMutex   sync.RWMutex

	// Shutdown handling
	shutdown chan bool
	done     chan bool
}

type DocumentState struct {
	URI     string
	Version int
	Content string
}

func NewLSPClient(serverPath string, serverArgs ...string) *LSPClient {
	return &LSPClient{
		serverPath:  serverPath,
		serverArgs:  serverArgs,
		nextID:      1,
		pendingReqs: make(map[int]chan *Message),
		documents:   make(map[string]*DocumentState),
		diagnostics: make(map[string][]Diagnostic),
		shutdown:    make(chan bool),
		done:        make(chan bool),
	}
}

func (c *LSPClient) Start() error {
	// Start the LSP server process
	c.cmd = exec.Command(c.serverPath, c.serverArgs...)

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Start message processing goroutines
	go c.readMessages()
	go c.readStderr()

	return nil
}

func (c *LSPClient) Stop() error {
	// Send shutdown request
	c.SendRequest("shutdown", nil)

	// Send exit notification
	c.SendNotification("exit", nil)

	// Close pipes
	if c.stdin != nil {
		c.stdin.Close()
	}

	// Wait for process to exit
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Wait()
	}

	return nil
}

func (c *LSPClient) readMessages() {
	reader := bufio.NewReader(c.stdout)

	for {
		// Read headers
		var contentLength int
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					fmt.Printf("Server closed connection (EOF)\n")
					// Check if server process is still running
					if c.cmd != nil && c.cmd.Process != nil {
						if state := c.cmd.ProcessState; state != nil && state.Exited() {
							fmt.Printf("Server process exited with code: %d\n", state.ExitCode())
						} else {
							fmt.Printf("Server process terminated unexpectedly\n")
						}
					}
				} else {
					fmt.Printf("Error reading header line: %v\n", err)
				}
				// Signal shutdown to prevent broken pipe errors
				if c.done != nil {
					select {
					case <-c.done:
					default:
						close(c.done)
					}
				}
				return
			}

			lineStr := string(line)
			if lineStr == "" {
				// Empty line separates headers from content
				break
			}

			if strings.HasPrefix(lineStr, "Content-Length: ") {
				lengthStr := strings.TrimPrefix(lineStr, "Content-Length: ")
				contentLength, err = strconv.Atoi(strings.TrimSpace(lengthStr))
				if err != nil {
					fmt.Printf("Error parsing Content-Length: %v\n", err)
					break
				}
			}
		}

		if contentLength <= 0 {
			continue
		}

		// Read message body
		content := make([]byte, contentLength)
		n, err := io.ReadFull(reader, content)
		if err != nil || n != contentLength {
			fmt.Printf("Error reading message body: %v (read %d, expected %d)\n", err, n, contentLength)
			continue
		}

		// Parse JSON message
		msg, err := FromJSON(content)
		if err != nil {
			fmt.Printf("Error parsing JSON message: %v\nContent: %s\n", err, string(content))
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *LSPClient) readStderr() {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		fmt.Printf("[SERVER STDERR] %s\n", scanner.Text())
	}
}

func (c *LSPClient) handleMessage(msg *Message) {
	if msg.ID != nil {
		// This is a response to a request
		c.reqMutex.Lock()
		id := int(msg.ID.(float64))
		if ch, ok := c.pendingReqs[id]; ok {
			ch <- msg
			delete(c.pendingReqs, id)
		}
		c.reqMutex.Unlock()
	} else if msg.Method != "" {
		// This is a notification from server
		c.handleNotification(msg)
	}
}

func (c *LSPClient) handleNotification(msg *Message) {
	switch msg.Method {
	case "textDocument/publishDiagnostics":
		var params PublishDiagnosticsParams
		if data, err := json.Marshal(msg.Params); err == nil {
			if err := json.Unmarshal(data, &params); err == nil {
				c.diagMutex.Lock()
				c.diagnostics[params.URI] = params.Diagnostics
				c.diagMutex.Unlock()
			}
		}
	default:
		fmt.Printf("[NOTIFICATION] %s: %+v\n", msg.Method, msg.Params)
	}
}

func (c *LSPClient) SendRequest(method string, params interface{}) (*Message, error) {
	c.reqMutex.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan *Message, 1)
	c.pendingReqs[id] = ch
	c.reqMutex.Unlock()

	msg := &Message{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.sendMessage(msg); err != nil {
		c.reqMutex.Lock()
		delete(c.pendingReqs, id)
		c.reqMutex.Unlock()
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-ch:
		return response, nil
	case <-time.After(5 * time.Second):
		c.reqMutex.Lock()
		delete(c.pendingReqs, id)
		c.reqMutex.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

func (c *LSPClient) SendNotification(method string, params interface{}) error {
	msg := &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.sendMessage(msg)
}

func (c *LSPClient) sendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Format as LSP message with Content-Length header
	content := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	_, err = c.stdin.Write([]byte(content))
	return err
}

// LSP Lifecycle Methods

func (c *LSPClient) Initialize(rootPath string) (*InitializeResult, error) {
	params := InitializeParams{
		ProcessID: func() *int { pid := os.Getpid(); return &pid }(),
		RootPath:  &rootPath,
		RootURI:   func() *string { uri := "file://" + rootPath; return &uri }(),
		Capabilities: ClientCapabilities{
			TextDocument: &TextDocumentClientCapabilities{
				Completion: &CompletionClientCapabilities{
					DynamicRegistration: false,
					CompletionItem: &struct {
						SnippetSupport          bool     `json:"snippetSupport,omitempty"`
						CommitCharactersSupport bool     `json:"commitCharactersSupport,omitempty"`
						DocumentationFormat     []string `json:"documentationFormat,omitempty"`
					}{
						SnippetSupport:      true,
						DocumentationFormat: []string{"markdown", "plaintext"},
					},
				},
				Hover: &HoverClientCapabilities{
					ContentFormat: []string{"markdown", "plaintext"},
				},
				Definition: &DefinitionClientCapabilities{
					DynamicRegistration: false,
				},
				References: &ReferencesClientCapabilities{
					DynamicRegistration: false,
				},
				DocumentSymbol: &DocumentSymbolClientCapabilities{
					DynamicRegistration: false,
				},
			},
		},
	}

	response, err := c.SendRequest("initialize", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", response.Error.Message)
	}

	var result InitializeResult
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse initialize result: %w", err)
		}
	}

	// Send initialized notification
	c.SendNotification("initialized", map[string]interface{}{})

	return &result, nil
}

// Document Methods

func (c *LSPClient) OpenDocument(uri, languageID, content string) error {
	c.docMutex.Lock()
	doc := &DocumentState{
		URI:     uri,
		Version: 1,
		Content: content,
	}
	c.documents[uri] = doc
	c.docMutex.Unlock()

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: languageID,
			Version:    1,
			Text:       content,
		},
	}

	return c.SendNotification("textDocument/didOpen", params)
}

func (c *LSPClient) ChangeDocument(uri, content string) error {
	c.docMutex.Lock()
	doc, exists := c.documents[uri]
	if !exists {
		c.docMutex.Unlock()
		return fmt.Errorf("document %s is not open", uri)
	}
	doc.Version++
	doc.Content = content
	c.docMutex.Unlock()

	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{URI: uri},
			Version:                doc.Version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: content,
			},
		},
	}

	return c.SendNotification("textDocument/didChange", params)
}

func (c *LSPClient) CloseDocument(uri string) error {
	c.docMutex.Lock()
	delete(c.documents, uri)
	c.docMutex.Unlock()

	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	}

	return c.SendNotification("textDocument/didClose", params)
}

// LSP Feature Methods

func (c *LSPClient) GetCompletion(uri string, line, character int) ([]CompletionItem, error) {
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: line, Character: character},
	}

	response, err := c.SendRequest("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("completion error: %s", response.Error.Message)
	}

	var items []CompletionItem
	if data, err := json.Marshal(response.Result); err == nil {
		// Try to parse as CompletionList first
		var list CompletionList
		if err := json.Unmarshal(data, &list); err == nil {
			items = list.Items
		} else {
			// Try to parse as array of CompletionItem
			json.Unmarshal(data, &items)
		}
	}

	return items, nil
}

func (c *LSPClient) GetHover(uri string, line, character int) (*Hover, error) {
	params := HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: line, Character: character},
	}

	response, err := c.SendRequest("textDocument/hover", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("hover error: %s", response.Error.Message)
	}

	if response.Result == nil {
		return nil, nil
	}

	var hover Hover
	if data, err := json.Marshal(response.Result); err == nil {
		if err := json.Unmarshal(data, &hover); err != nil {
			return nil, fmt.Errorf("failed to parse hover result: %w", err)
		}
	}

	return &hover, nil
}

func (c *LSPClient) GetDefinition(uri string, line, character int) ([]Location, error) {
	params := DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: line, Character: character},
	}

	response, err := c.SendRequest("textDocument/definition", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("definition error: %s", response.Error.Message)
	}

	var locations []Location
	if data, err := json.Marshal(response.Result); err == nil {
		// Try to parse as array first
		if err := json.Unmarshal(data, &locations); err != nil {
			// If that fails, try to parse as single Location
			var singleLocation Location
			if err := json.Unmarshal(data, &singleLocation); err == nil {
				locations = []Location{singleLocation}
			}
		}
	}

	return locations, nil
}

func (c *LSPClient) GetReferences(uri string, line, character int, includeDeclaration bool) ([]Location, error) {
	params := ReferenceParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: line, Character: character},
		Context: ReferenceContext{
			IncludeDeclaration: includeDeclaration,
		},
	}

	response, err := c.SendRequest("textDocument/references", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("references error: %s", response.Error.Message)
	}

	var locations []Location
	if data, err := json.Marshal(response.Result); err == nil {
		// References should always be an array, but handle gracefully
		if err := json.Unmarshal(data, &locations); err != nil {
			// If that fails, try to parse as single Location (edge case)
			var singleLocation Location
			if err := json.Unmarshal(data, &singleLocation); err == nil {
				locations = []Location{singleLocation}
			}
		}
	}

	return locations, nil
}

func (c *LSPClient) GetDocumentSymbols(uri string) ([]DocumentSymbol, error) {
	params := DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	}

	response, err := c.SendRequest("textDocument/documentSymbol", params)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("document symbols error: %s", response.Error.Message)
	}

	var symbols []DocumentSymbol
	if data, err := json.Marshal(response.Result); err == nil {
		json.Unmarshal(data, &symbols)
	}

	return symbols, nil
}

func (c *LSPClient) GetDiagnostics(uri string) []Diagnostic {
	c.diagMutex.RLock()
	defer c.diagMutex.RUnlock()

	if diags, ok := c.diagnostics[uri]; ok {
		return diags
	}
	return nil
}

func (c *LSPClient) WaitForDiagnostics(uri string, timeout time.Duration) []Diagnostic {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if diags := c.GetDiagnostics(uri); diags != nil {
			return diags
		}
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// TestFile is a convenience method for quick testing of a single file
// It initializes the server, opens the file, waits for diagnostics, and returns them
func (c *LSPClient) TestFile(filePath string) ([]Diagnostic, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get root directory (parent of file)
	rootPath := filepath.Dir(absPath)
	uri := "file://" + absPath

	// Initialize server
	_, err = c.Initialize(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	// Open document
	err = c.OpenDocument(uri, "kickasm", string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to open document: %w", err)
	}

	// Wait for diagnostics (give server time to analyze)
	time.Sleep(100 * time.Millisecond)
	diagnostics := c.WaitForDiagnostics(uri, 2*time.Second)

	return diagnostics, nil
}