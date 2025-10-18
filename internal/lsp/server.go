package lsp

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
	Type            string           `json:"type"`
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

// C64MemoryMap represents the structure of c64memory.json
type C64MemoryMap struct {
	MemoryMap C64MemoryMapData `json:"memoryMap"`
}

// C64MemoryMapData represents the memory map metadata and regions
type C64MemoryMapData struct {
	Version string                     `json:"version"`
	Source  string                     `json:"source"`
	Regions map[string]C64MemoryRegion `json:"regions"`
}

// C64MemoryRegion represents a single memory address or region
type C64MemoryRegion struct {
	Name        string            `json:"name"`
	Category    string            `json:"category"` // "VIC-II", "SID", "CIA", "System"
	Type        string            `json:"type"`     // "register", "ram", "rom"
	Size        int               `json:"size"`     // Size in bytes
	Description string            `json:"description"`
	Access      string            `json:"access"`    // "read", "write", "read/write"
	BitFields   map[string]string `json:"bitFields"` // "0-3": "Color value"
	Values      map[string]string `json:"values"`    // "0x00": "Black"
	Examples    []string          `json:"examples"`
	Related     []string          `json:"related"` // Related addresses
	Tips        []string          `json:"tips"`    // Programming tips
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
var c64MemoryMap C64MemoryMap
var kickassDirectives []KickassDirective
var builtinFunctions []BuiltinFunction
var builtinConstants []BuiltinConstant
var warnUnusedLabelsEnabled bool

// LSPConfiguration holds all configurable LSP settings
type LSPConfiguration struct {
	// General Analysis Settings
	WarnUnusedLabels bool `json:"warnUnusedLabels"`

	// 6502-Specific Features
	ZeroPageOptimization struct {
		Enabled   bool `json:"enabled"`
		ShowHints bool `json:"showHints"`
	} `json:"zeroPageOptimization"`

	BranchDistanceValidation struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	} `json:"branchDistanceValidation"`

	IllegalOpcodeDetection struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	} `json:"illegalOpcodeDetection"`

	HardwareBugDetection struct {
		Enabled        bool `json:"enabled"`
		ShowWarnings   bool `json:"showWarnings"`
		JMPIndirectBug bool `json:"jmpIndirectBug"`
	} `json:"hardwareBugDetection"`

	MemoryLayoutAnalysis struct {
		Enabled              bool `json:"enabled"`
		ShowIOAccess         bool `json:"showIOAccess"`
		ShowStackWarnings    bool `json:"showStackWarnings"`
		ShowROMWriteWarnings bool `json:"showROMWriteWarnings"`
	} `json:"memoryLayoutAnalysis"`

	MagicNumberDetection struct {
		Enabled      bool `json:"enabled"`
		ShowHints    bool `json:"showHints"`
		C64Addresses bool `json:"c64Addresses"`
	} `json:"magicNumberDetection"`

	DeadCodeDetection struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	} `json:"deadCodeDetection"`

	StyleGuideEnforcement struct {
		Enabled            bool `json:"enabled"`
		ShowHints          bool `json:"showHints"`
		UpperCaseConstants bool `json:"upperCaseConstants"`
		DescriptiveLabels  bool `json:"descriptiveLabels"`
	} `json:"styleGuideEnforcement"`

	// Parser Feature Flags for Context-Aware Redesign
	ParserFeatureFlags struct {
		UseContextAware    bool `json:"useContextAware"`    // Enable new context-aware parser
		FallbackToOld      bool `json:"fallbackToOld"`      // Fallback to old parser on errors
		DebugMode          bool `json:"debugMode"`          // Enable parser debug logging
		EnableExperimental bool `json:"enableExperimental"` // Enable experimental features

		// Feature-specific flags
		ContextAwareLexer  bool `json:"contextAwareLexer"`  // Use new lexer with state management
		EnhancedAST        bool `json:"enhancedAST"`        // Use enhanced AST nodes
		SmartCompletion    bool `json:"smartCompletion"`    // Context-aware completion
		SemanticValidation bool `json:"semanticValidation"` // Enhanced semantic validation
		PerformanceMode    bool `json:"performanceMode"`    // Optimize for performance
	} `json:"parserFeatureFlags"`
}

// lspConfig holds the current LSP configuration
var lspConfig = &LSPConfiguration{
	WarnUnusedLabels: true,
	ZeroPageOptimization: struct {
		Enabled   bool `json:"enabled"`
		ShowHints bool `json:"showHints"`
	}{
		Enabled:   true,
		ShowHints: true,
	},
	BranchDistanceValidation: struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	}{
		Enabled:      true,
		ShowWarnings: true,
	},
	IllegalOpcodeDetection: struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	}{
		Enabled:      true,
		ShowWarnings: true,
	},
	HardwareBugDetection: struct {
		Enabled        bool `json:"enabled"`
		ShowWarnings   bool `json:"showWarnings"`
		JMPIndirectBug bool `json:"jmpIndirectBug"`
	}{
		Enabled:        true,
		ShowWarnings:   true,
		JMPIndirectBug: true,
	},
	MemoryLayoutAnalysis: struct {
		Enabled              bool `json:"enabled"`
		ShowIOAccess         bool `json:"showIOAccess"`
		ShowStackWarnings    bool `json:"showStackWarnings"`
		ShowROMWriteWarnings bool `json:"showROMWriteWarnings"`
	}{
		Enabled:              true,
		ShowIOAccess:         true,
		ShowStackWarnings:    true,
		ShowROMWriteWarnings: true,
	},
	MagicNumberDetection: struct {
		Enabled      bool `json:"enabled"`
		ShowHints    bool `json:"showHints"`
		C64Addresses bool `json:"c64Addresses"`
	}{
		Enabled:      true,
		ShowHints:    true,
		C64Addresses: true,
	},
	DeadCodeDetection: struct {
		Enabled      bool `json:"enabled"`
		ShowWarnings bool `json:"showWarnings"`
	}{
		Enabled:      true,
		ShowWarnings: true,
	},
	StyleGuideEnforcement: struct {
		Enabled            bool `json:"enabled"`
		ShowHints          bool `json:"showHints"`
		UpperCaseConstants bool `json:"upperCaseConstants"`
		DescriptiveLabels  bool `json:"descriptiveLabels"`
	}{
		Enabled:            true,
		ShowHints:          true,
		UpperCaseConstants: true,
		DescriptiveLabels:  true,
	},

	// Parser Feature Flags - Use context-aware parser by default
	ParserFeatureFlags: struct {
		UseContextAware    bool `json:"useContextAware"`
		FallbackToOld      bool `json:"fallbackToOld"`
		DebugMode          bool `json:"debugMode"`
		EnableExperimental bool `json:"enableExperimental"`
		ContextAwareLexer  bool `json:"contextAwareLexer"`
		EnhancedAST        bool `json:"enhancedAST"`
		SmartCompletion    bool `json:"smartCompletion"`
		SemanticValidation bool `json:"semanticValidation"`
		PerformanceMode    bool `json:"performanceMode"`
	}{
		UseContextAware:    true,  // Use context-aware parser by default
		FallbackToOld:      false, // No fallback to old parser
		DebugMode:          false, // Disable debug by default
		EnableExperimental: true,  // Enable experimental features
		ContextAwareLexer:  true,  // Use context-aware lexer by default
		EnhancedAST:        true,  // Use enhanced AST by default
		SmartCompletion:    true,  // Use smart completion by default
		SemanticValidation: true,  // Use semantic validation by default
		PerformanceMode:    false, // Disable performance mode for better features
	},
}

// configMutex protects access to lspConfig
var configMutex sync.RWMutex

// GetLSPConfig returns a copy of the current configuration
func GetLSPConfig() LSPConfiguration {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return *lspConfig
}

