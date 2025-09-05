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

// Mnemonic represents the structure of a single mnemonic entry in mnemonic.json
type Mnemonic struct {
	Mnemonic        string           `json:"mnemonic"`
	Description     string           `json:"description"`
	AddressingModes []AddressingMode `json:"addressing_modes"`
	CPUFlags        []string         `json:"cpu_flags"`
}

// AddressingMode represents the structure of an addressing mode within a mnemonic
type AddressingMode struct {
	Opcode         string `json:"opcode"`
	AddressingMode string `json:"addressing_mode"`
	AssemblerFormat string `json:"assembler_format"`
	Length         int    `json:"length"`
	Cycles         string `json:"cycles"` // Can be "2", "4*", "2/3/4"
}

// Global variable to store mnemonic data
var mnemonics []Mnemonic

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

	// Load mnemonic data
	err := loadMnemonics("/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim/mnemonic.json")
	if err != nil {
		log.Logger.Printf("Error loading mnemonics: %v\n", err)
		// Depending on severity, you might want to exit or continue without mnemonic data
	}

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
						"completionProvider": map[string]interface{}{
							"resolveProvider":   false,
							"triggerCharacters": []string{" ", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"},
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
		case "textDocument/completion":
			log.Logger.Println("Handling textDocument/completion request.")

			var completionItems []map[string]interface{}

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

											// Basic context analysis
											parts := strings.Fields(lineContent[:int(charNum)])
											if len(parts) > 0 {
												lastPart := parts[len(parts)-1]
												jumpOpcodes := map[string]bool{"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true, "JMP": true, "JSR": true}

												if _, isJump := jumpOpcodes[strings.ToUpper(lastPart)]; isJump {
													// Label completion context
													definedLabels := make(map[string]int)
													allOpcodes := make(map[string]bool)
											for _, m := range mnemonics {
												allOpcodes[strings.ToUpper(m.Mnemonic)] = true
											}
													for i, l := range lines {
														p := strings.Fields(strings.TrimSpace(l))
														if len(p) > 0 {
															if _, isOpcode := allOpcodes[strings.ToUpper(p[0])]; !isOpcode {
																definedLabels[p[0]] = i
															}
														}
													}
													for label := range definedLabels {
														completionItems = append(completionItems, map[string]interface{}{
															"label": label,
															"kind":  float64(10), // 10 = Property (for labels)
														})
													}
												} else {
													// Opcode completion context
													allOpcodes := []string{"ADC", "AND", "ASL", "BCC", "BCS", "BEQ", "BIT", "BMI", "BNE", "BPL", "BRK", "BVC", "BVS", "CLC", "CLD", "CLI", "CLV", "CMP", "CPX", "CPY", "DEC", "DEX", "DEY", "EOR", "INC", "INX", "INY", "JMP", "JSR", "LDA", "LDX", "LDY", "LSR", "NOP", "ORA", "PHA", "PHP", "PLA", "PLP", "ROL", "ROR", "RTI", "RTS", "SBC", "SEC", "SED", "SEI", "STA", "STX", "STY", "TAX", "TAY", "TSX", "TXA", "TXS", "TYA"}
													for _, opcode := range allOpcodes {
														if strings.HasPrefix(opcode, strings.ToUpper(lastPart)) {
															completionItems = append(completionItems, map[string]interface{}{
																"label": opcode,
																"kind":  float64(14), // 14 = Keyword
															})
														}
													}
												}
											} else {
												// Opcode completion context (empty line)
												allOpcodes := []string{"ADC", "AND", "ASL", "BCC", "BCS", "BEQ", "BIT", "BMI", "BNE", "BPL", "BRK", "BVC", "BVS", "CLC", "CLD", "CLI", "CLV", "CMP", "CPX", "CPY", "DEC", "DEX", "DEY", "EOR", "INC", "INX", "INY", "JMP", "JSR", "LDA", "LDX", "LDY", "LSR", "NOP", "ORA", "PHA", "PHP", "PLA", "PLP", "ROL", "ROR", "RTI", "RTS", "SBC", "SEC", "SED", "SEI", "STA", "STX", "STY", "TAX", "TAY", "TSX", "TXA", "TXS", "TYA"}
												for _, opcode := range allOpcodes {
													completionItems = append(completionItems, map[string]interface{}{
														"label": opcode,
														"kind":  float64(14), // 14 = Keyword
													})
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

			completionList := map[string]interface{}{
				"isIncomplete": false,
				"items":        completionItems,
			}

			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"result":  completionList,
			}
			response, _ := json.Marshal(result)
			writeResponse(writer, response)
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
	allOpcodes := make(map[string]bool)
	for _, m := range mnemonics {
		allOpcodes[strings.ToUpper(m.Mnemonic)] = true
	}

	// First pass: find all defined labels and check for duplicates
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}
		parts := strings.Fields(trimmedLine)
		if len(parts) > 0 {
			potentialLabel := parts[0]

			// Check for labels ending with ':'
			if strings.HasSuffix(potentialLabel, ":") {
				label := potentialLabel[:len(potentialLabel)-1] // Remove the colon
				// Handle multi-labels starting with '!'
				if strings.HasPrefix(label, "!") {
					label = label[1:] // Remove the '!' for storage, actual resolution happens later
				}

				if _, isOpcode := allOpcodes[strings.ToUpper(label)]; !isOpcode {
					if _, exists := definedLabels[label]; exists {
						diagnostics = append(diagnostics, map[string]interface{}{
							"range": map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end":   map[string]interface{}{"line": i, "character": len(line)},},
							"severity": float64(1), // Error
							"message":  fmt.Sprintf("Duplicate label definition: %s", label),
							"source":   "6510lsp",
						})
					} else {
						definedLabels[label] = i
					}
				}
			} else {
				// If it doesn't end with ':', it's either an opcode or an invalid label definition
				// For now, we'll treat it as an opcode if it matches one, otherwise it's an unknown opcode.
				// The assumption here is that labels *must* end with ':' in Kick Assembler.
				// If the first word is not an opcode, and doesn't end with ':', it's an error.
				if _, isOpcode := allOpcodes[strings.ToUpper(potentialLabel)]; !isOpcode {
					diagnostics = append(diagnostics, map[string]interface{}{
						"range": map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end":   map[string]interface{}{"line": i, "character": len(line)},},
						"severity": float64(1), // Error
						"message":  fmt.Sprintf("Invalid label definition (missing colon?): %s", potentialLabel),
						"source":   "6510lsp",
					})
				}
			}
		}
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
				// Handle multi-label references (e.g., !label+, !label-)
				if strings.HasPrefix(operand, "!") {
					// Remove '!' and any '+' or '-' suffixes for lookup in definedLabels
					baseLabel := strings.TrimPrefix(operand, "!")
					baseLabel = strings.TrimRightFunc(baseLabel, func(r rune) bool {
						return r == '+' || r == '-'
					})
					usedLabels[baseLabel] = true
				} else {
					usedLabels[operand] = true
				}
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

func loadMnemonics(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read mnemonic.json: %w", err)
	}

	if err := json.Unmarshal(file, &mnemonics); err != nil {
		return fmt.Errorf("failed to unmarshal mnemonic.json: %w", err)
	}

	log.Logger.Println("Successfully loaded mnemonic.json")
	return nil
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
	for _, m := range mnemonics {
		if strings.ToUpper(m.Mnemonic) == opcode {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("**%s**\n\n%s\n\n", m.Mnemonic, m.Description))

			if len(m.AddressingModes) > 0 {
				// Calculate maximum widths for each column
				maxAddrModeLen := len("Addressing mode")
				maxAsmFormatLen := len("Assembler format")
				maxOpcodeLen := len("Opcode")
				maxLengthLen := len("Length")
				maxCyclesLen := len("Cycles")

				for _, am := range m.AddressingModes {
					if len(am.AddressingMode) > maxAddrModeLen {
						maxAddrModeLen = len(am.AddressingMode)
					}
					if len(am.AssemblerFormat) > maxAsmFormatLen {
						maxAsmFormatLen = len(am.AssemblerFormat)
					}
					if len(am.Opcode) > maxOpcodeLen {
						maxOpcodeLen = len(am.Opcode)
					}
					// Length is int, convert to string for length calculation
					if len(fmt.Sprintf("%d", am.Length)) > maxLengthLen {
						maxLengthLen = len(fmt.Sprintf("%d", am.Length))
					}
					if len(fmt.Sprintf("%v", am.Cycles)) > maxCyclesLen {
						maxCyclesLen = len(fmt.Sprintf("%v", am.Cycles))
					}
				}

				// Format header
			sb.WriteString(fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s | %-*s |\n",
					maxAddrModeLen, "Addressing mode",
					maxAsmFormatLen, "Assembler format",
					maxOpcodeLen, "Opcode",
					maxLengthLen, "Length",
					maxCyclesLen, "Cycles"))

				// Format separator
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					strings.Repeat("-", maxAddrModeLen),
					strings.Repeat("-", maxAsmFormatLen),
					strings.Repeat("-", maxOpcodeLen),
					strings.Repeat("-", maxLengthLen),
					strings.Repeat("-", maxCyclesLen)))

				// Format data rows
				for _, am := range m.AddressingModes {
					sb.WriteString(fmt.Sprintf("| %-*s | %-*s | %-*s | %-*d | %-*v |\n",
						maxAddrModeLen, am.AddressingMode,
						maxAsmFormatLen, am.AssemblerFormat,
						maxOpcodeLen, am.Opcode,
						maxLengthLen, am.Length,
						maxCyclesLen, am.Cycles)) // Use %v for interface{} type
				}
			}

			if len(m.CPUFlags) > 0 {
				sb.WriteString("\n**CPU Flags:**\n")
				for _, flag := range m.CPUFlags {
					sb.WriteString(fmt.Sprintf("- %s\n", flag))
				}
			}
			return sb.String()
		}
	}
	return "" // No description found
}


