package lsp

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "c64.nvim/internal/log"
)

// maxWorkspaceFiles caps the initial scan. A misconfigured root — a home
// directory, a mounted volume — would otherwise stall the server before it
// answers a single request.
const maxWorkspaceFiles = 2000

// maxImportDepth bounds the recursion when splicing a translation unit.
const maxImportDepth = 64

// sourceExtensions are the files the workspace scan picks up.
var sourceExtensions = map[string]bool{".asm": true, ".kasm": true, ".a": true}

// SourceProvider supplies the current text of a document. The server's
// implementation prefers an open editor buffer over the file on disk, so that
// unsaved edits are analysed; tests supply a map.
type SourceProvider interface {
	Read(uri string) (string, bool)
}

// sourceFile is one parsed document plus the imports it declares.
type sourceFile struct {
	uri     string
	hash    string
	program *Program
	diags   []Diagnostic // diagnostics from lexing and parsing this file alone
	imports []importEdge
	// importOnce is set when the file opens with #importonce, which limits it
	// to a single inclusion per assembly (manual 2.3).
	importOnce bool
}

// importEdge is one #import in a file. Target is empty when the file could not
// be found, in which case the edge only carries the diagnostic position.
type importEdge struct {
	target   string // resolved URI, empty if unresolved
	filename string // as written in the source
	token    Token  // the #import token, for diagnostics
}

// Workspace holds the parsed documents of a project and the import graph
// between them. It knows nothing about diagnostics or the LSP; it answers
// "which files exist", "who imports whom" and "give me a translation unit".
type Workspace struct {
	mu       sync.RWMutex
	root     string // filesystem path of the workspace folder, "" when none
	provider SourceProvider
	files    map[string]*sourceFile
}

// NewWorkspace creates an empty workspace. A root of "" disables scanning, in
// which case every document is its own translation unit — the behaviour the
// server had before imports were followed.
func NewWorkspace(root string, provider SourceProvider) *Workspace {
	return &Workspace{
		root:     root,
		provider: provider,
		files:    make(map[string]*sourceFile),
	}
}

// Root returns the workspace folder path, empty when none was supplied.
func (w *Workspace) Root() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.root
}

// Scan walks the workspace folder and parses every source file it finds.
func (w *Workspace) Scan() {
	root := w.Root()
	if root == "" {
		return
	}

	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entries are skipped, not fatal
		}
		if d.IsDir() {
			name := d.Name()
			if name != "." && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if name == "build" || name == "dist" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if sourceExtensions[strings.ToLower(filepath.Ext(path))] {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		log.Error("Workspace scan of %s failed: %v", root, err)
		return
	}

	if len(found) > maxWorkspaceFiles {
		log.Warn("Workspace %s holds %d source files, more than the limit of %d — imports will not be followed",
			root, len(found), maxWorkspaceFiles)
		return
	}

	for _, path := range found {
		w.Load(PathToURI(path))
	}
	log.Info("Workspace scan of %s: %d source files", root, len(found))
}

// Load parses a document into the workspace, reusing the cached program when
// the text is unchanged. It is safe to call for files outside the scanned root.
func (w *Workspace) Load(uri string) *sourceFile {
	text, ok := w.provider.Read(uri)
	if !ok {
		w.mu.Lock()
		delete(w.files, uri)
		w.mu.Unlock()
		return nil
	}

	hash := calculateContentHash(text)

	w.mu.RLock()
	cached, exists := w.files[uri]
	w.mu.RUnlock()
	if exists && cached.hash == hash {
		return cached
	}

	program, diags := parseFileProgram(uri, text)
	file := &sourceFile{
		uri:        uri,
		hash:       hash,
		program:    program,
		diags:      diags,
		importOnce: declaresImportOnce(program),
	}
	file.imports = w.collectImports(file)

	w.mu.Lock()
	w.files[uri] = file
	w.mu.Unlock()
	return file
}

// Invalidate drops a document so the next Load reparses it.
func (w *Workspace) Invalidate(uri string) {
	w.mu.Lock()
	delete(w.files, uri)
	w.mu.Unlock()
}

// Forget removes a document that no longer exists.
func (w *Workspace) Forget(uri string) {
	w.mu.Lock()
	delete(w.files, uri)
	w.mu.Unlock()
}

// collectImports extracts the #import edges of a file and resolves their paths.
func (w *Workspace) collectImports(file *sourceFile) []importEdge {
	if file.program == nil {
		return nil
	}

	var edges []importEdge
	for _, stmt := range file.program.Statements {
		directive, ok := stmt.(*DirectiveStatement)
		if !ok {
			continue
		}
		name := strings.ToLower(directive.Token.Literal)
		if name != "#import" && name != "#importif" {
			continue
		}
		literal, ok := directive.Value.(*StringLiteral)
		if !ok || literal.Value == "" {
			continue
		}

		edge := importEdge{filename: literal.Value, token: directive.Token}
		if target, found := w.resolveImport(file.uri, literal.Value); found {
			edge.target = target
		}
		edges = append(edges, edge)
	}
	return edges
}

