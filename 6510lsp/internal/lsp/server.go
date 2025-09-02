package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/c64-lsp/6510lsp/internal/log"
)

// documentStore holds the content of opened text documents.
var documentStore = struct {
	sync.RWMutex
	documents map[string]string
}{
	documents: make(map[string]string),
}

// Start initializes and runs the LSP server.
func Start() {
	log.Logger.Println("LSP server starting...")

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		// Read Content-Length header
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Logger.Println("EOF received, exiting.")
				break
			}
			log.Logger.Printf("Error reading header: %v\n", err)
			return
		}

		if len(line) < 16 || line[:16] != "Content-Length: " {
			// Skip empty lines or non-Content-Length headers
			continue
		}

		lengthStr := line[16 : len(line)-2] // Remove "\r\n"
		contentLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			log.Logger.Printf("Error parsing Content-Length: %v\n", err)
			return
		}

		// Read the empty line after headers
		_, err = reader.ReadString('\n')
		if err != nil {
			log.Logger.Printf("Error reading empty line: %v\n", err)
			return
		}

		// Read the JSON payload
		payload := make([]byte, contentLength)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			log.Logger.Printf("Error reading payload: %v\n", err)
			return
		}

		log.Logger.Printf("Received payload: %s\n", string(payload))

		var message map[string]interface{}
		if err := json.Unmarshal(payload, &message); err != nil {
			log.Logger.Printf("Error unmarshaling JSON: %v\n", err)
			continue
		}

		method, ok := message["method"].(string)
		if !ok {
			log.Logger.Println("Method not found or not a string.")
			continue
		}

		switch method {
		case "initialize":
			log.Logger.Println("Handling initialize request.")
			// Construct and send InitializeResult
			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id": message["id"],
				"result": map[string]interface{}{
					"capabilities": map[string]interface{}{
						"textDocumentSync": map[string]interface{}{
							"openClose": true,
							"change": float64(1), // TextDocumentSyncKindFull
						},
					},
				},
			}
			response, _ := json.Marshal(result)
			writeResponse(writer, response)
		case "initialized":
			log.Logger.Println("Handling initialized notification.")
		case "shutdown":
			log.Logger.Println("Handling shutdown request.")
			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id": message["id"],
				"result":  nil,
			}
			response, _ := json.Marshal(result)
			writeResponse(writer, response)
		case "exit":
			log.Logger.Println("Handling exit notification.")
			os.Exit(0)
		case "textDocument/didOpen":
			log.Logger.Println("Handling textDocument/didOpen notification.")
			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						if text, ok := textDocument["text"].(string); ok {
							documentStore.Lock()
							documentStore.documents[uri] = text
							documentStore.Unlock()
							log.Logger.Printf("Stored document %s, length: %d\n", uri, len(text))
						}
					}
				}
			}
		case "textDocument/didChange":
			log.Logger.Println("Handling textDocument/didChange notification.")
			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						if contentChanges, ok := params["contentChanges"].([]interface{}); ok && len(contentChanges) > 0 {
							// For simplicity, we assume full content update (TextDocumentSyncKindFull)
							// The first change entry should contain the full new text
							if change, ok := contentChanges[0].(map[string]interface{}); ok {
								if newText, ok := change["text"].(string); ok {
									documentStore.Lock()
									documentStore.documents[uri] = newText
									documentStore.Unlock()
									log.Logger.Printf("Updated document %s, new length: %d\n", uri, len(newText))
								}
							}
						}
					}
				}
			}
		case "textDocument/didClose":
			log.Logger.Println("Handling textDocument/didClose notification.")
			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						documentStore.Lock()
						delete(documentStore.documents, uri)
						documentStore.Unlock()
						log.Logger.Printf("Removed document %s from store.\n", uri)
					}
				}
			}
		default:
			log.Logger.Printf("Unhandled method: %s\n", method)
		}
	}

	log.Logger.Println("LSP server stopped.")
}

func writeResponse(writer *bufio.Writer, response []byte) {
	fmt.Fprintf(writer, "Content-Length: %d\r\n", len(response))
	fmt.Fprintf(writer, "\r\n")
	writer.Write(response)
	writer.Flush()
	log.Logger.Printf("Sent response: %s\n", string(response))
}