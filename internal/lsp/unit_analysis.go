package lsp

import (
	"os"
	"sort"

	log "c64.nvim/internal/log"
)

// documentSource reads a document for the workspace. An open editor buffer wins
// over the file on disk so that unsaved edits are analysed.
type documentSource struct{}

func (documentSource) Read(uri string) (string, bool) {
	documentStore.RLock()
	text, open := documentStore.documents[uri]
	documentStore.RUnlock()
	if open {
		return text, true
	}

	path := URIToPath(uri)
	if path == "" || path == uri {
		return "", false // not a file URI, and not an open buffer
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// workspace holds the import graph of the project. Its root stays empty until
// initialize supplies one, and with an empty root every document is its own
// translation unit — the behaviour the server had before imports were followed.
var workspace = NewWorkspace("", documentSource{})

// analyzeUnit analyses the translation unit the given document belongs to and
// returns the diagnostics grouped per document. Every member of the unit is a
// key of the result, with an empty slice where there is nothing to report, so
// that callers can clear stale diagnostics.
func analyzeUnit(uri string) (map[string][]Diagnostic, []string) {
	root := workspace.PrimaryRootFor(uri)
	program, members, diagnostics := workspace.AssembleUnit(root)

	scope, definitionDiagnostics := buildScopeFromAST(program, root)
	diagnostics = append(diagnostics, definitionDiagnostics...)

	analyzer := NewSemanticAnalyzerForUnit(scope, workspace.UnitSources(root))
	diagnostics = append(diagnostics, analyzer.Analyze(program)...)

	// One symbol tree serves the whole unit: hover, completion and
	// go-to-definition work from any member.
	symbolStore.Lock()
	for _, member := range members {
		symbolStore.trees[member] = scope
		symbolStore.contexts[member] = analyzer.GetContext()
	}
	symbolStore.Unlock()

	byDocument := make(map[string][]Diagnostic, len(members))
	for _, member := range members {
		byDocument[member] = []Diagnostic{}
	}
	for _, diagnostic := range diagnostics {
		target := diagnostic.URI
		if target == "" {
			target = uri // single document fallback
		}
		byDocument[target] = append(byDocument[target], diagnostic)
	}

	log.Debug("analyzeUnit: root=%s members=%d diagnostics=%d", root, len(members), len(diagnostics))
	return byDocument, members
}

// initWorkspace points the workspace at the folder the client opened and scans
// it. The scan is synchronous: a document opened right after initialize must
// already see the import graph, otherwise its translation unit is guessed
// wrong. Parsing is cheap and maxWorkspaceFiles bounds the work.
func initWorkspace(params map[string]interface{}) {
	root := workspaceRootFrom(params)
	if root == "" {
		log.Info("No workspace folder supplied; each document is analysed on its own")
		return
	}

	workspace.SetRoot(root)
	log.Info("Workspace root: %s", root)
	workspace.Scan()
}

// workspaceRootFrom extracts the project folder from the initialize params,
// accepting the modern workspaceFolders as well as the older rootUri/rootPath.
func workspaceRootFrom(params map[string]interface{}) string {
	if folders, ok := params["workspaceFolders"].([]interface{}); ok && len(folders) > 0 {
		if first, ok := folders[0].(map[string]interface{}); ok {
			if uri, ok := first["uri"].(string); ok && uri != "" {
				return URIToPath(uri)
			}
		}
	}
	if uri, ok := params["rootUri"].(string); ok && uri != "" {
		return URIToPath(uri)
	}
	if path, ok := params["rootPath"].(string); ok && path != "" {
		return path
	}
	return ""
}

// referencesInUnit collects references from every file of the translation unit
// the document belongs to. A symbol declared in an imported file is used across
// the whole unit, so searching the open buffer alone finds only some of them.
func referencesInUnit(scope *Scope, symbolName, uri, text string) []map[string]interface{} {
	sources := workspace.UnitSources(workspace.PrimaryRootFor(uri))
	if len(sources) == 0 {
		return scope.FindAllReferences(symbolName, text, uri)
	}

	// Sorted so the editor lists the hits the same way on every request.
	members := make([]string, 0, len(sources))
	for member := range sources {
		members = append(members, member)
	}
	sort.Strings(members)

	references := []map[string]interface{}{}
	for _, member := range members {
		references = append(references, scope.FindAllReferences(symbolName, sources[member], member)...)
	}
	return references
}