// UpdateLSPConfig updates the configuration from a map
func UpdateLSPConfig(settings map[string]interface{}) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Helper function to safely get bool from map
	getBool := func(m map[string]interface{}, key string, defaultValue bool) bool {
		if val, ok := m[key]; ok {
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return defaultValue
	}

	// Helper function to safely get nested object
	getObject := func(m map[string]interface{}, key string) map[string]interface{} {
		if val, ok := m[key]; ok {
			if obj, ok := val.(map[string]interface{}); ok {
				return obj
			}
		}
		return make(map[string]interface{})
	}

	// Update general settings
	lspConfig.WarnUnusedLabels = getBool(settings, "warnUnusedLabels", lspConfig.WarnUnusedLabels)

	// Update zero page optimization
	if zpo := getObject(settings, "zeroPageOptimization"); len(zpo) > 0 {
		lspConfig.ZeroPageOptimization.Enabled = getBool(zpo, "enabled", lspConfig.ZeroPageOptimization.Enabled)
		lspConfig.ZeroPageOptimization.ShowHints = getBool(zpo, "showHints", lspConfig.ZeroPageOptimization.ShowHints)
	}

	// Update branch distance validation
	if bdv := getObject(settings, "branchDistanceValidation"); len(bdv) > 0 {
		lspConfig.BranchDistanceValidation.Enabled = getBool(bdv, "enabled", lspConfig.BranchDistanceValidation.Enabled)
		lspConfig.BranchDistanceValidation.ShowWarnings = getBool(bdv, "showWarnings", lspConfig.BranchDistanceValidation.ShowWarnings)
	}

	// Update illegal opcode detection
	if iod := getObject(settings, "illegalOpcodeDetection"); len(iod) > 0 {
		lspConfig.IllegalOpcodeDetection.Enabled = getBool(iod, "enabled", lspConfig.IllegalOpcodeDetection.Enabled)
		lspConfig.IllegalOpcodeDetection.ShowWarnings = getBool(iod, "showWarnings", lspConfig.IllegalOpcodeDetection.ShowWarnings)
	}

	// Update hardware bug detection
	if hbd := getObject(settings, "hardwareBugDetection"); len(hbd) > 0 {
		lspConfig.HardwareBugDetection.Enabled = getBool(hbd, "enabled", lspConfig.HardwareBugDetection.Enabled)
		lspConfig.HardwareBugDetection.ShowWarnings = getBool(hbd, "showWarnings", lspConfig.HardwareBugDetection.ShowWarnings)
		lspConfig.HardwareBugDetection.JMPIndirectBug = getBool(hbd, "jmpIndirectBug", lspConfig.HardwareBugDetection.JMPIndirectBug)
	}

	// Update memory layout analysis
	if mla := getObject(settings, "memoryLayoutAnalysis"); len(mla) > 0 {
		lspConfig.MemoryLayoutAnalysis.Enabled = getBool(mla, "enabled", lspConfig.MemoryLayoutAnalysis.Enabled)
		lspConfig.MemoryLayoutAnalysis.ShowIOAccess = getBool(mla, "showIOAccess", lspConfig.MemoryLayoutAnalysis.ShowIOAccess)
		lspConfig.MemoryLayoutAnalysis.ShowStackWarnings = getBool(mla, "showStackWarnings", lspConfig.MemoryLayoutAnalysis.ShowStackWarnings)
		lspConfig.MemoryLayoutAnalysis.ShowROMWriteWarnings = getBool(mla, "showROMWriteWarnings", lspConfig.MemoryLayoutAnalysis.ShowROMWriteWarnings)
	}

	// Update magic number detection
	if mnd := getObject(settings, "magicNumberDetection"); len(mnd) > 0 {
		lspConfig.MagicNumberDetection.Enabled = getBool(mnd, "enabled", lspConfig.MagicNumberDetection.Enabled)
		lspConfig.MagicNumberDetection.ShowHints = getBool(mnd, "showHints", lspConfig.MagicNumberDetection.ShowHints)
		lspConfig.MagicNumberDetection.C64Addresses = getBool(mnd, "c64Addresses", lspConfig.MagicNumberDetection.C64Addresses)
	}

	// Update dead code detection
	if dcd := getObject(settings, "deadCodeDetection"); len(dcd) > 0 {
		lspConfig.DeadCodeDetection.Enabled = getBool(dcd, "enabled", lspConfig.DeadCodeDetection.Enabled)
		lspConfig.DeadCodeDetection.ShowWarnings = getBool(dcd, "showWarnings", lspConfig.DeadCodeDetection.ShowWarnings)
	}

	// Update style guide enforcement
	if sge := getObject(settings, "styleGuideEnforcement"); len(sge) > 0 {
		lspConfig.StyleGuideEnforcement.Enabled = getBool(sge, "enabled", lspConfig.StyleGuideEnforcement.Enabled)
		lspConfig.StyleGuideEnforcement.ShowHints = getBool(sge, "showHints", lspConfig.StyleGuideEnforcement.ShowHints)
		lspConfig.StyleGuideEnforcement.UpperCaseConstants = getBool(sge, "upperCaseConstants", lspConfig.StyleGuideEnforcement.UpperCaseConstants)
		lspConfig.StyleGuideEnforcement.DescriptiveLabels = getBool(sge, "descriptiveLabels", lspConfig.StyleGuideEnforcement.DescriptiveLabels)
	}

	// Update parser feature flags
	if pff := getObject(settings, "parserFeatureFlags"); len(pff) > 0 {
		// Main feature flags
		oldUseContextAware := lspConfig.ParserFeatureFlags.UseContextAware
		lspConfig.ParserFeatureFlags.UseContextAware = getBool(pff, "useContextAware", lspConfig.ParserFeatureFlags.UseContextAware)
		lspConfig.ParserFeatureFlags.FallbackToOld = getBool(pff, "fallbackToOld", lspConfig.ParserFeatureFlags.FallbackToOld)
		lspConfig.ParserFeatureFlags.DebugMode = getBool(pff, "debugMode", lspConfig.ParserFeatureFlags.DebugMode)
		lspConfig.ParserFeatureFlags.EnableExperimental = getBool(pff, "enableExperimental", lspConfig.ParserFeatureFlags.EnableExperimental)

		// Feature-specific flags
		lspConfig.ParserFeatureFlags.ContextAwareLexer = getBool(pff, "contextAwareLexer", lspConfig.ParserFeatureFlags.ContextAwareLexer)
		lspConfig.ParserFeatureFlags.EnhancedAST = getBool(pff, "enhancedAST", lspConfig.ParserFeatureFlags.EnhancedAST)
		lspConfig.ParserFeatureFlags.SmartCompletion = getBool(pff, "smartCompletion", lspConfig.ParserFeatureFlags.SmartCompletion)
		lspConfig.ParserFeatureFlags.SemanticValidation = getBool(pff, "semanticValidation", lspConfig.ParserFeatureFlags.SemanticValidation)
		lspConfig.ParserFeatureFlags.PerformanceMode = getBool(pff, "performanceMode", lspConfig.ParserFeatureFlags.PerformanceMode)

		// Log significant parser mode changes
		if oldUseContextAware != lspConfig.ParserFeatureFlags.UseContextAware {
			if lspConfig.ParserFeatureFlags.UseContextAware {
				log.Info("Switched to context-aware parser (experimental)")
			} else {
				log.Info("Switched to legacy parser")
			}
		}

		if lspConfig.ParserFeatureFlags.DebugMode {
			log.Debug("Parser debug mode enabled")
		}
	}

	log.Debug("LSP Configuration updated")
}

// Feature flag helper functions for context-aware parser
func IsContextAwareParserEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.UseContextAware
}

func ShouldFallbackToOldParser() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.FallbackToOld
}

func IsParserDebugModeEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.DebugMode
}

func IsContextAwareLexerEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.ContextAwareLexer
}

func IsEnhancedASTEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.EnhancedAST
}

func IsSmartCompletionEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.SmartCompletion
}

func IsSemanticValidationEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.SemanticValidation
}

func IsPerformanceModeEnabled() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.PerformanceMode
}

// Combined function to check if we should use the new parser infrastructure
func ShouldUseNewParser() bool {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lspConfig.ParserFeatureFlags.UseContextAware ||
		lspConfig.ParserFeatureFlags.ContextAwareLexer ||
		lspConfig.ParserFeatureFlags.EnhancedAST
}

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

// DocumentCache represents cached parsing results for a document
type DocumentCache struct {
	Content      string
	ContentHash  string
	Scope        *Scope
	Diagnostics  []Diagnostic
	LastModified time.Time
}

// parseCache holds cached parsing results to avoid re-parsing unchanged documents
var parseCache = struct {
	sync.RWMutex
	cache map[string]*DocumentCache
}{
	cache: make(map[string]*DocumentCache),
}

// calculateContentHash creates a SHA256 hash of the document content
func calculateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ParseDocumentCached parses a document with caching for unchanged content
func ParseDocumentCached(uri string, text string) (*Scope, []Diagnostic) {
	contentHash := calculateContentHash(text)

	// Check cache first
	parseCache.RLock()
	if cached, exists := parseCache.cache[uri]; exists {
		if cached.ContentHash == contentHash {
			// Cache hit - return cached results
			parseCache.RUnlock()
			log.Debug("Cache hit for document %s", uri)
			return cached.Scope, cached.Diagnostics
		}
	}
	parseCache.RUnlock()

	// Cache miss - parse document
	log.Debug("Cache miss for document %s - parsing", uri)
	scope, diagnostics := ParseDocument(uri, text)

	// Update cache
	parseCache.Lock()
	parseCache.cache[uri] = &DocumentCache{
		Content:      text,
		ContentHash:  contentHash,
		Scope:        scope,
		Diagnostics:  diagnostics,
		LastModified: time.Now(),
	}
	parseCache.Unlock()

	return scope, diagnostics
}

// ClearParseCache removes a document from the parse cache
func ClearParseCache(uri string) {
	parseCache.Lock()
	delete(parseCache.cache, uri)
	parseCache.Unlock()
	log.Debug("Cleared parse cache for document %s", uri)
}

// AnalysisJob represents a parsing/analysis job
type AnalysisJob struct {
	URI     string
	Content string
	Writer  *bufio.Writer
	IsOpen  bool // true for didOpen, false for didChange
}

// analysisQueue processes parsing jobs asynchronously
var analysisQueue = make(chan AnalysisJob, 10)
var analysisWorkerStarted = false

// startAnalysisWorker starts the background worker for processing analysis jobs
func startAnalysisWorker() {
	if analysisWorkerStarted {
		return
	}
	analysisWorkerStarted = true

	go func() {
		log.Debug("Analysis worker started")
		for job := range analysisQueue {
			processAnalysisJob(job)
		}
	}()
}

// processAnalysisJob processes a single analysis job
func processAnalysisJob(job AnalysisJob) {
	// Parse document with caching
	symbolTree, diagnostics := ParseDocumentCached(job.URI, job.Content)

	// Update symbol store
	symbolStore.Lock()
	symbolStore.trees[job.URI] = symbolTree
	symbolStore.Unlock()

	if job.IsOpen {
		log.Info("Parsed document and updated symbol store for %s", job.URI)
	} else {
		log.Info("Reparsed document and updated symbol store for %s", job.URI)
	}

	// Publish diagnostics
	publishDiagnostics(job.Writer, job.URI, diagnostics)

	// Note: We don't return diagnostics to the pool here because they may be
	// referenced in the cache. The pool is mainly for temporary diagnostic slices
	// during analysis. The cache will eventually be evicted and GC will clean up.
}

