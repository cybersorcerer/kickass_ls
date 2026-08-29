package lsp

import (
	"fmt"

	log "c64.nvim/internal/log"
)

// parseFileProgram lexes and parses one document. It deliberately stops before
// scope building and semantic analysis: those run once per translation unit,
// over the spliced program, not once per file.
func parseFileProgram(uri string, text string) (*Program, []Diagnostic) {
	processorCtx := GetProcessorContext()
	if processorCtx == nil {
		return &Program{Statements: []Statement{}}, []Diagnostic{{
			Range:    Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 0}},
			Severity: SeverityError,
			Source:   "parser",
			Message:  "Internal error: ProcessorContext not initialized",
			URI:      uri,
		}}
	}

	lexer := NewContextAwareLexer(text, uri, processorCtx)
	parser := NewContextAwareParser(lexer, processorCtx)
	program := parser.ParseProgram()
	return program, parser.Errors()
}

// isImportStatement reports whether a statement is an #import or #importif.
func isImportStatement(stmt Statement) bool {
	directive, ok := stmt.(*DirectiveStatement)
	if !ok {
		return false
	}
	switch lowerASCII(directive.Token.Literal) {
	case "#import", "#importif":
		return true
	}
	return false
}

// declaresImportOnce reports whether a program starts its file with
// #importonce, which limits it to a single inclusion per assembly.
func declaresImportOnce(program *Program) bool {
	if program == nil {
		return false
	}
	for _, stmt := range program.Statements {
		if directive, ok := stmt.(*DirectiveStatement); ok {
			if lowerASCII(directive.Token.Literal) == "#importonce" {
				return true
			}
		}
	}
	return false
}

// AssembleUnit builds the translation unit rooted at the given document: the
// statements of every imported file are spliced in at the position of their
// #import, exactly as the assembler would see them. The result is a single
// program that the existing scope builder and analyzer consume unchanged.
//
// It returns the combined program, the documents that contributed to it, and
// the diagnostics produced while parsing and splicing them.
func (w *Workspace) AssembleUnit(root string) (*Program, []string, []Diagnostic) {
	combined := &Program{Statements: []Statement{}}
	diagnostics := []Diagnostic{}
	members := []string{}

	visited := map[string]bool{}  // every file that contributed, for the member list
	included := map[string]bool{} // files already pulled in, for #importonce
	onPath := map[string]bool{}   // the current import chain, for cycle detection

	var splice func(uri string, depth int)
	splice = func(uri string, depth int) {
		if depth > maxImportDepth {
			log.Warn("AssembleUnit: import depth limit reached at %s", uri)
			return
		}

		file := w.Load(uri)
		if file == nil || file.program == nil {
			return
		}

		if !visited[uri] {
			visited[uri] = true
			members = append(members, uri)
			diagnostics = append(diagnostics, file.diags...)
		}
		included[uri] = true

		onPath[uri] = true
		defer delete(onPath, uri)

		importIndex := 0
		for _, stmt := range file.program.Statements {
			if !isImportStatement(stmt) {
				combined.Statements = append(combined.Statements, stmt)
				continue
			}

			if importIndex >= len(file.imports) {
				// The import list and the statements disagree; keep the
				// statement rather than losing it.
				combined.Statements = append(combined.Statements, stmt)
				continue
			}
			edge := file.imports[importIndex]
			importIndex++

			switch {
			case edge.target == "":
				diagnostics = append(diagnostics, importDiagnostic(edge,
					fmt.Sprintf("Imported file %q not found", edge.filename)))

			case onPath[edge.target]:
				diagnostics = append(diagnostics, importDiagnostic(edge,
					fmt.Sprintf("Import cycle: %q is already being imported", edge.filename)))

			default:
				if included[edge.target] {
					// A file may be imported more than once; the assembler then
					// includes its code twice and the resulting duplicate symbol
					// errors are correct. #importonce opts out of that.
					if target := w.Load(edge.target); target != nil && target.importOnce {
						continue
					}
				}
				splice(edge.target, depth+1)
			}
		}
	}

	splice(root, 0)
	return combined, members, diagnostics
}

// importDiagnostic reports a problem with an #import, positioned on the
// directive itself.
func importDiagnostic(edge importEdge, message string) Diagnostic {
	line := edge.token.Line - 1
	column := edge.token.Column - 1
	if line < 0 {
		line = 0
	}
	if column < 0 {
		column = 0
	}
	return Diagnostic{
		Range: Range{
			Start: Position{Line: line, Character: column},
			End:   Position{Line: line, Character: column + len(edge.token.Literal)},
		},
		Severity: SeverityError,
		Source:   "import",
		Message:  message,
		URI:      edge.token.File,
	}
}

// lowerASCII lowercases a directive name without pulling in a full Unicode
// fold; directives are ASCII.
func lowerASCII(s string) string {
	buf := []byte(s)
	for i := range buf {
		if buf[i] >= 'A' && buf[i] <= 'Z' {
			buf[i] += 'a' - 'A'
		}
	}
	return string(buf)
}