// resolveImport turns the filename of an #import into a URI. Kick Assembler
// searches the current directory and then the -libdir paths (manual 2.3); the
// server resolves relative to the importing file and, failing that, accepts any
// scanned file with the same base name, which covers library layouts without
// needing configuration.
func (w *Workspace) resolveImport(fromURI, filename string) (string, bool) {
	fromPath := URIToPath(fromURI)
	if fromPath != "" {
		candidate := filename
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(filepath.Dir(fromPath), filename)
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return PathToURI(candidate), true
		}
	}

	base := strings.ToLower(filepath.Base(filename))
	w.mu.RLock()
	defer w.mu.RUnlock()
	for uri := range w.files {
		if strings.ToLower(filepath.Base(URIToPath(uri))) == base {
			return uri, true
		}
	}
	return "", false
}

// RootsFor returns the documents whose import closure contains uri. A file that
// nobody imports is its own root, which keeps single file analysis working.
func (w *Workspace) RootsFor(uri string) []string {
	w.mu.RLock()
	candidates := make([]string, 0, len(w.files))
	for candidate := range w.files {
		candidates = append(candidates, candidate)
	}
	w.mu.RUnlock()

	var roots []string
	for _, candidate := range candidates {
		if w.isImported(candidate) {
			continue // not a root
		}
		for _, member := range w.Members(candidate) {
			if member == uri {
				roots = append(roots, candidate)
				break
			}
		}
	}

	if len(roots) == 0 {
		return []string{uri}
	}
	sortStringSlice(roots)
	return roots
}

// PrimaryRootFor picks the translation unit a document's diagnostics come from.
// When several roots import the same file the lowest path wins, so the result
// does not depend on which editor tab happens to be open.
func (w *Workspace) PrimaryRootFor(uri string) string {
	roots := w.RootsFor(uri)
	if len(roots) == 0 {
		return uri
	}
	return roots[0]
}

// isImported reports whether any file imports uri.
func (w *Workspace) isImported(uri string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, file := range w.files {
		for _, edge := range file.imports {
			if edge.target == uri {
				return true
			}
		}
	}
	return false
}

// Members returns the transitive closure of root, root included.
func (w *Workspace) Members(root string) []string {
	seen := map[string]bool{}
	var walk func(uri string, depth int)
	walk = func(uri string, depth int) {
		if depth > maxImportDepth || seen[uri] {
			return
		}
		seen[uri] = true

		w.mu.RLock()
		file := w.files[uri]
		w.mu.RUnlock()
		if file == nil {
			return
		}
		for _, edge := range file.imports {
			if edge.target != "" {
				walk(edge.target, depth+1)
			}
		}
	}
	walk(root, 0)

	members := make([]string, 0, len(seen))
	for uri := range seen {
		members = append(members, uri)
	}
	sortStringSlice(members)
	return members
}

// --- URI helpers ---------------------------------------------------------

// URIToPath converts a file URI to a filesystem path. Anything that is not a
// file URI comes back unchanged, which keeps in-memory test documents working.
func URIToPath(uri string) string {
	if !strings.HasPrefix(uri, "file://") {
		return uri
	}
	path := strings.TrimPrefix(uri, "file://")
	// file:///path keeps its leading slash; file://host/path is not supported.
	if unescaped, err := unescapeURIPath(path); err == nil {
		return unescaped
	}
	return path
}

// PathToURI converts a filesystem path to a file URI.
func PathToURI(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return "file://" + abs
}

// unescapeURIPath decodes the percent escapes a client may send for spaces and
// other characters in a path.
func unescapeURIPath(path string) (string, error) {
	if !strings.Contains(path, "%") {
		return path, nil
	}
	var b strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] == '%' && i+2 < len(path) {
			value, err := hexPair(path[i+1], path[i+2])
			if err != nil {
				return "", err
			}
			b.WriteByte(value)
			i += 2
			continue
		}
		b.WriteByte(path[i])
	}
	return b.String(), nil
}

func hexPair(high, low byte) (byte, error) {
	h, err := hexDigit(high)
	if err != nil {
		return 0, err
	}
	l, err := hexDigit(low)
	if err != nil {
		return 0, err
	}
	return h<<4 | l, nil
}

func hexDigit(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, os.ErrInvalid
}

// SetRoot points the workspace at a folder. Passing "" disables scanning, in
// which case every document stays its own translation unit.
func (w *Workspace) SetRoot(root string) {
	w.mu.Lock()
	w.root = root
	w.mu.Unlock()
}

// UnitSources returns the text of every document in a translation unit, keyed
// by URI. The analyzer needs it to map a token back to its own source line.
func (w *Workspace) UnitSources(root string) map[string]string {
	sources := map[string]string{}
	for _, member := range w.Members(root) {
		if text, ok := w.provider.Read(member); ok {
			sources[member] = text
		}
	}
	return sources
}
