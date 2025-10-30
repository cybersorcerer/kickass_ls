package lsp

import (
	"encoding/json"

	log "c64.nvim/internal/log"
)

// handleDocumentFormatting handles the textDocument/formatting LSP request
func handleDocumentFormatting(params map[string]interface{}) []interface{} {
	// Extract textDocument URI
	textDocument, ok := params["textDocument"].(map[string]interface{})
	if !ok {
		log.Error("Invalid textDocument in formatting request")
		return nil
	}

	uri, ok := textDocument["uri"].(string)
	if !ok {
		log.Error("Invalid URI in formatting request")
		return nil
	}

	// Get document content
	documentStore.RLock()
	content, exists := documentStore.documents[uri]
	documentStore.RUnlock()

	if !exists {
		log.Warn("Document not found for formatting: %s", uri)
		return nil
	}

	// Get formatting options from request (optional)
	// options, _ := params["options"].(map[string]interface{})
	// We could use these to override config, but for now we use global config

	// Get current formatting config
	configMutex.RLock()
	formattingConfig := lspConfig.Formatting
	configMutex.RUnlock()

	// Format the document
	formattedText, err := FormatDocument(content, formattingConfig)
	if err != nil {
		log.Error("Failed to format document %s: %v", uri, err)
		return nil
	}

	// If no changes, return nil (no edits needed)
	if formattedText == content {
		log.Debug("No formatting changes needed for %s", uri)
		return []interface{}{}
	}

	// Calculate text edits (replace entire document)
	lines := countLines(content)
	textEdit := map[string]interface{}{
		"range": map[string]interface{}{
			"start": map[string]interface{}{
				"line":      0,
				"character": 0,
			},
			"end": map[string]interface{}{
				"line":      lines,
				"character": 0,
			},
		},
		"newText": formattedText,
	}

	log.Debug("Formatting applied to %s: %d lines", uri, lines)

	return []interface{}{textEdit}
}

// handleRangeFormatting handles the textDocument/rangeFormatting LSP request
func handleRangeFormatting(params map[string]interface{}) []interface{} {
	// Extract textDocument URI
	textDocument, ok := params["textDocument"].(map[string]interface{})
	if !ok {
		log.Error("Invalid textDocument in range formatting request")
		return nil
	}

	uri, ok := textDocument["uri"].(string)
	if !ok {
		log.Error("Invalid URI in range formatting request")
		return nil
	}

	// Extract range
	rangeParam, ok := params["range"].(map[string]interface{})
	if !ok {
		log.Error("Invalid range in range formatting request")
		return nil
	}

	startPos, ok := rangeParam["start"].(map[string]interface{})
	if !ok {
		log.Error("Invalid start position in range formatting request")
		return nil
	}

	endPos, ok := rangeParam["end"].(map[string]interface{})
	if !ok {
		log.Error("Invalid end position in range formatting request")
		return nil
	}

	startLine := int(startPos["line"].(float64))
	endLine := int(endPos["line"].(float64))

	// Get document content
	documentStore.RLock()
	content, exists := documentStore.documents[uri]
	documentStore.RUnlock()

	if !exists {
		log.Warn("Document not found for range formatting: %s", uri)
		return nil
	}

	// Get current formatting config
	configMutex.RLock()
	formattingConfig := lspConfig.Formatting
	configMutex.RUnlock()

	// Format the range
	formattedText, err := FormatRange(content, startLine, endLine, formattingConfig)
	if err != nil {
		log.Error("Failed to format range in document %s: %v", uri, err)
		return nil
	}

	// If no changes, return empty array
	if formattedText == content {
		log.Debug("No formatting changes needed for range in %s", uri)
		return []interface{}{}
	}

	// Calculate text edits (replace entire document)
	lines := countLines(content)
	textEdit := map[string]interface{}{
		"range": map[string]interface{}{
			"start": map[string]interface{}{
				"line":      0,
				"character": 0,
			},
			"end": map[string]interface{}{
				"line":      lines,
				"character": 0,
			},
		},
		"newText": formattedText,
	}

	log.Debug("Range formatting applied to %s: lines %d-%d", uri, startLine, endLine)

	return []interface{}{textEdit}
}

// countLines counts the number of lines in a text string
func countLines(text string) int {
	count := 0
	for _, ch := range text {
		if ch == '\n' {
			count++
		}
	}
	return count
}

// formatDocumentResponse creates a properly formatted LSP response for formatting requests
func formatDocumentResponse(id interface{}, result []interface{}) []byte {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Error("Failed to marshal formatting response: %v", err)
		return nil
	}
	return responseBytes
}
