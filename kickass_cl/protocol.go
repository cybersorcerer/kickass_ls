package main

import "encoding/json"

// LSP Protocol Message Structures

// Base LSP Message
type Message struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Position and Range
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Text Document Identifier
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// Initialize Request/Response
type InitializeParams struct {
	ProcessID             *int                `json:"processId"`
	RootPath              *string             `json:"rootPath"`
	RootURI               *string             `json:"rootUri"`
	InitializationOptions interface{}         `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities  `json:"capabilities"`
	Trace                 *string             `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder   `json:"workspaceFolders,omitempty"`
}

type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Workspace    *WorkspaceClientCapabilities    `json:"workspace,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Completion     *CompletionClientCapabilities `json:"completion,omitempty"`
	Hover          *HoverClientCapabilities      `json:"hover,omitempty"`
	Definition     *DefinitionClientCapabilities `json:"definition,omitempty"`
	References     *ReferencesClientCapabilities `json:"references,omitempty"`
	DocumentSymbol *DocumentSymbolClientCapabilities `json:"documentSymbol,omitempty"`
	SemanticTokens *SemanticTokensClientCapabilities `json:"semanticTokens,omitempty"`
}

type CompletionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	CompletionItem      *struct {
		SnippetSupport          bool     `json:"snippetSupport,omitempty"`
		CommitCharactersSupport bool     `json:"commitCharactersSupport,omitempty"`
		DocumentationFormat     []string `json:"documentationFormat,omitempty"`
	} `json:"completionItem,omitempty"`
}

type HoverClientCapabilities struct {
	DynamicRegistration bool     `json:"dynamicRegistration,omitempty"`
	ContentFormat       []string `json:"contentFormat,omitempty"`
}

type DefinitionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type ReferencesClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type DocumentSymbolClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type SemanticTokensClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

type WorkspaceClientCapabilities struct {
	WorkspaceFolders         bool `json:"workspaceFolders,omitempty"`
	DidChangeConfiguration   bool `json:"didChangeConfiguration,omitempty"`
	DidChangeWatchedFiles    bool `json:"didChangeWatchedFiles,omitempty"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync                 interface{}                    `json:"textDocumentSync,omitempty"`
	CompletionProvider               *CompletionOptions             `json:"completionProvider,omitempty"`
	HoverProvider                    bool                          `json:"hoverProvider,omitempty"`
	DefinitionProvider               bool                          `json:"definitionProvider,omitempty"`
	ReferencesProvider               bool                          `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider           bool                          `json:"documentSymbolProvider,omitempty"`
	SemanticTokensProvider           *SemanticTokensOptions        `json:"semanticTokensProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
}

type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// Document Sync Notifications
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength *int   `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// Completion Request
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position              `json:"position"`
	Context      *CompletionContext    `json:"context,omitempty"`
}

type CompletionContext struct {
	TriggerKind      int     `json:"triggerKind"`
	TriggerCharacter *string `json:"triggerCharacter,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label         string                 `json:"label"`
	Kind          *int                   `json:"kind,omitempty"`
	Detail        *string                `json:"detail,omitempty"`
	Documentation interface{}            `json:"documentation,omitempty"`
	InsertText    *string                `json:"insertText,omitempty"`
	TextEdit      *TextEdit              `json:"textEdit,omitempty"`
	AdditionalTextEdits []TextEdit       `json:"additionalTextEdits,omitempty"`
	SortText      *string                `json:"sortText,omitempty"`
	FilterText    *string                `json:"filterText,omitempty"`
	Data          interface{}            `json:"data,omitempty"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Hover Request
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position              `json:"position"`
}

type Hover struct {
	Contents interface{} `json:"contents"`
	Range    *Range      `json:"range,omitempty"`
}

// Definition Request
type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position              `json:"position"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// References Request
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position              `json:"position"`
	Context      ReferenceContext      `json:"context"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// Document Symbol Request
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         *string          `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Deprecated     *bool            `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Diagnostics
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range              Range                   `json:"range"`
	Severity           *int                    `json:"severity,omitempty"`
	Code               interface{}             `json:"code,omitempty"`
	Source             *string                 `json:"source,omitempty"`
	Message            string                  `json:"message"`
	Tags               []int                   `json:"tags,omitempty"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// Semantic Tokens
type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SemanticTokens struct {
	ResultID *string `json:"resultId,omitempty"`
	Data     []int   `json:"data"`
}

// Document Formatting
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// Signature Help
type SignatureHelpParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *SignatureHelpContext  `json:"context,omitempty"`
}

type SignatureHelpContext struct {
	TriggerKind      int     `json:"triggerKind"`
	TriggerCharacter *string `json:"triggerCharacter,omitempty"`
	IsRetrigger      bool    `json:"isRetrigger"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature *int                   `json:"activeSignature,omitempty"`
	ActiveParameter *int                   `json:"activeParameter,omitempty"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation interface{}            `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label         interface{} `json:"label"` // string or [int, int]
	Documentation interface{} `json:"documentation,omitempty"`
}

// Helper functions for message creation
func NewRequest(method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      1, // Will be set by client
		Method:  method,
		Params:  params,
	}
}

func NewNotification(method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func NewResponse(id interface{}, result interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}