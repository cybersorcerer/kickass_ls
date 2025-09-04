package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
				"id":      message["id"],
				"result": map[string]interface{}{
					"capabilities": map[string]interface{}{
						"textDocumentSync": map[string]interface{}{
							"openClose": true,
							"change":    float64(1), // TextDocumentSyncKindFull
						},
						"hoverProvider": true,
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
				"id":      message["id"],
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
							// Publish diagnostics after opening
							publishDiagnostics(writer, uri, text)
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
									// Publish diagnostics after changing
									publishDiagnostics(writer, uri, newText)
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
						// Clear diagnostics when closing
						publishDiagnostics(writer, uri, "") // Send empty diagnostics to clear
					}
				}
			}
		case "textDocument/hover":
			log.Logger.Println("Handling textDocument/hover request.")

			var responseResult interface{} = nil

			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						if position, ok := params["position"].(map[string]interface{}); ok {
							if lineNum, ok := position["line"].(float64); ok {
								if charNum, ok := position["character"].(float64); ok {
									documentStore.RLock()
									text, docFound := documentStore.documents[uri]
									documentStore.RUnlock()

									if docFound {
										lines := strings.Split(text, "\n")
										if int(lineNum) < len(lines) {
											lineContent := lines[int(lineNum)]
											word := getWordAtPosition(lineContent, int(charNum))
											log.Logger.Printf("Hovering over: %s\n", word)

											description := getOpcodeDescription(strings.ToUpper(word))
											if description != "" {
												responseResult = map[string]interface{}{
													"contents": map[string]interface{}{
														"kind":  "markdown",
														"value": description,
													},
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			finalResponse := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"result":  responseResult,
			}
			responseBytes, _ := json.Marshal(finalResponse)
			writeResponse(writer, responseBytes)
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

// publishDiagnostics analyzes the text and sends diagnostics to the client.
func publishDiagnostics(writer *bufio.Writer, uri string, text string) {
	diagnostics := make([]map[string]interface{}, 0)
	lines := strings.Split(text, "\n")

	// Data structures for label diagnostics
	definedLabels := make(map[string]int) // label -> line number
	usedLabels := make(map[string]bool)
	jumpOpcodes := map[string]bool{
		"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true, "JMP": true, "JSR": true,
	}
	allOpcodes := map[string]bool{
		"ADC": true, "AND": true, "ASL": true, "BCC": true, "BCS": true, "BEQ": true, "BIT": true, "BMI": true, "BNE": true, "BPL": true, "BRK": true, "BVC": true, "BVS": true, "CLC": true, "CLD": true, "CLI": true, "CLV": true, "CMP": true, "CPX": true, "CPY": true, "DEC": true, "DEX": true, "DEY": true, "EOR": true, "INC": true, "INX": true, "INY": true, "JMP": true, "JSR": true, "LDA": true, "LDX": true, "LDY": true, "LSR": true, "NOP": true, "ORA": true, "PHA": true, "PHP": true, "PLA": true, "PLP": true, "ROL": true, "ROR": true, "RTI": true, "RTS": true, "SBC": true, "SEC": true, "SED": true, "SEI": true, "STA": true, "STX": true, "STY": true, "TAX": true, "TAY": true, "TSX": true, "TXA": true, "TXS": true, "TYA": true,
	}

	// First pass: find all defined labels and check for duplicates
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}
		parts := strings.Fields(trimmedLine)
		if len(parts) > 0 {
		
potentialLabel := strings.ToUpper(parts[0])
			if _, isOpcode := allOpcodes[potentialLabel]; !isOpcode {
				label := parts[0]
				if _, exists := definedLabels[label]; exists {
					diagnostics = append(diagnostics, map[string]interface{}{
						"range": map[string]interface{}{
							"start": map[string]interface{}{"line": i, "character": 0},
							"end":   map[string]interface{}{"line": i, "character": len(line)},
						},
						"severity": float64(1), // Error
						"message":  fmt.Sprintf("Duplicate label definition: %s", label),
						"source":   "6510lsp",
					})
				} else {
					definedLabels[label] = i
				}
			}
		}
	}

	// Second pass: find all used labels and check for unknown opcodes
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		parts := strings.Fields(trimmedLine)
		if len(parts) == 0 {
			continue
		}

		var opcode, operand string
		firstWordIsLabel := false
		if _, isLabel := definedLabels[parts[0]]; isLabel {
			firstWordIsLabel = true
		}

		if firstWordIsLabel {
			if len(parts) > 1 {
				opcode = strings.ToUpper(parts[1])
				if len(parts) > 2 {
					operand = parts[2]
				}
			}
		} else {
			opcode = strings.ToUpper(parts[0])
			if len(parts) > 1 {
				operand = parts[1]
			}
		}

		if opcode != "" {
			// Check for used labels
			if _, isJump := jumpOpcodes[opcode]; isJump && operand != "" {
				usedLabels[operand] = true
			}

			// Check for unknown opcodes
			if _, isKnown := allOpcodes[opcode]; !isKnown {
				diagnostics = append(diagnostics, map[string]interface{}{
					"range": map[string]interface{}{
						"start": map[string]interface{}{"line": i, "character": 0},
						"end":   map[string]interface{}{"line": i, "character": len(line)},
					},
					"severity": float64(1), // 1 = Error
					"message":  fmt.Sprintf("Unknown opcode: %s", opcode),
					"source":   "6510lsp",
				})
			}
		}
	}

	// Third pass: check for unused labels
	for label, lineNum := range definedLabels {
		if _, used := usedLabels[label]; !used {
			diagnostics = append(diagnostics, map[string]interface{}{
				"range": map[string]interface{}{
					"start": map[string]interface{}{"line": lineNum, "character": 0},
					"end":   map[string]interface{}{"line": lineNum, "character": len(lines[lineNum])},
				},
				"severity": float64(2), // Warning
				"message":  fmt.Sprintf("Unused label: %s", label),
				"source":   "6510lsp",
			})
		}
	}

	note := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		},
	}

	response, _ := json.Marshal(note)
	writeResponse(writer, response)
}

func getWordAtPosition(lineContent string, charNum int) string {
	if charNum < 0 || charNum >= len(lineContent) {
		return ""
	}
	// Find start of word
	start := charNum
	for start > 0 && isWordChar(rune(lineContent[start-1])) {
		start--
	}

	// Find end of word
	end := charNum
	for end < len(lineContent) && isWordChar(rune(lineContent[end])) {
		end++
	}

	if start >= end {
		return ""
	}

	return lineContent[start:end]
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

func getOpcodeDescription(opcode string) string {
	if opcode == "LDA" {
		return "**LDA LoaD Accumulator**\n\nretrieves a copy from the specified **RAM** or **I/O** address, and stores\nit in the accumulator. The content of the memory location is not affected\nby the operation.\n\n| Addressing mode | Assembler format | Opcode / Bytes |\n| --------------- | ---------------- | -------------- |\n| Immediate       | LDA #nn          | A9 / 2         |\n| Absolute        | LDA nnnn         | AD / 3         |\n| Absolute,X      | LDA nnnn,X       | BD / 3         |\n| Absolute,Y      | LDA nnnn,Y       | B9 / 3         |\n| Zeropage        | LDA nn           | A5 / 2         |\n| Zeropage,X      | LDA nn,X         | B5 / 2         |\n| Indexed-indirect| LDA (nn,X).      | A1 / 2         |\n| Indirect-indexed| LDA (nn),Y       | B1 / 2         |"
	}
	return ""
}
