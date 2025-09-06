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
	Opcode          string `json:"opcode"`
	AddressingMode  string `json:"addressing_mode"`
	AssemblerFormat string `json:"assembler_format"`
	Length          int    `json:"length"`
	Cycles          string `json:"cycles"` // Can be "2", "4*", "2/3/4"
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
	log.Info("LSP server starting...")

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
				log.Info("EOF received, exiting.")
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
			log.Warn("Method not found or not a string.")
			continue
		}

		switch method {
		case "initialize":
			log.Debug("Handling initialize request.")
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
						"definitionProvider": true, // <--- HIER HINZUFÜGEN!
					},
				},
			}
			response, _ := json.Marshal(result)
			writeResponse(writer, response)
		case "initialized":
			log.Debug("Handling initialized notification.")
		case "shutdown":
			log.Debug("Handling shutdown request.")
			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"result":  nil,
			}
			response, _ := json.Marshal(result)
			writeResponse(writer, response)
		case "exit":
			log.Debug("Handling exit notification.")
			os.Exit(0)
		case "textDocument/didOpen":
			log.Debug("Handling textDocument/didOpen notification.")
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
			log.Debug("Handling textDocument/didChange notification.")
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
			log.Debug("Handling textDocument/didClose notification.")
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
			log.Debug("Handling textDocument/hover request.")

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
			log.Debug("Handling textDocument/completion request.")

			// IMMER ein leeres Array initialisieren!
			completionItems := make([]map[string]interface{}, 0)

			completionList := map[string]interface{}{
				"isIncomplete": false,
				"items":        completionItems,
			}

			var id interface{}
			if val, ok := message["id"]; ok {
				id = val
			} else {
				id = 1 // Fallback für buggy Clients
			}

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
											context := ""
											if int(charNum) <= len(lineContent) {
												context = lineContent[:int(charNum)]
											} else {
												context = lineContent
											}
											parts := strings.Fields(context)
											lastPart := ""
											if len(parts) > 0 {
												lastPart = parts[len(parts)-1]
											}

											jumpOpcodes := map[string]bool{
												"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true, "JMP": true, "JSR": true,
											}

											// Label completion, wenn vorher ein Sprungbefehl steht
											if len(parts) > 1 && jumpOpcodes[strings.ToUpper(parts[len(parts)-2])] {
												definedLabels := make(map[string]struct{})
												var currentGlobalLabel string
												for _, l := range lines {
													p := strings.Fields(strings.TrimSpace(l))
													if len(p) > 0 {
														potentialLabel := p[0]
														if strings.HasSuffix(potentialLabel, ":") {
															label := normalizeLabel(potentialLabel)
															if strings.HasPrefix(label, ".") && currentGlobalLabel != "" {
																label = currentGlobalLabel + label
															} else if !strings.HasPrefix(label, ".") {
																currentGlobalLabel = label
															}
															definedLabels[label] = struct{}{}
														}
													}
												}
												for label := range definedLabels {
													completionItems = append(completionItems, map[string]interface{}{
														"label": label,
														"kind":  float64(10), // Property
													})
												}
											} else {
												// Opcode completion
												for _, m := range mnemonics {
													if lastPart == "" || strings.HasPrefix(strings.ToUpper(m.Mnemonic), strings.ToUpper(lastPart)) {
														completionItems = append(completionItems, map[string]interface{}{
															"label": m.Mnemonic,
															"kind":  float64(14), // Keyword
														})
													}
												}
											}
											// Set items for response (IMMER ein Array, nie nil!)
											completionList["items"] = completionItems
										}
									}
								}
							}
						}
					}
				}
			}

			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result":  completionList,
			}
			response, err := json.Marshal(result)
			if err != nil {
				log.Error("Failed to marshal completion response: %v", err)
				return
			}
			log.Debug("Sending completion response: %s", string(response))
			writeResponse(writer, response)
		case "textDocument/definition":
			log.Debug("Handling textDocument/definition request.")

			var result interface{} = nil

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
											if word != "" {
												// Label-Index aufbauen
												type LabelPosition struct {
													Line      int
													Character int
												}
												labelIndex := make(map[string]LabelPosition)
												var currentGlobalLabel string
												for i, l := range lines {
													trimmed := strings.TrimSpace(l)
													if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "*") {
														continue
													}
													parts := strings.Fields(trimmed)
													if len(parts) > 0 {
														potentialLabel := parts[0]
														if strings.HasSuffix(potentialLabel, ":") {
															label := normalizeLabel(potentialLabel)
															if strings.HasPrefix(label, ".") && currentGlobalLabel != "" {
																label = currentGlobalLabel + label
															} else if !strings.HasPrefix(label, ".") {
																currentGlobalLabel = label
															}
															labelIndex[label] = LabelPosition{
																Line:      i,
																Character: strings.Index(l, potentialLabel),
															}
														}
													}
												}
												// Gesuchten Labelnamen normalisieren und ggf. scopen
												searchLabel := normalizeLabel(word)
												if strings.HasPrefix(searchLabel, ".") {
													// Scope: Finde globales Label oberhalb der aktuellen Zeile
													currentGlobalLabel = ""
													for i := int(lineNum); i >= 0; i-- {
														trimmed := strings.TrimSpace(lines[i])
														if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "*") {
															continue
														}
														parts := strings.Fields(trimmed)
														if len(parts) > 0 {
															potentialLabel := parts[0]
															if strings.HasSuffix(potentialLabel, ":") {
																label := normalizeLabel(potentialLabel)
																if !strings.HasPrefix(label, ".") {
																	currentGlobalLabel = label
																	break
																}
															}
														}
													}
													if currentGlobalLabel != "" {
														searchLabel = currentGlobalLabel + searchLabel
													}
												}
												if pos, ok := labelIndex[searchLabel]; ok {
													result = []map[string]interface{}{
														{
															"uri": uri,
															"range": map[string]interface{}{
																"start": map[string]interface{}{
																	"line":      pos.Line,
																	"character": pos.Character,
																},
																"end": map[string]interface{}{
																	"line":      pos.Line,
																	"character": pos.Character + len(word),
																},
															},
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
			}

			defResp := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"result":  result,
			}
			response, _ := json.Marshal(defResp)
			writeResponse(writer, response)
		default:
			log.Logger.Printf("Unhandled method: %s\n", method)
		}
	}

	log.Info("LSP server stopped.")
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
	invalidLabelLines := make(map[int]bool) // <--- NEU: Zeilen mit Label-Fehler merken

	jumpOpcodes := map[string]bool{
		"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true, "JMP": true, "JSR": true,
	}
	allOpcodes := make(map[string]bool)
	for _, m := range mnemonics {
		allOpcodes[strings.ToUpper(m.Mnemonic)] = true
	}

	var currentGlobalLabel string

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
				label := normalizeLabel(potentialLabel)
				originalLabel := label
				if strings.HasPrefix(label, ".") {
					// Local label: scope with current global label
					if currentGlobalLabel != "" {
						label = currentGlobalLabel + label
					}
				} else {
					// Global label: remember for scoping
					currentGlobalLabel = label
				}
				if _, isOpcode := allOpcodes[label]; !isOpcode {
					if _, exists := definedLabels[label]; exists {
						diagnostics = append(diagnostics, map[string]interface{}{
							"range": map[string]interface{}{
								"start": map[string]interface{}{"line": i, "character": 0},
								"end":   map[string]interface{}{"line": i, "character": len(line)},
							},
							"severity": float64(1), // Error
							"message":  fmt.Sprintf("Duplicate label definition: %s", originalLabel),
							"source":   "6510lsp",
						})
					} else {
						definedLabels[label] = i
					}
				}
			} else {
				// If it doesn't end with ':', it's either an opcode or an invalid label definition
				if _, isOpcode := allOpcodes[normalizeLabel(potentialLabel)]; !isOpcode {
					diagnostics = append(diagnostics, map[string]interface{}{
						"range": map[string]interface{}{
							"start": map[string]interface{}{"line": i, "character": 0},
							"end":   map[string]interface{}{"line": i, "character": len(line)},
						},
						"severity": float64(1), // Error
						"message":  fmt.Sprintf("Invalid label definition (missing colon?): %s", potentialLabel),
						"source":   "6510lsp",
					})
					invalidLabelLines[i] = true // <--- Zeile merken!
				}
			}
		}
	}

	// Second pass: find all used labels and check for unknown opcodes
	currentGlobalLabel = ""
	for i, line := range lines {
		if invalidLabelLines[i] {
			continue // <--- Zeile überspringen, wenn schon als Label-Fehler markiert!
		}
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		parts := strings.Fields(trimmedLine)
		if len(parts) == 0 {
			continue
		}

		var opcode, operand string
		firstWord := parts[0]
		isLabelDef := strings.HasSuffix(firstWord, ":")
		if isLabelDef {
			label := normalizeLabel(firstWord)
			if !strings.HasPrefix(label, ".") {
				currentGlobalLabel = label
			}
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
			// Check for used labels (jump targets)
			if _, isJump := jumpOpcodes[opcode]; isJump && operand != "" {
				normalizedOperand := normalizeLabel(operand)
				if strings.HasPrefix(normalizedOperand, ".") && currentGlobalLabel != "" {
					normalizedOperand = currentGlobalLabel + normalizedOperand
				}
				usedLabels[normalizedOperand] = true
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

	log.Info("Successfully loaded mnemonic.json")
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

// normalizeLabel removes a leading '!' and any trailing ':' or '+'/'-' characters from a label
// and converts it to upper case for case-insensitive comparison.
func normalizeLabel(label string) string {
	label = strings.TrimSpace(label)
	label = strings.TrimPrefix(label, "!")
	label = strings.TrimSuffix(label, ":")
	label = strings.TrimRight(label, "+-")
	label = strings.ToUpper(label) // Case-insensitive!
	return label
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
