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

	log "c64.nvim/internal/log"
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

// DocumentSymbol represents a symbol in a text document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           float64          `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Global variable to store mnemonic data
var mnemonics []Mnemonic
var kickassDirectives []KickassDirective
var warnUnusedLabelsEnabled bool

// documentStore holds the content of opened text documents.
var documentStore = struct {
	sync.RWMutex
	documents map[string]string
}{
	documents: make(map[string]string),
}

// symbolStore holds the parsed symbol trees for each document.
var symbolStore = struct {
	sync.RWMutex
	trees map[string]*Scope
}{
	trees: make(map[string]*Scope),
}

func SetWarnUnusedLabels(enabled bool) {
	warnUnusedLabelsEnabled = enabled
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

	// Load kickass directives
	kickassDirectives, err = LoadKickassDirectives("/Users/Ronald.Funk/My_Documents/source/gitlab/c64.nvim")
	if err != nil {
		log.Logger.Printf("Error loading kickass directives: %v\n", err)
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
							"triggerCharacters": []string{" ", "."},
						},
						"definitionProvider":     true,
						"referencesProvider":     true,
						"documentSymbolProvider": true,
						"semanticTokensProvider": map[string]interface{}{
							"legend": map[string]interface{}{
								"tokenTypes": []string{
									"keyword", "variable", "function", "macro", "number", "comment", "string", "operator",
								},
								"tokenModifiers": []string{
									"declaration", "readonly",
								},
							},
							"full": true,
						},
					},
					"serverInfo": map[string]interface{}{
						"name":    "6510lsp",
						"version": "0.5.0",
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
							// Store document content
							documentStore.Lock()
							documentStore.documents[uri] = text
							documentStore.Unlock()
							log.Info("Stored document %s", uri)

							// Parse document and store symbol tree
							symbolTree := ParseDocument(uri, text)
							symbolStore.Lock()
							symbolStore.trees[uri] = symbolTree
							symbolStore.Unlock()
							log.Info("Parsed document and updated symbol store for %s", uri)

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
							if change, ok := contentChanges[0].(map[string]interface{}); ok {
								if newText, ok := change["text"].(string); ok {
									// Update document content
									documentStore.Lock()
									documentStore.documents[uri] = newText
									documentStore.Unlock()
									log.Info("Updated document %s", uri)

									// Re-parse document and update symbol tree
									symbolTree := ParseDocument(uri, newText)
									symbolStore.Lock()
									symbolStore.trees[uri] = symbolTree
									symbolStore.Unlock()
									log.Info("Reparsed document and updated symbol store for %s", uri)

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
						// Remove document from stores
						documentStore.Lock()
						delete(documentStore.documents, uri)
						documentStore.Unlock()

						symbolStore.Lock()
						delete(symbolStore.trees, uri)
						symbolStore.Unlock()

						log.Info("Removed document %s from stores.", uri)

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

									symbolStore.RLock()
									symbolTree, treeFound := symbolStore.trees[uri]
									symbolStore.RUnlock()

									if docFound && treeFound {
										lines := strings.Split(text, "\n")
										if int(lineNum) < len(lines) {
											lineContent := lines[int(lineNum)]
											word := getWordAtPosition(lineContent, int(charNum))
											log.Logger.Printf("Hovering over: %s\n", word)

											// First, check for opcode description
											description := getOpcodeDescription(strings.ToUpper(word))
											if description != "" {
												responseResult = map[string]interface{}{
													"contents": map[string]interface{}{
														"kind":  "markdown",
														"value": description,
													},
												}
											} else {
												// Check for directive description
												directiveDescription := getDirectiveDescription(strings.ToLower(word))
												if directiveDescription != "" {
													responseResult = map[string]interface{}{
														"contents": map[string]interface{}{
															"kind":  "markdown",
															"value": directiveDescription,
														},
													}
												} else {
													// If not an opcode or directive, check for symbol in the symbol tree
													searchSymbol := normalizeLabel(word)
													if symbol, found := symbolTree.FindSymbol(searchSymbol); found {
														var markdown string
														if symbol.Value != "" {
															markdown = fmt.Sprintf("(%s) **%s** = `%s`", symbol.Kind.String(), symbol.Name, symbol.Value)
														} else {
															markdown = fmt.Sprintf("(%s) **%s**", symbol.Kind.String(), symbol.Name)
														}
														responseResult = map[string]interface{}{
															"contents": map[string]interface{}{
																"kind":  "markdown",
																"value": markdown,
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

			completionItems := make([]map[string]interface{}, 0)
			id := message["id"]

			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						if position, ok := params["position"].(map[string]interface{}); ok {
							if lineNum, ok := position["line"].(float64); ok {
								if charNum, ok := position["character"].(float64); ok {
									documentStore.RLock()
									text, docFound := documentStore.documents[uri]
									documentStore.RUnlock()

									symbolStore.RLock()
									symbolTree, treeFound := symbolStore.trees[uri]
									symbolStore.RUnlock()

									if docFound && treeFound {
										lines := strings.Split(text, "\n")
										if int(lineNum) < len(lines) {
											lineContent := lines[int(lineNum)]

											isOperand, wordToComplete := getCompletionContext(lineContent, int(charNum))
											log.Debug("Completion context: isOperand=%v, wordToComplete='%s'", isOperand, wordToComplete)

											if isOperand {
												wordToComplete = strings.TrimPrefix(wordToComplete, "#")
												// Handle namespace completion
												if strings.Contains(wordToComplete, ".") {
													parts := strings.Split(wordToComplete, ".")
													namespaceName := parts[0]
													partialSymbol := ""
													if len(parts) > 1 {
														partialSymbol = parts[1]
													}
													log.Debug("Namespace completion: namespace='%s', partialSymbol='%s'", namespaceName, partialSymbol)

													namespaceScope := symbolTree.FindNamespace(namespaceName)
													if namespaceScope != nil {
														log.Debug("Found namespace scope: %s", namespaceScope.Name)
														for _, symbol := range namespaceScope.Symbols {
															log.Debug("Checking symbol: %s", symbol.Name)
															if strings.HasPrefix(strings.ToUpper(symbol.Name), strings.ToUpper(partialSymbol)) {
																completionItems = append(completionItems, map[string]interface{}{
																	"label":  symbol.Name,
																	"kind":   toCompletionItemKind(symbol.Kind),
																	"detail": symbol.Value,
																})
															}
														}
													}
												} else {
													// Offer global symbols (labels, constants, variables)
													symbols := symbolTree.FindAllVisibleSymbols(int(lineNum))
													for _, symbol := range symbols {
														if strings.HasPrefix(strings.ToUpper(symbol.Name), strings.ToUpper(wordToComplete)) {
															completionItems = append(completionItems, map[string]interface{}{
																"label":  symbol.Name,
																"kind":   toCompletionItemKind(symbol.Kind),
																"detail": symbol.Value,
															})
														}
													}
												}
											} else {
												// Offer mnemonics and directives
												if strings.HasPrefix(wordToComplete, ".") {
													for _, d := range kickassDirectives {
														if strings.HasPrefix(strings.ToLower(d.Directive), strings.ToLower(wordToComplete)) {
															completionItems = append(completionItems, map[string]interface{}{
																"label":  d.Directive,
																"kind":   float64(14), // Keyword
																"detail": d.Description,
															})
														}
													}
												} else {
													for _, m := range mnemonics {
														if strings.HasPrefix(strings.ToUpper(m.Mnemonic), strings.ToUpper(wordToComplete)) {
															completionItems = append(completionItems, map[string]interface{}{
																"label":  m.Mnemonic,
																"kind":   float64(14), // Keyword
																"detail": m.Description,
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
				}
			}

			completionList := map[string]interface{}{
				"isIncomplete": false,
				"items":        completionItems,
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

									symbolStore.RLock()
									symbolTree, treeFound := symbolStore.trees[uri]
									symbolStore.RUnlock()

									if docFound && treeFound {
										lines := strings.Split(text, "\n")
										if int(lineNum) < len(lines) {
											lineContent := lines[int(lineNum)]
											word := getWordAtPosition(lineContent, int(charNum))
											if word != "" {
												searchSymbol := normalizeLabel(word)
												if symbol, found := symbolTree.FindSymbol(searchSymbol); found {
													result = []map[string]interface{}{
														{
															"uri": uri,
															"range": map[string]interface{}{
																"start": map[string]interface{}{
																	"line":      symbol.Position.Line,
																	"character": symbol.Position.Character,
																},
																"end": map[string]interface{}{
																	"line":      symbol.Position.Line,
																	"character": symbol.Position.Character + len(symbol.Name),
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
		case "textDocument/references":
			log.Debug("Handling textDocument/references request.")

			var locations []map[string]interface{}

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
												for i, l := range lines {
													lineWithoutComments := l
													if idx := strings.Index(l, "//"); idx != -1 {
														lineWithoutComments = l[:idx]
													}
													if idx := strings.Index(lineWithoutComments, ";"); idx != -1 {
														lineWithoutComments = lineWithoutComments[:idx]
													}

													if strings.Contains(lineWithoutComments, word) {
														charIndex := strings.Index(l, word)
														locations = append(locations, map[string]interface{}{
															"uri": uri,
															"range": map[string]interface{}{
																"start": map[string]interface{}{"line": i, "character": charIndex},
																"end":   map[string]interface{}{"line": i, "character": charIndex + len(word)},
															},
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
									}

							refResp := map[string]interface{}{
							"jsonrpc": "2.0",
						"id":      message["id"],
						"result":  locations,
					}
					response, _ := json.Marshal(refResp)
					writeResponse(writer, response)

		case "textDocument/documentSymbol":
			log.Debug("Handling textDocument/documentSymbol request.")
			id := message["id"]
			var symbolsResult interface{} = nil

			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						symbolsResult = generateDocumentSymbols(uri)
					}
				}
			}

			finalResponse := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result":  symbolsResult,
			}
			response, _ := json.Marshal(finalResponse)
			writeResponse(writer, response)
		case "textDocument/semanticTokens/full":
			log.Debug("Handling textDocument/semanticTokens/full request.")
			id := message["id"]
			var tokensResult interface{} = nil

			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						documentStore.RLock()
						text, docFound := documentStore.documents[uri]
						documentStore.RUnlock()

						if docFound {
							tokens := generateSemanticTokens(uri, text)
							tokensResult = map[string]interface{}{
								"data": tokens,
							}
						}
					}
				}
			}

			finalResponse := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result":  tokensResult,
			}
			response, _ := json.Marshal(finalResponse)
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

// Hilfsfunktion: Addressing Mode erkennen (angepasst für Branches)
func detectAddressingMode(opcode, operand string) string {
	operand = strings.TrimSpace(operand)
	if operand == "" {
		return "Implied"
	}
	if strings.HasPrefix(operand, "#") {
		return "Immediate"
	}
	if strings.HasPrefix(operand, "(") && strings.HasSuffix(operand, ",Y)") {
		return "Indirect-indexed"
	}
	if strings.HasPrefix(operand, "(") && strings.Contains(operand, ",X)") {
		return "Indexed-indirect"
	}
	if strings.HasSuffix(operand, ",X") {
		if len(operand) <= 5 {
			return "Zeropage,X"
		}
		return "Absolute,X"
	}
	if strings.HasSuffix(operand, ",Y") {
		if len(operand) <= 5 {
			return "Zeropage,Y"
		}
		return "Absolute,Y"
	}
	if strings.HasPrefix(operand, "(") && strings.HasSuffix(operand, ")") {
		return "Indirect"
	}
	// Branch-Befehle (relative Sprünge)
	jumpOpcodes := map[string]bool{
		"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true,
	}
	if jumpOpcodes[strings.ToUpper(opcode)] {
		return "Relative"
	}
	if len(operand) <= 3 {
		return "Zeropage"
	}
	return "Absolute"
}

// publishDiagnostics analyzes the text and sends diagnostics to the client.
func publishDiagnostics(writer *bufio.Writer, uri string, text string) {

	diagnostics := make([]map[string]interface{}, 0)
	lines := strings.Split(text, "\n")

	// Data structures for label diagnostics
	definedLabels := make(map[string]int)
	usedLabels := make(map[string]bool)
	invalidLabelLines := make(map[int]bool)

	jumpOpcodes := map[string]bool{
		"BCC": true, "BCS": true, "BEQ": true, "BMI": true, "BNE": true, "BPL": true, "BVC": true, "BVS": true, "JMP": true, "JSR": true,
	}
	allOpcodes := make(map[string]bool)
	mnemonicMap := make(map[string]Mnemonic)
	for _, m := range mnemonics {
		upper := strings.ToUpper(m.Mnemonic)
		allOpcodes[upper] = true
		mnemonicMap[upper] = m
	}

	kickAssemblerDirectives := map[string]bool{
		".CONST": true, ".VAR": true, ".WORD": true, ".BYTE": true, ".NAMESPACE": true, ".FUNCTION": true, ".MACRO": true, ".LABEL": true, ".PSEUDOCOMMAND": true, ".IF": true, ".FOR": true, ".WHILE": true, ".RETURN": true, "#IMPORT": true, "#INCLUDE": true,
	}

	var currentGlobalLabel string

	// First pass: find all defined labels and check for duplicates
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") || trimmedLine == "{" || trimmedLine == "}" {
			continue
		}
		parts := strings.Fields(trimmedLine)
		if len(parts) > 0 {
			potentialLabel := parts[0]

			// Handle macro invocation with '+'
			if strings.HasPrefix(potentialLabel, "+") {
				continue // It's a macro call, not a definition, skip for this pass.
			}

			// Check for labels ending with ':'
			if strings.HasSuffix(potentialLabel, ":") {
				label := normalizeLabel(potentialLabel)
				originalLabel := label
				if strings.HasPrefix(label, ".") {
					if currentGlobalLabel != "" {
						label = currentGlobalLabel + label
					}
				} else {
					currentGlobalLabel = label
				}
				if _, isOpcode := allOpcodes[label]; !isOpcode {
					if _, exists := definedLabels[label]; exists {

						diagnostics = append(diagnostics, map[string]interface{}{
							"range":    map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end": map[string]interface{}{"line": i, "character": len(line)}},
							"severity": float64(1), // Error
							"message":  fmt.Sprintf("Duplicate label definition: %s", originalLabel),
							"source":   "6510lsp",
						})
					} else {
						definedLabels[label] = i
					}
				}
			} else {
				normalizedFirstWord := normalizeLabel(potentialLabel)
				if _, isOpcode := allOpcodes[normalizedFirstWord]; !isOpcode {
					if _, isDirective := kickAssemblerDirectives[strings.ToUpper(potentialLabel)]; !isDirective {

						diagnostics = append(diagnostics, map[string]interface{}{
							"range":    map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end": map[string]interface{}{"line": i, "character": len(line)}},
							"severity": float64(1), // Error
							"message":  fmt.Sprintf("Invalid syntax: '%s' is not a valid command, directive, or label.", potentialLabel),
							"source":   "6510lsp",
						})
						invalidLabelLines[i] = true
					}
				}
			}
		}
	}

	// Second pass: find all used labels, check for unknown opcodes and addressing mode errors
	currentGlobalLabel = ""
	for i, line := range lines {
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}
		if invalidLabelLines[i] {
			continue
		}
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") || trimmedLine == "{" || trimmedLine == "}" {
			continue
		}

		parts := strings.Fields(trimmedLine)
		if len(parts) == 0 {
			continue
		}

		var opcode, operand string
		firstWord := parts[0]
		if strings.HasPrefix(firstWord, "+") {
			// It's a macro call, for now we don't do advanced diagnostics on it.
			continue
		} else if strings.HasSuffix(firstWord, ":") {
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
			if _, isJump := jumpOpcodes[opcode]; isJump && operand != "" {
				normalizedOperand := normalizeLabel(operand)
				if strings.HasPrefix(normalizedOperand, ".") && currentGlobalLabel != "" {
					normalizedOperand = currentGlobalLabel + normalizedOperand
				}
				usedLabels[normalizedOperand] = true
			}

			if _, isKnown := allOpcodes[opcode]; !isKnown {
				if _, isDirective := kickAssemblerDirectives[opcode]; !isDirective {

					diagnostics = append(diagnostics, map[string]interface{}{
						"range":    map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end": map[string]interface{}{"line": i, "character": len(line)}},
						"severity": float64(1), // 1 = Error
						"message":  fmt.Sprintf("Unknown opcode or directive: %s", opcode),
						"source":   "6510lsp",
					})
				}
			} else {
				mode := detectAddressingMode(opcode, operand)
				mnemonic := mnemonicMap[opcode]
				allowed := false
				for _, am := range mnemonic.AddressingModes {
					if strings.EqualFold(am.AddressingMode, mode) {
						allowed = true
						break
					}
				}
				if !allowed {

					diagnostics = append(diagnostics, map[string]interface{}{
						"range":    map[string]interface{}{"start": map[string]interface{}{"line": i, "character": 0}, "end": map[string]interface{}{"line": i, "character": len(line)}},
						"severity": float64(1), // Error
						"message":  fmt.Sprintf("Invalid addressing mode for %s: %s (detected: %s)", opcode, operand, mode),
						"source":   "6510lsp",
					})
				}
			}
		}
	}

	// Third pass: check for unused labels
	if warnUnusedLabelsEnabled {
		for label, lineNum := range definedLabels {
			if _, used := usedLabels[label]; !used {

				diagnostics = append(diagnostics, map[string]interface{}{
					"range":    map[string]interface{}{"start": map[string]interface{}{"line": lineNum, "character": 0}, "end": map[string]interface{}{"line": lineNum, "character": len(lines[lineNum])}},
					"severity": float64(2), // Warning
					"message":  fmt.Sprintf("Unused label: %s", label),
					"source":   "6510lsp",
				})
			}
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
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.'
}

// normalizeLabel removes a leading '!' and any trailing ':' or '+'-' characters from a label
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

func getDirectiveDescription(directive string) string {
	for _, d := range kickassDirectives {
		if strings.EqualFold(d.Directive, directive) {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("**%s**\n\n%s\n\n", d.Directive, d.Description))

			if len(d.Examples) > 0 {
				sb.WriteString("**Examples:**\n")
				sb.WriteString("```asm\n")
				for _, example := range d.Examples {
					sb.WriteString(example)
					sb.WriteString("\n")
				}
				sb.WriteString("```\n")
			}
			return sb.String()
		}
	}
	return ""
}

func toCompletionItemKind(kind SymbolKind) float64 {
	switch kind {
	case Constant:
		return 21 // Constant
	case Variable:
		return 6 // Variable
	case Label:
		return 10 // Property
	case Function:
		return 3 // Function
	case Macro:
		return 15 // Snippet
	case Namespace:
		return 19 // Module
	default:
		return 1 // Text
	}
}

// getCompletionContext determines if we are completing an operand or a mnemonic
// and returns the word being completed.
func getCompletionContext(line string, char int) (isOperand bool, word string) {
	log.Debug("getCompletionContext line: '%s', char: %d", line, char)
	// Trim whitespace from the beginning of the line
	trimmedLine := strings.TrimSpace(line)
	if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "*") {
		log.Debug("Line is empty or a comment, returning mnemonic context.")
		return false, ""
	}

	// Extract the part of the line before the cursor
	if char < 0 || char > len(line) {
		char = len(line)
	}
	context := line[:char]
	log.Debug("Context: '%s'", context)

	// Tokenize the line up to the cursor
	parts := strings.Fields(context)
	log.Debug("Parts: %v", parts)

	if len(parts) == 0 {
		// This can happen if the line has leading spaces and the cursor is among them.
		// Or if the line is empty.
		log.Debug("No parts found, assuming mnemonic context.")
		return false, ""
	}

	// Check if the cursor is at the end of a word or in the middle of it.
	lastPart := parts[len(parts)-1]
	if !strings.HasSuffix(context, lastPart) {
		// Cursor is likely in whitespace after the last word.
		// Example: "lda #$12 |" (cursor at |)
		// We need to check if the last word was an opcode.
		log.Debug("Cursor is in whitespace after the last word.")
		for _, m := range mnemonics {
			if strings.EqualFold(m.Mnemonic, lastPart) {
				log.Debug("Last word was an opcode, so we are in operand context.")
				return true, "" // We are starting a new operand
			}
		}
		log.Debug("Last word was not an opcode, assuming mnemonic context.")
		return false, "" // Not after an opcode, so it's a new mnemonic/label
	}

	// Cursor is part of the last word.
	// Example: "lda #MAX_SPRI|"
	log.Debug("Cursor is part of the last word: '%s'", lastPart)

	if len(parts) == 1 {
		// Only one word on the line up to the cursor.
		// It could be a label, a mnemonic, or a directive.
		if strings.HasPrefix(lastPart, ".") {
			log.Debug("Single word starts with a dot, assuming directive context.")
			return false, lastPart
		}
		// If it contains operand-like characters, treat as operand.
		if strings.Contains(lastPart, "#") || strings.Contains(lastPart, "$") {
			log.Debug("Single word contains operand characters, treating as operand.")
			return true, lastPart
		}
		// Otherwise, it's a mnemonic or a label definition.
		log.Debug("Single word, assuming mnemonic/label context.")
		return false, lastPart
	}

	// More than one part.
	// The word before the current one is a good indicator.
	prevPart := parts[len(parts)-2]
	log.Debug("Previous part: '%s'", prevPart)

	// Is the previous part an opcode?
	for _, m := range mnemonics {
		if strings.EqualFold(m.Mnemonic, prevPart) {
			log.Debug("Previous part was an opcode, so current is an operand.")
			return true, lastPart
		}
	}

	// If we are here, the context is more complex.
	// e.g., "lda #$12,x" or "jmp namespace.label"
	// A simple check for operand characters in the current word is a good heuristic.
	if strings.HasPrefix(lastPart, ".") {
		log.Debug("Current part starts with a dot, could be a local label or a directive.")
		// if it's the second word, it's likely a directive
		if len(parts) == 2 && !strings.HasSuffix(parts[0], ":") {
			return false, lastPart
		}
		return true, lastPart
	}
	if strings.Contains(lastPart, "#") {
		log.Debug("Current part contains operand characters, treating as operand.")
		return true, lastPart
	}

	// Default case: assume it's a mnemonic if we haven't found an operand context.
	// This could be a new instruction on a line after a label, e.g. "start: lda"
	log.Debug("Defaulting to mnemonic context.")
	return false, lastPart
}