// submitAnalysisJob submits a job to the analysis queue (non-blocking)
func submitAnalysisJob(uri, content string, writer *bufio.Writer, isOpen bool) {
	job := AnalysisJob{
		URI:     uri,
		Content: content,
		Writer:  writer,
		IsOpen:  isOpen,
	}

	select {
	case analysisQueue <- job:
		// Job queued successfully
		log.Debug("Queued analysis job for %s", uri)
	default:
		// Queue is full - process synchronously as fallback
		log.Debug("Analysis queue full, processing %s synchronously", uri)
		processAnalysisJob(job)
	}
}

// DiagnosticPool manages a pool of diagnostic slices to reduce allocations
type DiagnosticPool struct {
	pool sync.Pool
}

// diagnosticPool is the global pool for diagnostic slices
var diagnosticPool = &DiagnosticPool{
	pool: sync.Pool{
		New: func() interface{} {
			// Pre-allocate slice with reasonable capacity
			return make([]Diagnostic, 0, 32)
		},
	},
}

// Get retrieves a diagnostic slice from the pool
func (dp *DiagnosticPool) Get() []Diagnostic {
	return dp.pool.Get().([]Diagnostic)
}

// Put returns a diagnostic slice to the pool
func (dp *DiagnosticPool) Put(diagnostics []Diagnostic) {
	// Reset slice but keep underlying array if capacity is reasonable
	if cap(diagnostics) < 128 { // Don't pool overly large slices
		diagnostics = diagnostics[:0] // Reset length to 0
		dp.pool.Put(diagnostics)
	}
}

// GetPooledDiagnostics gets a diagnostic slice from the pool
func GetPooledDiagnostics() []Diagnostic {
	return diagnosticPool.Get()
}

// ReturnPooledDiagnostics returns a diagnostic slice to the pool
func ReturnPooledDiagnostics(diagnostics []Diagnostic) {
	diagnosticPool.Put(diagnostics)
}

func SetWarnUnusedLabels(enabled bool) {
	warnUnusedLabelsEnabled = enabled
}

