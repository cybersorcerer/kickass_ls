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
	"unicode"

	log "c64.nvim/internal/log"
)

// CompletionItemKind defines the type of a completion item.
type CompletionItemKind float64

const (
	TextCompletion          CompletionItemKind = 1
	MethodCompletion        CompletionItemKind = 2
	FunctionCompletion      CompletionItemKind = 3
	ConstructorCompletion   CompletionItemKind = 4
	FieldCompletion         CompletionItemKind = 5
	VariableCompletion      CompletionItemKind = 6
	ClassCompletion         CompletionItemKind = 7
	InterfaceCompletion     CompletionItemKind = 8
	ModuleCompletion        CompletionItemKind = 9
	PropertyCompletion      CompletionItemKind = 10
	UnitCompletion          CompletionItemKind = 11
	ValueCompletion         CompletionItemKind = 12
	EnumCompletion          CompletionItemKind = 13
	KeywordCompletion       CompletionItemKind = 14
	SnippetCompletion       CompletionItemKind = 15
	ColorCompletion         CompletionItemKind = 16
	FileCompletion          CompletionItemKind = 17
	ReferenceCompletion     CompletionItemKind = 18
	FolderCompletion        CompletionItemKind = 19
	EnumMemberCompletion    CompletionItemKind = 20
	ConstantCompletion      CompletionItemKind = 21
	StructCompletion        CompletionItemKind = 22
	EventCompletion         CompletionItemKind = 23
	OperatorCompletion      CompletionItemKind = 24
	TypeParameterCompletion CompletionItemKind = 25
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
func Start(mnemonicPath string, kickassPath string) {
	log.Info("LSP server starting...")

	// Load mnemonic data
	err := loadMnemonics(mnemonicPath)
	if err != nil {
		log.Logger.Printf("Error loading mnemonics: %v\n", err)
	}

	// Load kickass directives
	kickassDirectives, err = LoadKickassDirectives(kickassPath)
	if err != nil {
		log.Error("Failed to load kickass directives: %v", err)
	} else {
		log.Info("Successfully loaded %d kickass directives.", len(kickassDirectives))
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
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
			continue
		}

		lengthStr := line[16 : len(line)-2]
		contentLength, err := strconv.Atoi(lengthStr)
		if err != nil {
			log.Logger.Printf("Error parsing Content-Length: %v\n", err)
			return
		}

		_, err = reader.ReadString('\n')
		if err != nil {
			log.Logger.Printf("Error reading empty line: %v\n", err)
			return
		}

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
			result := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"result": map[string]interface{}{
					"capabilities": map[string]interface{}{
						"textDocumentSync": map[string]interface{}{
							"openClose": true,
							"change":    float64(1), // Full sync
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
						"version": "0.7.8", // Version updated
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
							log.Info("Stored document %s", uri)

							// Parse document, build symbol tree, and get diagnostics in one go
							symbolTree, diagnostics := ParseDocument(uri, text)
							symbolStore.Lock()
							symbolStore.trees[uri] = symbolTree
							symbolStore.Unlock()
							log.Info("Parsed document and updated symbol store for %s", uri)

							// Publish diagnostics found during parsing
							publishDiagnostics(writer, uri, diagnostics)
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
									documentStore.Lock()
									documentStore.documents[uri] = newText
									documentStore.Unlock()
									log.Info("Updated document %s", uri)

									// Parse document, build symbol tree, and get diagnostics in one go
									symbolTree, diagnostics := ParseDocument(uri, newText)
									symbolStore.Lock()
									symbolStore.trees[uri] = symbolTree
									symbolStore.Unlock()
									log.Info("Reparsed document and updated symbol store for %s", uri)

									// Publish diagnostics found during parsing
									publishDiagnostics(writer, uri, diagnostics)
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

						symbolStore.Lock()
						delete(symbolStore.trees, uri)
						symbolStore.Unlock()

						log.Info("Removed document %s from stores.", uri)

						publishDiagnostics(writer, uri, []Diagnostic{}) // Clear diagnostics
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

											description := getOpcodeDescription(strings.ToUpper(word))
											if description != "" {
												responseResult = map[string]interface{}{
													"contents": map[string]interface{}{
														"kind":  "markdown",
														"value": description,
													},
												}
											} else {
												directiveDescription := getDirectiveDescription(strings.ToLower(word))
												if directiveDescription != "" {
													responseResult = map[string]interface{}{
														"contents": map[string]interface{}{
															"kind":  "markdown",
															"value": directiveDescription,
														},
													}
												} else {
													searchSymbol := normalizeLabel(word)
													if symbol, found := symbolTree.FindSymbol(searchSymbol); found {
														var markdown string
														if symbol.Signature != "" {
															markdown = fmt.Sprintf("(%s) **%s**", symbol.Kind.String(), symbol.Signature)
														} else if symbol.Value != "" {
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
											completionItems = generateCompletions(symbolTree, int(lineNum), isOperand, wordToComplete)
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
											if symbol, found := symbolTree.FindSymbol(normalizeLabel(word)); found {
												responseResult = map[string]interface{}{
													"uri": uri,
													"range": map[string]interface{}{
														"start": map[string]interface{}{"line": symbol.Position.Line, "character": symbol.Position.Character},
														"end":   map[string]interface{}{"line": symbol.Position.Line, "character": symbol.Position.Character + len(symbol.Name)},
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

		case "textDocument/references":
			log.Debug("Handling textDocument/references request.")
			var responseResult interface{} = nil

			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						if position, ok := params["position"].(map[string]interface{}); ok {
							if lineNum, ok := position["line"].(float64); ok {
								if charNum, ok := position["character"].(float64); ok {
									// Get the context parameter for includeDeclaration
									includeDeclaration := true
									if context, ok := params["context"].(map[string]interface{}); ok {
										if incDec, ok := context["includeDeclaration"].(bool); ok {
											includeDeclaration = incDec
										}
									}

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
												normalizedWord := normalizeLabel(word)

												// First check if the symbol exists
												if symbol, found := symbolTree.FindSymbol(normalizedWord); found {
													// Find all references to this symbol
													references := symbolTree.FindAllReferences(normalizedWord, text, uri)

													// If includeDeclaration is false, filter out the declaration
													if !includeDeclaration && len(references) > 0 {
														filteredReferences := []map[string]interface{}{}
														for _, ref := range references {
															if refRange, ok := ref["range"].(map[string]interface{}); ok {
																if start, ok := refRange["start"].(map[string]interface{}); ok {
																	if refLine, ok := start["line"].(float64); ok {
																		if refChar, ok := start["character"].(float64); ok {
																			// Skip if this is the declaration position
																			if int(refLine) != symbol.Position.Line ||
																				int(refChar) != symbol.Position.Character {
																				filteredReferences = append(filteredReferences, ref)
																			}
																		}
																	}
																}
															}
														}
														responseResult = filteredReferences
													} else {
														responseResult = references
													}

													log.Debug("Found %d references for symbol '%s'", len(references), word)
												} else {
													log.Debug("Symbol '%s' not found for references", word)
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

		case "textDocument/documentSymbol":
			log.Debug("Handling textDocument/documentSymbol request.")
			var responseResult interface{} = nil
			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						responseResult = generateDocumentSymbols(uri)
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

		case "textDocument/semanticTokens/full":
			log.Debug("Handling textDocument/semanticTokens/full request.")
			var responseResult interface{} = nil
			if params, ok := message["params"].(map[string]interface{}); ok {
				if textDocument, ok := params["textDocument"].(map[string]interface{}); ok {
					if uri, ok := textDocument["uri"].(string); ok {
						documentStore.RLock()
						text, _ := documentStore.documents[uri]
						documentStore.RUnlock()
						tokens := generateSemanticTokens(uri, text)
						responseResult = map[string]interface{}{"data": tokens}
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
			log.Warn("Unhandled method: %s", method)
		}
	}
}

// publishDiagnostics sends a list of diagnostics to the client.
func publishDiagnostics(writer *bufio.Writer, uri string, diagnostics []Diagnostic) {
	lspDiagnostics := make([]map[string]interface{}, len(diagnostics))
	for i, d := range diagnostics {
		lspDiagnostics[i] = map[string]interface{}{
			"range":    d.Range,
			"severity": d.Severity,
			"message":  d.Message,
			"source":   d.Source,
		}
	}

	note := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]interface{}{
			"uri":         uri,
			"diagnostics": lspDiagnostics,
		},
	}

	response, _ := json.Marshal(note)
	writeResponse(writer, response)
}

func writeResponse(writer *bufio.Writer, response []byte) {
	log.Logger.Printf("Sending response: %s\n", string(response))
	fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(response))
	writer.Write(response)
	writer.Flush()
}

func loadMnemonics(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &mnemonics)
}

func getOpcodeDescription(mnemonic string) string {
	for _, m := range mnemonics {
		if m.Mnemonic == mnemonic {
			var builder strings.Builder

			// Header with mnemonic name and description
			builder.WriteString(fmt.Sprintf("**%s**\n\n", m.Mnemonic))
			builder.WriteString(fmt.Sprintf("%s\n\n", m.Description))

			// Properly formatted Markdown table with correct newlines
			builder.WriteString("| Opcode | Addressing Mode | Assembler Format | Length | Cycles |\n")
			builder.WriteString("|:------ |:---------------- |:----------------- |:------ |:------ |\n")

			for _, am := range m.AddressingModes {
				// Clean assembler format - remove any backticks that might interfere
				assemblerFormat := strings.ReplaceAll(am.AssemblerFormat, "`", "")
				builder.WriteString(fmt.Sprintf("| `$%s` | %s | `%s` | %d | %s |\n",
					am.Opcode, am.AddressingMode, assemblerFormat, am.Length, am.Cycles))
			}

			// CPU Flags section with proper formatting
			builder.WriteString("\n**CPU Flags Affected:**\n\n")
			if len(m.CPUFlags) > 0 {
				for _, flag := range m.CPUFlags {
					builder.WriteString(fmt.Sprintf("%s\n", flag))
				}
			} else {
				builder.WriteString("None\n")
			}

			return builder.String()
		}
	}
	return ""
}

func getDirectiveDescription(directive string) string {
	for _, d := range kickassDirectives {
		if d.Directive == directive {
			var builder strings.Builder

			// Header with directive name and signature
			builder.WriteString(fmt.Sprintf("**%s**\n\n", strings.ToUpper(d.Directive)))

			// Signature in code block
			if d.Signature != "" {
				builder.WriteString("```kickassembler\n")
				builder.WriteString(d.Signature)
				builder.WriteString("\n```\n\n")
			}

			// Description
			if d.Description != "" {
				builder.WriteString(d.Description)
				builder.WriteString("\n\n")
			}

			// Examples
			if len(d.Examples) > 0 {
				builder.WriteString("**Examples:**\n\n")
				builder.WriteString("```kickassembler\n")
				builder.WriteString(strings.Join(d.Examples, "\n"))
				builder.WriteString("\n```")
			}

			return builder.String()
		}
	}
	return ""
}

func getWordAtPosition(line string, char int) string {
	if char < 0 || char >= len(line) {
		return ""
	}

	start := char
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	end := char
	for end < len(line)-1 && isWordChar(line[end+1]) {
		end++
	}

	return line[start : end+1]
}

func generateCompletions(symbolTree *Scope, lineNum int, isOperand bool, wordToComplete string) []map[string]interface{} {
	items := []map[string]interface{}{}

	if isOperand {
		wordToComplete = strings.TrimPrefix(wordToComplete, "#")
		if strings.Contains(wordToComplete, ".") {
			parts := strings.Split(wordToComplete, ".")
			namespaceName := parts[0]
			partialSymbol := ""
			if len(parts) > 1 {
				partialSymbol = parts[1]
			}
			namespaceScope := symbolTree.FindNamespace(namespaceName)
			if namespaceScope != nil {
				for _, symbol := range namespaceScope.Symbols {
					if strings.HasPrefix(symbol.Name, partialSymbol) {
						item := map[string]interface{}{
							"label":  symbol.Name,
							"kind":   toCompletionItemKind(symbol.Kind),
							"detail": symbol.Value,
						}
						if symbol.Kind == Function {
							item["insertText"] = symbol.Name
						}
						items = append(items, item)
					}
				}
			}
		} else {
			symbols := symbolTree.FindAllVisibleSymbols(lineNum)
			for _, symbol := range symbols {
				if strings.HasPrefix(symbol.Name, wordToComplete) {
					item := map[string]interface{}{
						"label":  symbol.Name,
						"kind":   toCompletionItemKind(symbol.Kind),
						"detail": symbol.Value,
					}
					if symbol.Kind == Function {
						item["insertText"] = symbol.Name
					}
					items = append(items, item)
				}
			}
		}
	} else {
		// Offer directives
		for _, d := range kickassDirectives {
			if strings.HasPrefix(strings.ToLower(d.Directive), strings.ToLower(wordToComplete)) {
				items = append(items, map[string]interface{}{
					"label":         applyCase(wordToComplete, d.Directive),
					"kind":          float64(14), // Keyword
					"detail":        "Kick Assembler Directive",
					"documentation": d.Description,
				})
			}
		}

		// Offer macros, functions, and pseudocommands
		symbols := symbolTree.FindAllVisibleSymbols(lineNum)
		for _, symbol := range symbols {
			if (symbol.Kind == Macro && strings.HasPrefix(symbol.Name, strings.TrimPrefix(wordToComplete, "+"))) ||
				(symbol.Kind == Function && strings.HasPrefix(symbol.Name, wordToComplete)) ||
				(symbol.Kind == PseudoCommand && strings.HasPrefix(symbol.Name, wordToComplete)) {
				label := symbol.Name
				if symbol.Kind == Macro {
					label = "+" + symbol.Name
				}
				item := map[string]interface{}{
					"label":  label,
					"kind":   toCompletionItemKind(symbol.Kind),
					"detail": symbol.Signature,
				}
				items = append(items, item)
			}
		}

		// Offer mnemonics
		for _, m := range mnemonics {
			if strings.HasPrefix(strings.ToUpper(m.Mnemonic), strings.ToUpper(wordToComplete)) {
				items = append(items, map[string]interface{}{
					"label":         applyCase(wordToComplete, m.Mnemonic),
					"kind":          float64(14), // Keyword
					"detail":        "6502/6510 Opcode",
					"documentation": m.Description,
				})
			}
		}
	}
	return items
}

func isMnemonic(word string) bool {
	for _, m := range mnemonics {
		if strings.EqualFold(m.Mnemonic, word) {
			return true
		}
	}
	return false
}

func isDirective(word string) bool {
	for _, d := range kickassDirectives {
		if strings.EqualFold(d.Directive, word) {
			return true
		}
	}
	return false
}

// getCompletionContext determines if we are completing an operand or a mnemonic
// and returns the word being completed.
func getCompletionContext(line string, char int) (isOperand bool, word string) {
	log.Debug("getCompletionContext line: '%s', char: %d", line, char)

	// Extract the part of the line before the cursor
	if char < 0 || char > len(line) {
		char = len(line)
	}
	context := line[:char]
	log.Debug("Context: '%s'", context)

	// Ignore comments
	if idx := strings.Index(context, ";"); idx != -1 {
		context = context[:idx]
	}
	if idx := strings.Index(context, "//"); idx != -1 {
		context = context[:idx]
	}

	trimmedContext := strings.TrimSpace(context)
	if trimmedContext == "" {
		log.Debug("Context is empty or only whitespace, assuming mnemonic context.")
		return false, ""
	}

	parts := strings.Fields(trimmedContext)
	log.Debug("Parts: %v", parts)

	if len(parts) == 0 {
		log.Debug("No parts found, assuming mnemonic context.")
		return false, ""
	}

	// Determine which part the cursor is on.
	// If the context ends with whitespace, the cursor is for a new word.
	if unicode.IsSpace(rune(context[len(context)-1])) {
		verb := parts[0]
		if strings.HasSuffix(verb, ":") { // It's a label
			if len(parts) > 1 {
				verb = parts[1]
			} else {
				// After "label: ", starting a new word (mnemonic)
				log.Debug("Cursor after a label, assuming mnemonic context.")
				return false, ""
			}
		}
		if strings.HasPrefix(verb, ":") { // It's a macro/pseudocommand call with ':' prefix
			log.Debug("Cursor after a ':' prefixed macro/pseudocommand, assuming operand context.")
			return true, ""
		}
		if isMnemonic(verb) || isDirective(verb) {
			log.Debug("Cursor after a mnemonic/directive, assuming operand context.")
			return true, ""
		}
		// e.g. after a constant definition "MAX_SPRITES = 8 |"
		log.Debug("Cursor in whitespace, but not after a known mnemonic/directive. Assuming mnemonic context for a new line.")
		return false, ""
	}

	// Cursor is in the middle of a word.
	wordToComplete := parts[len(parts)-1]
	log.Debug("Word to complete: '%s'", wordToComplete)

	// Is this word the "verb" (mnemonic/directive) or an operand?
	verbIndex := 0
	if len(parts) > 0 && strings.HasSuffix(parts[0], ":") {
		verbIndex = 1
	}

	// If we are completing a word at or before the verb index, it's a mnemonic/directive context.
	if len(parts)-1 <= verbIndex {
		log.Debug("Completing the verb part of the line.")
		return false, wordToComplete
	}

	// We are completing a word after the verb. This is an operand.
	verb := parts[verbIndex]
	if isMnemonic(verb) || isDirective(verb) {
		log.Debug("Completing after a known verb ('%s'), this is an operand.", verb)
		return true, wordToComplete
	}

	// Fallback: if the "verb" is not a known mnemonic/directive (e.g. a macro call),
	// we can assume what follows is an operand.
	if verbIndex < len(parts)-1 {
		log.Debug("Completing after an unknown verb ('%s'), assuming operand.", verb)
		return true, wordToComplete
	}

	// Default fallback
	log.Debug("Defaulting to mnemonic context.")
	return false, wordToComplete
}
func toCompletionItemKind(kind SymbolKind) CompletionItemKind {
	switch kind {
	case Constant:
		return ConstantCompletion
	case Variable:
		return VariableCompletion
	case Label:
		return PropertyCompletion
	case Function:
		return FunctionCompletion
	case Macro:
		return SnippetCompletion
	case PseudoCommand:
		return SnippetCompletion
	case Namespace:
		return ModuleCompletion
	default:
		return TextCompletion
	}
}

func applyCase(original, suggestion string) string {
	// Count lower and upper case letters
	lowerCount := 0
	upperCount := 0
	for _, r := range original {
		if unicode.IsLower(r) {
			lowerCount++
		} else if unicode.IsUpper(r) {
			upperCount++
		}
	}

	// No letters typed (e.g., just "." or "#"), default to lower
	if lowerCount == 0 && upperCount == 0 {
		return strings.ToLower(suggestion)
	}

	// All letters are upper, return upper
	if lowerCount == 0 && upperCount > 0 {
		return strings.ToUpper(suggestion)
	}

	// All letters are lower, return lower
	if upperCount == 0 && lowerCount > 0 {
		return strings.ToLower(suggestion)
	}

	// First letter is upper, rest are lower (or non-letters), return capitalized
	if unicode.IsUpper(rune(original[0])) {
		isCapitalized := true
		for i, r := range original {
			if i > 0 && unicode.IsUpper(r) {
				isCapitalized = false
				break
			}
		}
		if isCapitalized {
			return strings.ToUpper(string(suggestion[0])) + strings.ToLower(suggestion[1:])
		}
	}

	// Mixed case (e.g. lDa) or other weirdness, default to lower
	return strings.ToLower(suggestion)
}
