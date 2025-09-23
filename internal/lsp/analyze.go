package lsp

// AnalyzeDocument performs semantic analysis on the document and returns a list of diagnostics.
// It will be expanded in later phases to include more checks.
func AnalyzeDocument(uri string, tree *Scope) []Diagnostic {
	diagnostics := []Diagnostic{}

	// TODO: Implement undefined symbol check (Phase 2)
	// TODO: Implement addressing mode validation (Phase 3)
	// TODO: Implement unused symbol check (Phase 4)

	return diagnostics
}