// Start initializes and runs the LSP server.
func Start() {
	log.Info("LSP server starting...")

	// Check for config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("Failed to get user home directory: %v", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".config", "kickass_ls")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Error("Configuration directory %s does not exist. Please create it and install the required JSON files.", configDir)
		os.Exit(1)
	}

	log.Info("Using configuration directory: %s", configDir)

	// Start the analysis worker
	startAnalysisWorker()

	// Load JSON Source of Truth files in parallel from config directory
	var wg sync.WaitGroup
	wg.Add(3)

	// Load mnemonic data
	go func() {
		defer wg.Done()
		mnemonicPath := filepath.Join(configDir, "mnemonic.json")
		err := loadMnemonics(mnemonicPath)
		if err != nil {
			log.Error("Error loading mnemonics from %s: %v", mnemonicPath, err)
		} else {
			log.Info("Successfully loaded mnemonics from %s", mnemonicPath)
		}

		// Set mnemonic.json path for lexer
		SetMnemonicJSONPath(mnemonicPath)
	}()

	// Load C64 memory map data
	go func() {
		defer wg.Done()
		c64MemoryPath := filepath.Join(configDir, "c64memory.json")
		err := loadC64MemoryMap(c64MemoryPath)
		if err != nil {
			log.Error("Could not load C64 memory map from %s: %v", c64MemoryPath, err)
			log.Error("Memory address hover information will be limited.")
		} else {
			log.Info("Successfully loaded C64 memory map with %d regions from %s", len(c64MemoryMap.MemoryMap.Regions), c64MemoryPath)
		}
	}()

	// Load kickass data
	go func() {
		defer wg.Done()
		kickassPath := filepath.Join(configDir, "kickass.json")

		// Load kickass directives from single file
		var err error
		builtinFunctions, builtinConstants, err = LoadBuiltins(kickassPath)
		if err != nil {
			log.Error("Failed to load kickass builtins from %s: %v", kickassPath, err)
		} else {
			log.Info("Successfully loaded %d built-in functions and %d built-in constants from %s", len(builtinFunctions), len(builtinConstants), kickassPath)
		}

		// Set kickass.json path for lexer
		SetKickassJSONPath(kickassPath)
	}()

	// Wait for all JSON files to load
	wg.Wait()
	log.Info("All JSON Source of Truth files loaded successfully from %s", configDir)

	// Initialize lexer token definitions AFTER all JSON files are loaded
	InitTokenDefs()

	// Initialize ProcessorContext (used by completion and context-aware parser)
	mnemonicPath := filepath.Join(configDir, "mnemonic.json")
	kickassPath := filepath.Join(configDir, "kickass.json")
	c64MemoryPath := filepath.Join(configDir, "c64memory.json")

	err = InitializeProcessorContext(configDir)
	if err != nil {
		log.Error("Failed to initialize ProcessorContext: %v", err)
		log.Error("Context-aware features (completion, parsing) will fall back to legacy mode")
	} else {
		log.Info("Successfully initialized ProcessorContext from %s, %s, %s", mnemonicPath, kickassPath, c64MemoryPath)
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
							"triggerCharacters": []string{" ", ".", "$"},
						},
						"definitionProvider":     true,
						"referencesProvider":     true,
						"documentSymbolProvider": true,
						"semanticTokensProvider": map[string]interface{}{
							"legend": map[string]interface{}{
								"tokenTypes": []string{
									"keyword",       // 0
									"variable",      // 1
									"function",      // 2
									"macro",         // 3
									"pseudocommand", // 4
									"number",        // 5
									"comment",       // 6
									"string",        // 7
									"operator",      // 8
									"mnemonic",      // 9
									"directive",     // 10
									"preprocessor",  // 11
									"label",         // 12
								},
								"tokenModifiers": []string{
									"declaration", "readonly",
								},
							},
							"full": true,
						},
						"workspace": map[string]interface{}{
							"workspaceFolders": map[string]interface{}{
								"supported": true,
							},
						},
					},
					"serverInfo": map[string]interface{}{
						"name":    "kickass_ls",
						"version": "1.0.0", // Version updated
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
		case "workspace/didChangeConfiguration":
			log.Debug("Handling workspace/didChangeConfiguration notification.")
			if params, ok := message["params"].(map[string]interface{}); ok {
				if settings, ok := params["settings"].(map[string]interface{}); ok {
					// Look for our specific LSP settings
					if lspSettings, ok := settings["kickass_ls"].(map[string]interface{}); ok {
						log.Debug("Updating LSP configuration")
						UpdateLSPConfig(lspSettings)

						// Initialize ProcessorContext if context-aware lexer is now enabled and not yet loaded
						if IsContextAwareLexerEnabled() && GetProcessorContext() == nil {
							homeDir, err := os.UserHomeDir()
							if err == nil {
								configDir := filepath.Join(homeDir, ".config", "kickass_ls")
								err = InitializeProcessorContext(configDir)
								if err != nil {
									log.Error("Failed to initialize ProcessorContext: %v", err)
								} else {
									log.Info("Successfully initialized ProcessorContext after config update")
								}
							}
						}

						// Invalidate all parse caches to trigger re-analysis with new settings
						parseCache.Lock()
						for uri := range parseCache.cache {
							delete(parseCache.cache, uri)
							log.Debug("Invalidated parse cache for %s due to config change", uri)
						}
						parseCache.Unlock()

						// Re-analyze all open documents with new configuration
						documentStore.RLock()
						for uri, content := range documentStore.documents {
							submitAnalysisJob(uri, content, writer, false)
						}
						documentStore.RUnlock()

						log.Info("Configuration updated and documents re-analyzed")
					} else {
						log.Debug("No kickass_ls settings found in configuration update")
					}
				}
			}
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

							// Submit analysis job asynchronously
							submitAnalysisJob(uri, text, writer, true)
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

									// Submit analysis job asynchronously
									submitAnalysisJob(uri, newText, writer, false)
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

						// Clear parse cache for closed document
						ClearParseCache(uri)

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

											// Also try to extract memory address (priority over regular words)
											memoryAddr := getMemoryAddressAtPosition(lineContent, int(charNum))
											if memoryAddr != "" {
												log.Logger.Printf("Memory address found: %s\n", memoryAddr)
												word = memoryAddr // Use memory address instead of regular word
											}

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
													// Check for built-in functions
													builtinFuncDescription := getBuiltinFunctionDescription(word)
													if builtinFuncDescription != "" {
														responseResult = map[string]interface{}{
															"contents": map[string]interface{}{
																"kind":  "markdown",
																"value": builtinFuncDescription,
															},
														}
													} else {
														// Check for built-in constants
														builtinConstDescription := getBuiltinConstantDescription(word)
														if builtinConstDescription != "" {
															responseResult = map[string]interface{}{
																"contents": map[string]interface{}{
																	"kind":  "markdown",
																	"value": builtinConstDescription,
																},
															}
														} else {
															// Check for C64 memory address description
															memoryDescription := getMemoryAddressDescription(word)
															if memoryDescription != "" {
																responseResult = map[string]interface{}{
																	"contents": map[string]interface{}{
																		"kind":  "markdown",
																		"value": memoryDescription,
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
											contextType, wordToComplete := getCompletionContext(lineContent, int(charNum))
											log.Debug("Completion context: contextType=%v, wordToComplete='%s'", contextType, wordToComplete)
											completionItems = generateCompletions(symbolTree, int(lineNum), contextType, wordToComplete, lineContent, int(charNum), text)
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

// LoadMnemonics is the exported version of loadMnemonics for test mode
func LoadMnemonics(path string) error {
	return loadMnemonics(path)
}

// loadC64MemoryMap loads the C64 memory map from c64memory.json
func loadC64MemoryMap(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &c64MemoryMap)
}

// LoadC64MemoryMap is the exported version of loadC64MemoryMap for test mode
func LoadC64MemoryMap(path string) error {
	return loadC64MemoryMap(path)
}

// GetCompletionContext is the exported version of getCompletionContext for test mode
func GetCompletionContext(line string, char int) (contextType CompletionContextType, word string) {
	return getCompletionContext(line, char)
}

// GenerateCompletions is the exported version of generateCompletions for test mode
func GenerateCompletions(symbolTree *Scope, lineNum int, contextType CompletionContextType, wordToComplete string) []map[string]interface{} {
	return generateCompletions(symbolTree, lineNum, contextType, wordToComplete, "", 0, "")
}

// GenerateCompletionsWithContext is the extended version for test mode with line context
func GenerateCompletionsWithContext(symbolTree *Scope, lineNum int, contextType CompletionContextType, wordToComplete string, lineContent string, cursorPos int) []map[string]interface{} {
	return generateCompletions(symbolTree, lineNum, contextType, wordToComplete, lineContent, cursorPos, "")
}

func getOpcodeDescription(mnemonic string) string {
	for _, m := range mnemonics {
		if m.Mnemonic == mnemonic {
			var builder strings.Builder

			// Header with mnemonic name and description
			builder.WriteString(fmt.Sprintf("**%s**\n\n", m.Mnemonic))
			builder.WriteString(fmt.Sprintf("%s\n\n", m.Description))

			// Check if this is an illegal opcode and add warning
			if m.Type == "Illegal" {
				builder.WriteString("⚠️ **ILLEGAL OPCODE** - Undocumented instruction, behavior may vary between processors\n\n")
			}

			// Use code block format instead of markdown table for better Neovim compatibility
			builder.WriteString("**Addressing Modes:**\n")
			builder.WriteString("```\n")
			builder.WriteString("Opcode   Mode              Format          Bytes   Cycles\n")
			builder.WriteString("------   ----              ------          -----   ------\n")

			for _, am := range m.AddressingModes {
				// Format as fixed-width columns for better alignment
				opcode := fmt.Sprintf("$%s", am.Opcode)
				mode := am.AddressingMode
				format := am.AssemblerFormat
				bytes := fmt.Sprintf("%d", am.Length)
				cycles := am.Cycles

				// Pad columns for precise alignment - increased Format column width
				builder.WriteString(fmt.Sprintf("%-8s %-17s %-15s %-7s %s\n",
					opcode, mode, format, bytes, cycles))
			}
			builder.WriteString("```\n")

			// CPU Flags section with enhanced formatting
			builder.WriteString("\n**CPU Flags Affected:**\n")
			if len(m.CPUFlags) > 0 {
				for _, flag := range m.CPUFlags {
					// Remove leading/trailing whitespace and ensure proper formatting
					cleanFlag := strings.TrimSpace(flag)
					if cleanFlag != "" {
						builder.WriteString(fmt.Sprintf("\n• %s", cleanFlag))
					}
				}
				builder.WriteString("\n")
			} else {
				builder.WriteString("\n• None\n")
			}

			return builder.String()
		}
	}
	return ""
}

func getDirectiveDescription(directive string) string {
	// Try ProcessorContext first (Context-Aware Parser)
	if ctx := GetProcessorContext(); ctx != nil {
		if info := ctx.GetDirectiveInfo(directive); info != nil {
			var builder strings.Builder

			// Header with directive name and signature
			builder.WriteString(fmt.Sprintf("**%s**\n\n", strings.ToUpper(info.Name)))

			// Signature in code block
			if info.Signature != "" {
				builder.WriteString("```kickassembler\n")
				builder.WriteString(info.Signature)
				builder.WriteString("\n```\n\n")
			}

			// Description
			if info.Description != "" {
				builder.WriteString(info.Description)
				builder.WriteString("\n\n")
			}

			// Examples
			if len(info.Examples) > 0 {
				builder.WriteString("**Examples:**\n\n")
				builder.WriteString("```kickassembler\n")
				builder.WriteString(strings.Join(info.Examples, "\n"))
				builder.WriteString("\n```")
			}

			return builder.String()
		}
	}

	// Fallback to old kickassDirectives array (legacy parser)
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

// getBuiltinFunctionDescription returns markdown description for built-in functions
func getBuiltinFunctionDescription(function string) string {
	// Try ProcessorContext first (Context-Aware Parser)
	if ctx := GetProcessorContext(); ctx != nil {
		if info := ctx.GetFunctionInfo(function); info != nil {
			var builder strings.Builder

			// Header with function name and category
			category := info.Category
			if category == "" {
				category = "builtin"
			}
			builder.WriteString(fmt.Sprintf("**%s** (%s function)\n\n", info.Name, category))

			// Signature in code block
			if info.Signature != "" {
				builder.WriteString("```kickassembler\n")
				builder.WriteString(info.Signature)
				builder.WriteString("\n```\n\n")
			}

			// Description
			if info.Description != "" {
				builder.WriteString(info.Description)
				builder.WriteString("\n\n")
			}

			// Examples
			if len(info.Examples) > 0 {
				builder.WriteString("**Examples:**\n\n")
				builder.WriteString("```kickassembler\n")
				builder.WriteString(strings.Join(info.Examples, "\n"))
				builder.WriteString("\n```")
			}

			return builder.String()
		}
	}

	// Fallback to old builtinFunctions array (legacy parser)
	for _, f := range builtinFunctions {
		if strings.EqualFold(f.Name, function) {
			var builder strings.Builder

			// Header with function name and category
			builder.WriteString(fmt.Sprintf("**%s** (%s function)\n\n", f.Name, f.Category))

			// Signature in code block
			if f.Signature != "" {
				builder.WriteString("```kickassembler\n")
				builder.WriteString(f.Signature)
				builder.WriteString("\n```\n\n")
			}

			// Description
			if f.Description != "" {
				builder.WriteString(f.Description)
				builder.WriteString("\n\n")
			}

			// Examples
			if len(f.Examples) > 0 {
				builder.WriteString("**Examples:**\n\n")
				builder.WriteString("```kickassembler\n")
				builder.WriteString(strings.Join(f.Examples, "\n"))
				builder.WriteString("\n```")
			}

			return builder.String()
		}
	}
	return ""
}

// getBuiltinConstantDescription returns markdown description for built-in constants
func getBuiltinConstantDescription(constant string) string {
	// Try ProcessorContext first (Context-Aware Parser)
	if ctx := GetProcessorContext(); ctx != nil {
		if info := ctx.GetConstantInfo(constant); info != nil {
			var builder strings.Builder

			// Header with constant name and category
			category := info.Category
			if category == "" {
				category = "builtin"
			}
			builder.WriteString(fmt.Sprintf("**%s** (%s constant)\n\n", info.Name, category))

			// Value in code block
			if info.Value != "" {
				builder.WriteString(fmt.Sprintf("**Value:** `%s`\n\n", info.Value))
			}

			// Description
			if info.Description != "" {
				builder.WriteString(info.Description)
				builder.WriteString("\n")
			}

			return builder.String()
		}
	}

	// Fallback to old builtinConstants array (legacy parser)
	for _, c := range builtinConstants {
		if strings.EqualFold(c.Name, constant) {
			var builder strings.Builder

			// Header with constant name and category
			builder.WriteString(fmt.Sprintf("**%s** (%s constant)\n\n", c.Name, c.Category))

			// Value in code block
			if c.Value != "" {
				builder.WriteString(fmt.Sprintf("**Value:** `%s`\n\n", c.Value))
			}

			// Description
			if c.Description != "" {
				builder.WriteString(c.Description)
				builder.WriteString("\n\n")
			}

			// Examples
			if len(c.Examples) > 0 {
				builder.WriteString("**Examples:**\n\n")
				builder.WriteString("```kickassembler\n")
				builder.WriteString(strings.Join(c.Examples, "\n"))
				builder.WriteString("\n```")
			}

			return builder.String()
		}
	}
	return ""
}

// getMemoryAddressDescription provides hover information for C64 memory addresses
func getMemoryAddressDescription(word string) string {
	// Debug: Log what word we're checking
	log.Debug("Checking memory address for word: '%s'", word)

	// Check if the word looks like a hex address ($xxxx or 0xxxxx)
	var addressStr string
	if strings.HasPrefix(word, "$") {
		addressStr = "0x" + strings.ToUpper(word[1:]) // Convert $d020 to 0xD020
	} else if strings.HasPrefix(strings.ToLower(word), "0x") {
		addressStr = "0x" + strings.ToUpper(word[2:]) // Convert 0xd020 to 0xD020
	} else {
		return "" // Not a hex address
	}

	// Check if we have information for this address in our memory map
	if region, found := c64MemoryMap.MemoryMap.Regions[addressStr]; found {
		var builder strings.Builder

		// Header with register name and category
		builder.WriteString(fmt.Sprintf("**%s** - %s\n\n", addressStr, region.Name))
		builder.WriteString(fmt.Sprintf("*%s %s*\n\n", region.Category, region.Type))

		// Description
		if region.Description != "" {
			builder.WriteString(region.Description)
			builder.WriteString("\n\n")
		}

		// Access mode
		if region.Access != "" {
			builder.WriteString(fmt.Sprintf("**Access:** %s\n\n", region.Access))
		}

		// Bit fields
		if len(region.BitFields) > 0 {
			builder.WriteString("**Bit Fields:**\n")
			for bits, desc := range region.BitFields {
				builder.WriteString(fmt.Sprintf("- **Bits %s:** %s\n", bits, desc))
			}
			builder.WriteString("\n")
		}

		// Values/Colors
		if len(region.Values) > 0 {
			builder.WriteString("**Values:**\n")
			count := 0
			for value, desc := range region.Values {
				if count >= 8 { // Limit to first 8 values for brevity
					builder.WriteString("- ...\n")
					break
				}
				builder.WriteString(fmt.Sprintf("- **%s:** %s\n", value, desc))
				count++
			}
			builder.WriteString("\n")
		}

		// Examples
		if len(region.Examples) > 0 {
			builder.WriteString("**Examples:**\n\n")
			builder.WriteString("```assembly\n")
			builder.WriteString(strings.Join(region.Examples, "\n"))
			builder.WriteString("\n```\n\n")
		}

		// Related addresses
		if len(region.Related) > 0 {
			builder.WriteString("**Related:** ")
			builder.WriteString(strings.Join(region.Related, ", "))
			builder.WriteString("\n\n")
		}

		// Programming tips
		if len(region.Tips) > 0 {
			builder.WriteString("**Tips:**\n")
			for _, tip := range region.Tips {
				builder.WriteString(fmt.Sprintf("- %s\n", tip))
			}
		}

		return builder.String()
	}

	return ""
}

func getWordAtPosition(line string, char int) string {
	if char < 0 || char >= len(line) {
		return ""
	}

	// Define word character set for assembly language identifiers
	// Include '.' for directive names like .byte, .const, .macro
	// Include '#' for preprocessor directives like #import, #define
	isWordChar := func(c byte) bool {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' || c == '#'
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

// getMemoryAddressAtPosition extracts memory addresses like $D020, $0800, etc.
func getMemoryAddressAtPosition(line string, char int) string {
	if char < 0 || char >= len(line) {
		return ""
	}

	log.Debug("getMemoryAddressAtPosition: line='%s', char=%d, charValue='%c'", line, char, line[char])

	// Check if we're on a hex digit or $ sign
	isHexChar := func(c byte) bool {
		return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
	}

	// First, check if we're anywhere within a potential memory address
	// Look backwards and forwards to find a complete $xxxx pattern

	// Find the start by going backwards until we hit a non-hex/non-$ character
	start := char
	for start > 0 {
		prevChar := line[start-1]
		if prevChar == '$' || isHexChar(prevChar) {
			start--
		} else {
			break
		}
	}

	// Find the end by going forwards from current position
	end := char
	for end < len(line) {
		currentChar := line[end]
		if currentChar == '$' || isHexChar(currentChar) {
			end++
		} else {
			break
		}
	}

	// Extract the potential memory address
	if start < end {
		candidate := line[start:end]
		log.Debug("getMemoryAddressAtPosition: candidate='%s'", candidate)

		// Check if it's a valid memory address (starts with $ and has hex digits)
		if strings.HasPrefix(candidate, "$") && len(candidate) >= 2 {
			// Verify the part after $ contains only hex digits
			hexPart := candidate[1:]
			if len(hexPart) > 0 && len(hexPart) <= 4 {
				for _, c := range hexPart {
					if !isHexChar(byte(c)) {
						log.Debug("getMemoryAddressAtPosition: invalid hex char '%c'", c)
						return ""
					}
				}
				log.Debug("getMemoryAddressAtPosition: returning '%s'", candidate)
				return candidate
			}
		}
	}

	log.Debug("getMemoryAddressAtPosition: no valid address found")
	return ""
}

// getUsedOperandsForMnemonic scans the document and collects operands that were already used with a specific mnemonic/directive
func getUsedOperandsForMnemonic(text string, mnemonicOrDirective string) []string {
	lines := strings.Split(text, "\n")
	usedOperands := make(map[string]bool) // Use map to avoid duplicates
	result := []string{}

	targetUpper := strings.ToUpper(mnemonicOrDirective)

	for _, line := range lines {
		// Remove comments
		if idx := strings.Index(line, ";"); idx != -1 {
			line = line[:idx]
		}
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Split into fields
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		// Skip label if present
		firstWord := fields[0]
		if strings.HasSuffix(firstWord, ":") {
			if len(fields) < 3 {
				continue
			}
			// Check if second word matches our mnemonic/directive
			if strings.ToUpper(fields[1]) == targetUpper || fields[1] == mnemonicOrDirective {
				// Operand is everything after the mnemonic/directive
				operand := strings.Join(fields[2:], " ")
				if operand != "" && !usedOperands[operand] {
					usedOperands[operand] = true
					result = append(result, operand)
				}
			}
		} else {
			// No label - check if first word matches
			if strings.ToUpper(firstWord) == targetUpper || firstWord == mnemonicOrDirective {
				// Operand is everything after the mnemonic/directive
				operand := strings.Join(fields[1:], " ")

				// Special handling for .var and .const: extract only the value after "="
				if mnemonicOrDirective == ".var" || mnemonicOrDirective == ".const" {
					// Format: .var name = value
					// or: .const name = value
					// We want only "value"
					if eqIdx := strings.Index(operand, "="); eqIdx != -1 {
						// Extract everything after "="
						operand = strings.TrimSpace(operand[eqIdx+1:])
					} else {
						// No "=" found, skip this line
						continue
					}
				}

				if operand != "" && !usedOperands[operand] {
					usedOperands[operand] = true
					result = append(result, operand)
				}
			}
		}
	}

	return result
}

// getPrecedingMnemonicOrDirective extracts the mnemonic or directive that precedes the cursor
// Returns (name, isMnemonic, isDirective)
func getPrecedingMnemonicOrDirective(lineContent string, cursorPos int) (string, bool, bool) {
	if cursorPos > len(lineContent) {
		cursorPos = len(lineContent)
	}

	// Get text before cursor
	before := lineContent[:cursorPos]

	// Remove comments
	if idx := strings.Index(before, ";"); idx != -1 {
		before = before[:idx]
	}
	if idx := strings.Index(before, "//"); idx != -1 {
		before = before[:idx]
	}

	// Split into words
	fields := strings.Fields(before)
	if len(fields) == 0 {
		return "", false, false
	}

	// Find the first word (after optional label)
	firstWord := fields[0]
	if strings.HasSuffix(firstWord, ":") {
		// Has label, take second word
		if len(fields) < 2 {
			return "", false, false
		}
		firstWord = fields[1]
	}

	// Check if it's a directive
	if strings.HasPrefix(firstWord, ".") {
		return firstWord, false, true
	}

	// Check if it's a mnemonic
	if isMnemonic(firstWord) {
		return strings.ToUpper(firstWord), true, false
	}

	return "", false, false
}

func generateCompletions(symbolTree *Scope, lineNum int, contextType CompletionContextType, wordToComplete string, lineContent string, cursorPos int, documentText string) []map[string]interface{} {
	items := []map[string]interface{}{}

	log.Debug("generateCompletions called: lineNum=%d, contextType=%v, wordToComplete='%s', lineContent='%s', cursorPos=%d",
		lineNum, contextType, wordToComplete, lineContent, cursorPos)

	// Determine context: are we after a mnemonic or directive?
	precedingName, isMnem, isDir := getPrecedingMnemonicOrDirective(lineContent, cursorPos)
	log.Debug("Preceding context: name='%s', isMnemonic=%v, isDirective=%v", precedingName, isMnem, isDir)

	// Check if we're after a directive that expects a NAME (declaration), not a value
	// For these directives, we should NOT offer completions until after the "="
	declarationDirectives := map[string]bool{
		".var":           true,
		".const":         true,
		".label":         true,
		".macro":         true,
		".function":      true,
		".namespace":     true,
		".pseudocommand": true,
		".enum":          true,
	}

	// If after a declaration directive and no "=" yet, don't offer completions
	if contextType == ContextDirectiveOperand && isDir && declarationDirectives[precedingName] {
		// Check if there's already a "=" in the line - if so, we're past the name part
		if !strings.Contains(lineContent[:cursorPos], "=") {
			// We're at the name part - don't offer completions (user should type new name)
			log.Debug("After declaration directive '%s' before '=' - no completions offered", precedingName)
			return items // Return empty list
		}
	}

	// CRITICAL FIX: Handle space trigger character case
	// When user types "lda<space>", LSP sends completion request BEFORE space is in buffer
	// So we get: contextType=ContextMnemonic, wordToComplete='lda', cursor at end of 'lda'
	// We need to detect this and offer addressing mode hints anyway
	offerAddressingHints := false
	targetMnemonic := ""

	// Case 1: Normal - "lda " with space already in buffer
	if contextType == ContextOperandGeneral && precedingName != "" && isMnem {
		offerAddressingHints = true
		targetMnemonic = precedingName
		log.Debug("Case 1: offering hints for mnemonic %s (ContextOperandGeneral, after space)", targetMnemonic)
	}
	// Case 2: Space trigger - user typed "lda<space>", space not yet in buffer
	if contextType == ContextMnemonic && isMnemonic(wordToComplete) && cursorPos == len(strings.TrimRight(lineContent, " \t")) {
		offerAddressingHints = true
		targetMnemonic = strings.ToUpper(wordToComplete)
		log.Debug("Case 2: offering hints for mnemonic %s (ContextMnemonic, space trigger)", targetMnemonic)
	}

	// Offer addressing mode hints if applicable
	if offerAddressingHints && targetMnemonic != "" {
		if ctx := GetProcessorContext(); ctx != nil {
			log.Debug("ProcessorContext is available")
			mnemonicInfo := ctx.GetMnemonicInfo(targetMnemonic)
			log.Debug("MnemonicInfo for %s: %+v", targetMnemonic, mnemonicInfo)
			if mnemonicInfo != nil && len(mnemonicInfo.AddressingModes) > 0 {
				log.Debug("Found %d addressing modes for %s", len(mnemonicInfo.AddressingModes), targetMnemonic)
				// Build unique set of addressing mode prefixes
				addedModes := make(map[string]bool)

				for _, mode := range mnemonicInfo.AddressingModes {
					var prefix, doc string
					switch mode.Mode {
					case "Immediate":
						prefix = "#"
						doc = fmt.Sprintf("Immediate addressing - %s", mode.AssemblerFormat)
					case "Absolute", "Absolute,X", "Absolute,Y", "Zeropage", "Zeropage,X", "Zeropage,Y":
						if !addedModes["$"] {
							prefix = "$"
							doc = "Memory address (absolute or zero page)"
						}
					case "Indexed-indirect", "Indirect-indexed", "Indirect":
						if !addedModes["("] {
							prefix = "("
							doc = "Indirect addressing"
						}
					}

					if prefix != "" && !addedModes[prefix] {
						items = append(items, map[string]interface{}{
							"label":         prefix,
							"kind":          float64(14), // Keyword
							"detail":        "Addressing Mode",
							"documentation": doc,
							"sortText":      "0_" + prefix, // Sort first
						})
						addedModes[prefix] = true
					}
				}
				log.Debug("Added %d addressing mode completion items", len(addedModes))
			} else {
				if mnemonicInfo == nil {
					log.Debug("MnemonicInfo is nil for %s", targetMnemonic)
				} else {
					log.Debug("No addressing modes found for %s", targetMnemonic)
				}
			}
		} else {
			log.Warn("ProcessorContext is nil - cannot offer addressing mode hints")
		}
	}

	// If we're completing an operand, offer context-specific hints
	if (contextType == ContextOperandGeneral || contextType == ContextJumpTarget || contextType == ContextDirectiveOperand) && precedingName != "" {
		// Offer previously used operands for this mnemonic/directive
		if documentText != "" {
			usedOperands := getUsedOperandsForMnemonic(documentText, precedingName)
			for _, operand := range usedOperands {
				// Only add if it matches what user is typing
				if wordToComplete == "" || strings.HasPrefix(strings.ToLower(operand), strings.ToLower(wordToComplete)) {
					items = append(items, map[string]interface{}{
						"label":         operand,
						"kind":          float64(12), // Value
						"detail":        "Recently used",
						"documentation": fmt.Sprintf("Previously used with %s", precedingName),
						"sortText":      "1_" + operand, // Sort after addressing mode hints
					})
				}
			}
		}
	}

	// Special case: check for memory address completion even in mnemonic context
	// This handles cases where cursor is on or near $ symbol
	memoryPrefix := ""
	log.Debug("Checking for memory completion: wordToComplete='%s', lineContent='%s', cursorPos=%d", wordToComplete, lineContent, cursorPos)

	if strings.HasPrefix(wordToComplete, "$") {
		memoryPrefix = strings.ToUpper(wordToComplete)
		log.Debug("Found $ prefix in wordToComplete: '%s'", memoryPrefix)
	} else if lineContent != "" && cursorPos <= len(lineContent) {
		// Look backwards in the line to see if we're completing a memory address
		startPos := cursorPos
		if startPos > len(lineContent) {
			startPos = len(lineContent)
		}

		for i := startPos - 1; i >= 0; i-- {
			log.Debug("Checking character at position %d: '%c'", i, lineContent[i])
			if i < len(lineContent) && lineContent[i] == '$' {
				// Found $, extract from $ to cursor position
				memoryPrefix = strings.ToUpper(lineContent[i:startPos])
				log.Debug("Found $ in line at position %d, extracted: '%s'", i, memoryPrefix)
				break
			} else if i < len(lineContent) && lineContent[i] == '#' {
				// Found # - check if we should offer memory addresses
				// Only if cursor is directly after # with no other characters
				if i+1 == startPos {
					memoryPrefix = "$"
					log.Debug("Found # at position %d with cursor directly after, offering memory addresses", i)
				} else {
					log.Debug("Found # at position %d but cursor not directly after, skipping memory completion", i)
				}
				break
			} else if i < len(lineContent) && (lineContent[i] == ' ' || lineContent[i] == '\t') {
				// Hit whitespace, stop looking
				log.Debug("Hit whitespace at position %d, stopping search", i)
				break
			}
		}
	}

	// If we found memory prefix, add memory completions regardless of isOperand
	if memoryPrefix != "" {
		log.Debug("Memory address completion requested with prefix: '%s'", memoryPrefix)

		// Add all memory registers that match the prefix
		for address, region := range c64MemoryMap.MemoryMap.Regions {
			// Convert 0xD000 format to $D000 format for matching
			addressHex := strings.TrimPrefix(address, "0x")
			addressWithDollar := "$" + addressHex
			if strings.HasPrefix(strings.ToUpper(addressWithDollar), strings.ToUpper(memoryPrefix)) {
				// Create documentation string
				documentation := fmt.Sprintf("**%s** - %s\n\n*%s %s*\n\n%s",
					"0x"+address, region.Name, region.Category, region.Type, region.Description)

				// Add access information
				if region.Access != "" {
					documentation += fmt.Sprintf("\n\n**Access:** %s", region.Access)
				}

				// Add bit fields if available
				if len(region.BitFields) > 0 {
					documentation += "\n\n**Bit Fields:**"
					for bits, desc := range region.BitFields {
						documentation += fmt.Sprintf("\n- **Bits %s:** %s", bits, desc)
					}
				}

				item := map[string]interface{}{
					"label":         addressWithDollar,
					"kind":          float64(12), // Value/Constant
					"detail":        fmt.Sprintf("%s - %s", region.Category, region.Name),
					"documentation": documentation,
					"insertText":    addressWithDollar,
					"sortText":      fmt.Sprintf("0_%s", address), // Sort memory addresses first
				}
				items = append(items, item)
			}
		}
	}

	// Handle completions based on context type
	switch contextType {
	case ContextJumpTarget:
		// After jmp/jsr/branch - suggest ONLY labels
		log.Debug("ContextJumpTarget: suggesting labels only")
		if memoryPrefix != "" {
			log.Debug("Already found memory completions, skipping other completions")
			return items
		}
		symbols := symbolTree.FindAllVisibleSymbols(lineNum)
		for _, symbol := range symbols {
			if symbol.Kind == Label && strings.HasPrefix(symbol.Name, wordToComplete) {
				item := map[string]interface{}{
					"label":  symbol.Name,
					"kind":   toCompletionItemKind(symbol.Kind),
					"detail": symbol.Value,
				}
				items = append(items, item)
			}
		}

	case ContextImmediate:
		// After # - suggest constants and numbers only
		log.Debug("ContextImmediate: suggesting constants only")
		wordToComplete = strings.TrimPrefix(wordToComplete, "#")

		// Add built-in constants
		for _, const_ := range builtinConstants {
			if strings.HasPrefix(strings.ToLower(const_.Name), strings.ToLower(wordToComplete)) {
				item := map[string]interface{}{
					"label":         const_.Name,
					"kind":          float64(21), // Constant
					"detail":        fmt.Sprintf("%s constant", const_.Category),
					"documentation": const_.Description,
				}
				if const_.Value != "" {
					item["detail"] = fmt.Sprintf("%s = %s", const_.Name, const_.Value)
				}
				items = append(items, item)
			}
		}

		// Add user-defined constants
		symbols := symbolTree.FindAllVisibleSymbols(lineNum)
		for _, symbol := range symbols {
			if symbol.Kind == Constant && strings.HasPrefix(symbol.Name, wordToComplete) {
				item := map[string]interface{}{
					"label":  symbol.Name,
					"kind":   toCompletionItemKind(symbol.Kind),
					"detail": symbol.Value,
				}
				items = append(items, item)
			}
		}

	case ContextOperandGeneral, ContextDirectiveOperand:
		// General operand context - suggest labels, constants, variables
		log.Debug("ContextOperandGeneral/DirectiveOperand: suggesting all operands")
		if memoryPrefix != "" {
			log.Debug("Already found memory completions, skipping other operand completions")
			return items
		}

		wordToComplete = strings.TrimPrefix(wordToComplete, "#")
		// Check for namespace access (e.g. "foo.bar"), but NOT directives starting with "."
		if strings.Contains(wordToComplete, ".") && !strings.HasPrefix(wordToComplete, ".") {
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
			// Only offer built-in functions and constants if we're NOT after a mnemonic
			// (Mnemonics expect addresses, symbols, or immediate values - not function calls)
			offerBuiltins := !isMnem

			// Check if we should use relaxed matching for built-ins
			useRelaxedMatching := len(wordToComplete) <= 2 ||
				(len(wordToComplete) <= 3 && strings.TrimSpace(wordToComplete) != "" &&
					strings.ContainsAny(wordToComplete, "0123456789"))

			// Add built-in functions (only if appropriate for context)
			if offerBuiltins {
				for _, fn := range builtinFunctions {
					shouldInclude := useRelaxedMatching ||
						strings.HasPrefix(strings.ToLower(fn.Name), strings.ToLower(wordToComplete))
					if shouldInclude {
						item := map[string]interface{}{
							"label":            fn.Name,
							"kind":             float64(3), // Function
							"detail":           fn.Signature,
							"documentation":    fmt.Sprintf("**%s**\n\n%s", fn.Category, fn.Description),
							"insertText":       fn.Name + "(${1})",
							"insertTextFormat": 2, // Snippet
						}
						if len(fn.Examples) > 0 {
							item["documentation"] = fmt.Sprintf("**%s**\n\n%s\n\n**Example:** `%s`",
								fn.Category, fn.Description, fn.Examples[0])
						}
						items = append(items, item)
					}
				}

				// Add built-in constants
				for _, const_ := range builtinConstants {
					shouldInclude := useRelaxedMatching ||
						strings.HasPrefix(strings.ToLower(const_.Name), strings.ToLower(wordToComplete))
					if shouldInclude {
						item := map[string]interface{}{
							"label":         const_.Name,
							"kind":          float64(21), // Constant
							"detail":        fmt.Sprintf("%s constant", const_.Category),
							"documentation": const_.Description,
						}
						if const_.Value != "" {
							item["detail"] = fmt.Sprintf("%s = %s", const_.Name, const_.Value)
						}
						if len(const_.Examples) > 0 {
							item["documentation"] = fmt.Sprintf("%s\n\n**Example:** `%s`",
								const_.Description, const_.Examples[0])
						}
						items = append(items, item)
					}
				}
			}

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

	case ContextDirective:
		// After . - suggest directives only
		log.Debug("ContextDirective: suggesting directives only")
		wordWithoutDot := strings.TrimPrefix(wordToComplete, ".")

		// Try ProcessorContext first (Context-Aware Parser)
		if ctx := GetProcessorContext(); ctx != nil {
			for _, directiveName := range ctx.DirectiveNames {
				displayName := strings.TrimPrefix(directiveName, ".")
				if strings.HasPrefix(strings.ToLower(displayName), strings.ToLower(wordWithoutDot)) {
					directiveInfo := ctx.GetDirectiveInfo(directiveName)
					documentation := ""
					if directiveInfo != nil {
						documentation = directiveInfo.Description
					}
					items = append(items, map[string]interface{}{
						"label":         "." + displayName,
						"kind":          float64(14), // Keyword
						"detail":        "Kick Assembler Directive",
						"documentation": documentation,
					})
				}
			}
		} else {
			// Fallback to old kickassDirectives array (legacy parser)
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
		}

	case ContextMnemonic:
		// Offer directives
		// Strip leading dot from wordToComplete for matching (user types "." or ".by" etc.)
		wordWithoutDot := strings.TrimPrefix(wordToComplete, ".")

		// Try ProcessorContext first (Context-Aware Parser)
		if ctx := GetProcessorContext(); ctx != nil {
			for _, directiveName := range ctx.DirectiveNames {
				// Remove the dot prefix for matching (directives stored as ".byte" etc.)
				displayName := strings.TrimPrefix(directiveName, ".")
				if strings.HasPrefix(strings.ToLower(displayName), strings.ToLower(wordWithoutDot)) {
					directiveInfo := ctx.GetDirectiveInfo(directiveName)
					documentation := ""
					if directiveInfo != nil {
						documentation = directiveInfo.Description
					}
					// Always include the dot in the label
					items = append(items, map[string]interface{}{
						"label":         "." + displayName,
						"kind":          float64(14), // Keyword
						"detail":        "Kick Assembler Directive",
						"documentation": documentation,
					})
				}
			}
		} else {
			// Fallback to old kickassDirectives array (legacy parser)
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

		// Only offer mnemonics if we're truly at the beginning of a line or after a label
		// Check if we're completing after a mnemonic + space (which would be wrong)
		shouldOfferMnemonics := true
		if len(lineContent) > 0 && cursorPos > 0 {
			// Look at the line content before cursor to see if there's already a mnemonic
			// Ensure cursor position doesn't exceed line length
			actualCursorPos := cursorPos
			if actualCursorPos > len(lineContent) {
				actualCursorPos = len(lineContent)
			}
			beforeCursor := lineContent[:actualCursorPos]
			trimmed := strings.TrimSpace(beforeCursor)
			if trimmed != "" {
				parts := strings.Fields(trimmed)
				if len(parts) > 0 {
					lastWord := parts[len(parts)-1]
					// If the last word is a mnemonic, don't offer more mnemonics
					if isMnemonic(lastWord) {
						shouldOfferMnemonics = false
						log.Debug("Not offering mnemonics after existing mnemonic: '%s'", lastWord)
					}
				}
			}
		}

		// Offer mnemonics only if appropriate
		if shouldOfferMnemonics {
			// Try ProcessorContext first (Context-Aware Parser)
			if ctx := GetProcessorContext(); ctx != nil {
				for mnemonicName := range ctx.AllMnemonics {
					if strings.HasPrefix(strings.ToUpper(mnemonicName), strings.ToUpper(wordToComplete)) {
						mnemonicInfo := ctx.GetMnemonicInfo(mnemonicName)
						detail := "6502/6510 Opcode"
						documentation := ""
						if mnemonicInfo != nil {
							documentation = mnemonicInfo.Description
							if mnemonicInfo.Type == "Illegal" {
								detail = "6502/6510 Illegal Opcode"
							}
						}
						items = append(items, map[string]interface{}{
							"label":         applyCase(wordToComplete, mnemonicName),
							"kind":          float64(14), // Keyword
							"detail":        detail,
							"documentation": documentation,
						})
					}
				}
			} else {
				// Fallback to old mnemonics array (legacy parser)
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

// isJumpInstruction checks if the word is a jump/branch instruction
// Uses ProcessorContext to check mnemonic type from mnemonic.json
// Returns true for both Jump and Branch types (for completion context)
func isJumpInstruction(word string) bool {
	if ctx := GetProcessorContext(); ctx != nil {
		mnemonicInfo := ctx.GetMnemonicInfo(strings.ToUpper(word))
		if mnemonicInfo != nil {
			// For completion context, we want labels after BOTH jumps and branches
			return mnemonicInfo.Type == "Jump" || mnemonicInfo.Type == "Branch"
		}
	}
	return false
}

func isDirective(word string) bool {
	// Try ProcessorContext first (Context-Aware Parser)
	if ctx := GetProcessorContext(); ctx != nil {
		// ProcessorContext stores directives with dot prefix, so add it if missing
		directive := word
		if !strings.HasPrefix(directive, ".") {
			directive = "." + directive
		}
		if ctx.IsValidDirective(directive) {
			return true
		}
	}

	// Fallback to old kickassDirectives array (legacy parser)
	for _, d := range kickassDirectives {
		if strings.EqualFold(d.Directive, word) {
			return true
		}
	}
	return false
}

// extractWordAtPosition extracts the word being typed at the given position in the text.
// It looks backward and forward from the position to find word boundaries.
func extractWordAtPosition(text string, pos int) string {
	if pos < 0 || pos > len(text) {
		return ""
	}

	// Define word character set for assembly language identifiers
	// Include '.' for directive names like .byte, .const, .macro
	isWordChar := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.'
	}

	// Find start of word by going backwards
	start := pos
	for start > 0 && isWordChar(rune(text[start-1])) {
		start--
	}

	// Find end of word by going forwards
	end := pos
	for end < len(text) && isWordChar(rune(text[end])) {
		end++
	}

	return text[start:end]
}

// CompletionContextType specifies what kind of completion context we're in
type CompletionContextType int

const (
	ContextMnemonic         CompletionContextType = iota // At line start or after label - suggest mnemonics/directives/macros
	ContextDirective                                     // After . - suggest directives only
	ContextJumpTarget                                    // After jmp/jsr/bra/etc - suggest labels only
	ContextImmediate                                     // After # - suggest constants/numbers only
	ContextOperandGeneral                                // After other mnemonics - suggest all operands
	ContextDirectiveOperand                              // After directive - context-specific suggestions
)

// getCompletionContext determines what kind of completion context we're in
// and returns the context type and word being completed.
func getCompletionContext(line string, char int) (CompletionContextType, string) {
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
		return ContextMnemonic, ""
	}

	parts := strings.Fields(trimmedContext)
	log.Debug("Parts: %v", parts)

	if len(parts) == 0 {
		log.Debug("No parts found, assuming mnemonic context.")
		return ContextMnemonic, ""
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
				return ContextMnemonic, ""
			}
		}
		if strings.HasPrefix(verb, ":") { // It's a macro/pseudocommand call with ':' prefix
			log.Debug("Cursor after a ':' prefixed macro/pseudocommand, assuming operand context.")
			return ContextOperandGeneral, ""
		}

		// Check if verb is a jump instruction
		if isJumpInstruction(verb) {
			log.Debug("Cursor after jump instruction '%s', suggesting labels only.", verb)
			return ContextJumpTarget, ""
		}

		if isMnemonic(verb) {
			log.Debug("Cursor after a mnemonic '%s', assuming operand context.", verb)
			return ContextOperandGeneral, ""
		}

		if isDirective(verb) {
			log.Debug("Cursor after a directive '%s', assuming directive operand context.", verb)
			return ContextDirectiveOperand, ""
		}

		// e.g. after a constant definition "MAX_SPRITES = 8 |"
		log.Debug("Cursor in whitespace, but not after a known mnemonic/directive. Assuming mnemonic context for a new line.")
		return ContextMnemonic, ""
	}

	// Cursor is in the middle of a word.
	// Instead of using the last field from strings.Fields, we need to extract
	// the actual word being typed at the cursor position.
	wordToComplete := extractWordAtPosition(context, len(context))
	log.Debug("Word to complete: '%s'", wordToComplete)

	// Special case: if word starts with '.', it's definitely a directive (not an operand)
	if strings.HasPrefix(wordToComplete, ".") {
		log.Debug("Word starts with '.', definitely a directive context.")
		return ContextDirective, wordToComplete
	}

	// Special case: if word starts with '#', it's immediate addressing (constants only)
	if strings.HasPrefix(wordToComplete, "#") {
		log.Debug("Word starts with '#', immediate addressing - constants only.")
		return ContextImmediate, wordToComplete
	}

	// Is this word the "verb" (mnemonic/directive) or an operand?
	verbIndex := 0
	if len(parts) > 0 && strings.HasSuffix(parts[0], ":") {
		verbIndex = 1
	}

	// If we are completing a word at or before the verb index, it's a mnemonic/directive context.
	if len(parts)-1 <= verbIndex {
		log.Debug("Completing the verb part of the line.")
		return ContextMnemonic, wordToComplete
	}

	// We are completing a word after the verb. This is an operand.
	verb := parts[verbIndex]

	// Check if verb is a jump instruction
	if isJumpInstruction(verb) {
		log.Debug("Completing after jump instruction '%s', suggesting labels only.", verb)
		return ContextJumpTarget, wordToComplete
	}

	if isMnemonic(verb) {
		log.Debug("Completing after a known mnemonic ('%s'), this is an operand.", verb)
		return ContextOperandGeneral, wordToComplete
	}

	if isDirective(verb) {
		log.Debug("Completing after a directive ('%s'), this is a directive operand.", verb)
		return ContextDirectiveOperand, wordToComplete
	}

	// Fallback: if the "verb" is not a known mnemonic/directive (e.g. a macro call),
	// we can assume what follows is an operand.
	if verbIndex < len(parts)-1 {
		log.Debug("Completing after an unknown verb ('%s'), assuming operand.", verb)
		return ContextOperandGeneral, wordToComplete
	}

	// Default fallback
	log.Debug("Defaulting to mnemonic context.")
	return ContextMnemonic, wordToComplete
}
func toCompletionItemKind(kind SymbolKind) CompletionItemKind {
	switch kind {
	case Constant:
		return ConstantCompletion
	case Variable:
		return VariableCompletion
	case Label:
		return VariableCompletion
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

// LoadBuiltins loads built-in functions and constants from kickass.json
func LoadBuiltins(filePath string) ([]BuiltinFunction, []BuiltinConstant, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var config struct {
		BuiltinFunctions []BuiltinFunction `json:"builtinFunctions"`
		BuiltinConstants []BuiltinConstant `json:"builtinConstants"`
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, nil, err
	}

	return config.BuiltinFunctions, config.BuiltinConstants, nil
}

// SetBuiltins sets the global built-in functions and constants
func SetBuiltins(functions []BuiltinFunction, constants []BuiltinConstant) {
	builtinFunctions = functions
	builtinConstants = constants
}

// GenerateHover generates hover information for a word at a specific position
func GenerateHover(symbolTree *Scope, line string, char int) (string, bool) {
	word := getWordAtPosition(line, char)
	log.Debug("GenerateHover called: line='%s', char=%d, word='%s'", line, char, word)

	// Also try to extract memory address (priority over regular words)
	memoryAddr := getMemoryAddressAtPosition(line, char)
	if memoryAddr != "" {
		log.Debug("Memory address found: %s", memoryAddr)
		word = memoryAddr // Use memory address instead of regular word
	}

	if word == "" {
		return "", false
	}

	// Check for opcode description
	description := getOpcodeDescription(strings.ToUpper(word))
	if description != "" {
		return description, true
	}

	// Check for directive description
	directiveDescription := getDirectiveDescription(strings.ToLower(word))
	if directiveDescription != "" {
		return directiveDescription, true
	}

	// Check for built-in functions
	builtinFuncDescription := getBuiltinFunctionDescription(word)
	if builtinFuncDescription != "" {
		return builtinFuncDescription, true
	}

	// Check for built-in constants
	builtinConstDescription := getBuiltinConstantDescription(word)
	if builtinConstDescription != "" {
		return builtinConstDescription, true
	}

	// Check for memory address documentation
	log.Debug("About to call getMemoryAddressDescription for word: '%s'", word)
	memoryDescription := getMemoryAddressDescription(word)
	if memoryDescription != "" {
		log.Debug("Found memory description for: '%s'", word)
		return memoryDescription, true
	} else {
		log.Debug("No memory description found for: '%s'", word)
	}

	// Check for symbol in symbol tree
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
		return markdown, true
	}

	return "", false
}

// GenerateSignatureHelp generates signature help for function calls at a specific position
func GenerateSignatureHelp(symbolTree *Scope, line string, char int) (string, bool) {
	// Debug: show what we're analyzing
	if char > len(line) {
		char = len(line)
	}

	// Find function call context - look backwards for opening parenthesis
	openParen := -1
	for i := char; i >= 0; i-- {
		if line[i] == '(' {
			openParen = i
			break
		} else if line[i] == ')' || line[i] == ';' {
			// Found closing paren or comment before opening - not in function call
			return "", false
		}
	}

	if openParen == -1 {
		return "", false
	}

	// Extract function name before the opening parenthesis
	funcName := ""
	for i := openParen - 1; i >= 0; i-- {
		c := line[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			funcName = string(c) + funcName
		} else {
			break
		}
	}

	if funcName == "" {
		return "", false
	}

	// Check built-in functions first

	for _, fn := range builtinFunctions {
		if strings.EqualFold(fn.Name, funcName) {
			// Count current parameter position by counting commas
			paramIndex := 0
			for i := openParen + 1; i < char; i++ {
				if line[i] == ',' {
					paramIndex++
				}
			}

			// Format signature with current parameter highlighted
			signature := fmt.Sprintf("**%s**\n\n```kickassembler\n%s\n```\n\n%s", fn.Name, fn.Signature, fn.Description)

			if len(fn.Examples) > 0 {
				signature += fmt.Sprintf("\n\n**Example:**\n```kickassembler\n%s\n```", fn.Examples[0])
			}

			// Add parameter position info
			signature += fmt.Sprintf("\n\n*Parameter %d*", paramIndex+1)

			return signature, true
		}
	}

	// Check user-defined functions/macros in symbol tree
	if symbol, found := symbolTree.FindSymbol(normalizeLabel(funcName)); found {
		if symbol.Kind == Function || symbol.Kind == Macro {
			signature := fmt.Sprintf("**%s** (%s)", symbol.Name, symbol.Kind.String())
			if symbol.Signature != "" {
				signature += fmt.Sprintf("\n\n```kickassembler\n%s\n```", symbol.Signature)
			}
			return signature, true
		}
	}

	return "", false
}

// ListSymbols returns all symbols in a scope with their metadata
func ListSymbols(scope *Scope) []map[string]interface{} {
	var symbols []map[string]interface{}

	// Add symbols from all scopes recursively
	addSymbolsFromScope(scope, &symbols)

	return symbols
}

func addSymbolsFromScope(scope *Scope, symbols *[]map[string]interface{}) {
	if scope == nil {
		return
	}

	// Add symbols from current scope
	for _, symbol := range scope.Symbols {
		symbolInfo := map[string]interface{}{
			"name":     symbol.Name,
			"type":     strings.ToLower(symbol.Kind.String()),
			"location": fmt.Sprintf("line %d", symbol.Position.Line+1),
		}

		// Add additional details based on symbol type
		if symbol.Value != "" {
			if symbol.Kind == Constant || symbol.Kind == Variable {
				symbolInfo["detail"] = fmt.Sprintf("= %s", symbol.Value)
			} else {
				symbolInfo["detail"] = symbol.Value
			}
		}

		if symbol.Signature != "" {
			symbolInfo["detail"] = symbol.Signature
		}

		*symbols = append(*symbols, symbolInfo)
	}

	// Recursively add symbols from child scopes
	for _, childScope := range scope.Children {
		addSymbolsFromScope(childScope, symbols)
	}
}

// GetBuiltins returns the current built-in functions and constants
func GetBuiltins() ([]BuiltinFunction, []BuiltinConstant) {
	return builtinFunctions, builtinConstants
}

// GetBuiltinFunctions returns the global builtin functions for validation
func GetBuiltinFunctions() []BuiltinFunction {
	return builtinFunctions
}

// FindReferences finds all references to a symbol at a specific position
func FindReferences(scope *Scope, line string, char int, lineNum int) ([]map[string]interface{}, string) {
	// Get word at position
	word := getWordAtPosition(line, char)
	if word == "" {
		return nil, ""
	}

	// Check if it's a built-in
	for _, fn := range builtinFunctions {
		if strings.EqualFold(fn.Name, word) {
			return []map[string]interface{}{
				{
					"location": "Built-in function",
					"type":     "built-in",
				},
			}, word
		}
	}

	for _, const_ := range builtinConstants {
		if strings.EqualFold(const_.Name, word) {
			return []map[string]interface{}{
				{
					"location": "Built-in constant",
					"type":     "built-in",
				},
			}, word
		}
	}

	// Find symbol in scope
	normalizedSymbol := normalizeLabel(word)
	symbol, found := scope.FindSymbol(normalizedSymbol)
	if !found {
		return nil, word
	}

	// Find all references to this symbol
	references := []map[string]interface{}{}

	// Add definition location
	references = append(references, map[string]interface{}{
		"location": fmt.Sprintf("line %d:%d", symbol.Position.Line+1, symbol.Position.Character+1),
		"type":     "definition",
	})

	// Find usages (this is simplified - in reality we'd need to parse and track all usages)
	// For now, we'll just show the definition
	references = append(references, map[string]interface{}{
		"location": fmt.Sprintf("line %d:%d", lineNum+1, char+1),
		"type":     "reference",
	})

	return references, word
}

// GotoDefinition finds the definition of a symbol at a specific position
func GotoDefinition(scope *Scope, line string, char int) (map[string]interface{}, string, bool) {
	// Get word at position
	word := getWordAtPosition(line, char)
	if word == "" {
		return nil, "", false
	}

	// Check if it's a built-in function
	for _, fn := range builtinFunctions {
		if strings.EqualFold(fn.Name, word) {
			return map[string]interface{}{
				"type": "built-in",
				"kind": "function",
			}, word, true
		}
	}

	// Check if it's a built-in constant
	for _, const_ := range builtinConstants {
		if strings.EqualFold(const_.Name, word) {
			return map[string]interface{}{
				"type": "built-in",
				"kind": "constant",
			}, word, true
		}
	}

	// Find symbol in scope
	normalizedSymbol := normalizeLabel(word)
	symbol, found := scope.FindSymbol(normalizedSymbol)
	if !found {
		return nil, word, false
	}

	definition := map[string]interface{}{
		"location": fmt.Sprintf("line %d:%d", symbol.Position.Line+1, symbol.Position.Character+1),
		"type":     strings.ToLower(symbol.Kind.String()),
	}

	return definition, word, true
}
