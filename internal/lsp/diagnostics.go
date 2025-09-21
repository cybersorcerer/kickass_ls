package lsp

// DiagnosticSeverity indicates the severity of a diagnostic message.
type DiagnosticSeverity int

const (
	SeverityError   DiagnosticSeverity = 1
	SeverityWarning DiagnosticSeverity = 2
	SeverityInfo    DiagnosticSeverity = 3
	SeverityHint    DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic message, such as a compiler error or warning.
type Diagnostic struct {
	Range    Range
	Severity DiagnosticSeverity
	Source   string
	Message  string
}
