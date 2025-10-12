package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	log "c64.nvim/internal/log"
)

// JSON loading implementation for ProcessorContext

// loadMnemonics loads 6510 mnemonics from mnemonic.json
func (ctx *ProcessorContext) loadMnemonics(path string) error {
	log.Debug("Loading mnemonics from %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read mnemonic file %s: %v", path, err)
	}

	var mnemonics []*EnhancedMnemonicInfo
	if err := json.Unmarshal(data, &mnemonics); err != nil {
		return fmt.Errorf("failed to parse mnemonic JSON: %v", err)
	}

	// Categorize mnemonics by type
	standardCount := 0
	illegalCount := 0
	controlCount := 0

	for _, mnemonic := range mnemonics {
		name := strings.ToUpper(mnemonic.Name)

		// Determine category based on type or name patterns
		if mnemonic.Type == "Illegal" || isIllegalOpcode(name) {
			ctx.IllegalMnemonics[name] = mnemonic
			illegalCount++
		} else if mnemonic.Type == "Jump" || isControlOpcode(name) {
			ctx.ControlMnemonics[name] = mnemonic
			controlCount++
		} else {
			ctx.StandardMnemonics[name] = mnemonic
			standardCount++
		}

		// Also add to combined cache
		ctx.AllMnemonics[name] = mnemonic
	}

	log.Info("Loaded %d mnemonics: %d standard, %d illegal, %d control",
		len(mnemonics), standardCount, illegalCount, controlCount)

	return nil
}

// loadKickAssemblerData loads directives, functions, and constants from kickass.json
func (ctx *ProcessorContext) loadKickAssemblerData(path string) error {
	log.Debug("Loading Kick Assembler data from %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read kickass file %s: %v", path, err)
	}

	var kickassData struct {
		Directives             []KickDirectiveInfo `json:"directives"`
		PreprocessorStatements []KickDirectiveInfo `json:"preprocessorStatements"`
		Functions              []FunctionInfo      `json:"functions"`
		Constants              []ConstantInfo      `json:"constants"`
	}

	if err := json.Unmarshal(data, &kickassData); err != nil {
		return fmt.Errorf("failed to parse kickass JSON: %v", err)
	}

	// Load directives
	for _, directive := range kickassData.Directives {
		name := strings.ToLower(directive.Name)
		// Don't add prefix if it already has . or # prefix
		if !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "#") {
			name = "." + name
		}
		directive.SourceType = SourceDirective
		ctx.Directives[name] = &directive
		ctx.DirectiveNames = append(ctx.DirectiveNames, name)
	}

	// Load preprocessor statements
	for _, directive := range kickassData.PreprocessorStatements {
		name := strings.ToLower(directive.Name)
		directive.SourceType = SourcePreprocessor
		ctx.PreprocessorStatements[name] = &directive
	}

	// Load functions
	for _, function := range kickassData.Functions {
		name := strings.ToLower(function.Name)
		ctx.Functions[name] = &function
		ctx.FunctionNames = append(ctx.FunctionNames, name)
	}

	// Load constants
	for _, constant := range kickassData.Constants {
		name := strings.ToUpper(constant.Name)
		ctx.Constants[name] = &constant
		ctx.ConstantNames = append(ctx.ConstantNames, name)
	}

	log.Info("Loaded Kick Assembler data: %d directives, %d preprocessor statements, %d functions, %d constants",
		len(ctx.Directives), len(ctx.PreprocessorStatements), len(ctx.Functions), len(ctx.Constants))

	return nil
}

// loadC64Memory loads C64 memory map from c64memory.json
func (ctx *ProcessorContext) loadC64Memory(path string) error {
	log.Debug("Loading C64 memory map from %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read c64memory file %s: %v", path, err)
	}

	var memoryData struct {
		MemoryMap struct {
			Version string                       `json:"version"`
			Source  string                       `json:"source"`
			Regions map[string]json.RawMessage   `json:"regions"`
		} `json:"memoryMap"`
	}

	if err := json.Unmarshal(data, &memoryData); err != nil {
		return fmt.Errorf("failed to parse c64memory JSON: %v", err)
	}

	// Parse memory regions
	for addrStr, regionData := range memoryData.MemoryMap.Regions {
		// Parse address
		addr, err := parseAddress(addrStr)
		if err != nil {
			log.Warn("Failed to parse address %s: %v", addrStr, err)
			continue
		}

		// Parse region data
		var region MemoryRegion
		if err := json.Unmarshal(regionData, &region); err != nil {
			log.Warn("Failed to parse region data for %s: %v", addrStr, err)
			continue
		}

		region.Address = addr
		ctx.MemoryRegions = append(ctx.MemoryRegions, &region)

		// Add to memory map for fast lookup
		// Map all addresses covered by this region
		for i := 0; i < region.Size; i++ {
			ctx.MemoryMap[addr+uint16(i)] = &region
		}
	}

	log.Info("Loaded C64 memory map: %d regions covering %d addresses",
		len(ctx.MemoryRegions), len(ctx.MemoryMap))

	return nil
}

// buildCaches builds lookup caches for performance
func (ctx *ProcessorContext) buildCaches() {
	log.Debug("Building processor context caches")

	// Sort names for better completion order
	sortStringSlice(ctx.DirectiveNames)
	sortStringSlice(ctx.FunctionNames)
	sortStringSlice(ctx.ConstantNames)

	log.Debug("Built caches: %d directive names, %d function names, %d constant names",
		len(ctx.DirectiveNames), len(ctx.FunctionNames), len(ctx.ConstantNames))
}

// Helper functions

// isIllegalOpcode checks if a mnemonic is an illegal 6510 opcode
func isIllegalOpcode(name string) bool {
	illegalOpcodes := []string{
		"AHX", "ALR", "ANC", "ARR", "AXS", "DCP", "ISC", "LAS", "LAX", "RLA", "RRA", "SAX", "SHX", "SHY", "SLO", "SRE", "TAS", "XAA",
	}
	for _, illegal := range illegalOpcodes {
		if name == illegal {
			return true
		}
	}
	return false
}

// isControlOpcode checks if a mnemonic is a control flow opcode
func isControlOpcode(name string) bool {
	controlOpcodes := []string{
		"BCC", "BCS", "BEQ", "BMI", "BNE", "BPL", "BVC", "BVS", "JMP", "JSR", "RTS", "RTI",
	}
	for _, control := range controlOpcodes {
		if name == control {
			return true
		}
	}
	return false
}

// parseAddress parses a hex address string (e.g., "0x1000" or "$1000")
func parseAddress(addrStr string) (uint16, error) {
	// Remove common prefixes
	addrStr = strings.TrimPrefix(addrStr, "0x")
	addrStr = strings.TrimPrefix(addrStr, "0X")
	addrStr = strings.TrimPrefix(addrStr, "$")

	// Parse as hex
	addr, err := strconv.ParseUint(addrStr, 16, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid address format: %s", addrStr)
	}

	return uint16(addr), nil
}

// sortStringSlice sorts a string slice in place (case-insensitive)
func sortStringSlice(slice []string) {
	// Simple case-insensitive sort
	for i := 0; i < len(slice)-1; i++ {
		for j := i + 1; j < len(slice); j++ {
			if strings.ToLower(slice[i]) > strings.ToLower(slice[j]) {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}

// Context helper functions for enhanced analysis

// GetMnemonicInfo looks up mnemonic information by name
func (ctx *ProcessorContext) GetMnemonicInfo(name string) *EnhancedMnemonicInfo {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	name = strings.ToUpper(name)
	return ctx.AllMnemonics[name]
}

// GetDirectiveInfo looks up directive information by name
func (ctx *ProcessorContext) GetDirectiveInfo(name string) *KickDirectiveInfo {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	name = strings.ToLower(name)

	// Check preprocessor statements first (they start with #)
	if strings.HasPrefix(name, "#") {
		if info, found := ctx.PreprocessorStatements[name]; found {
			return info
		}
	}

	// Don't add prefix if it already has . or # prefix
	if !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "#") {
		name = "." + name
	}
	return ctx.Directives[name]
}

// GetFunctionInfo looks up function information by name
func (ctx *ProcessorContext) GetFunctionInfo(name string) *FunctionInfo {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	name = strings.ToLower(name)
	return ctx.Functions[name]
}

// GetConstantInfo looks up constant information by name
func (ctx *ProcessorContext) GetConstantInfo(name string) *ConstantInfo {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	name = strings.ToUpper(name)
	return ctx.Constants[name]
}

// GetMemoryRegion looks up memory region by address
func (ctx *ProcessorContext) GetMemoryRegion(address uint16) *MemoryRegion {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	return ctx.MemoryMap[address]
}

// IsValidMnemonic checks if a name is a valid 6510 mnemonic
func (ctx *ProcessorContext) IsValidMnemonic(name string) bool {
	return ctx.GetMnemonicInfo(name) != nil
}

// IsIllegalMnemonic checks if a mnemonic is illegal
func (ctx *ProcessorContext) IsIllegalMnemonic(name string) bool {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	name = strings.ToUpper(name)
	_, exists := ctx.IllegalMnemonics[name]
	return exists
}

// IsValidDirective checks if a name is a valid Kick Assembler directive
func (ctx *ProcessorContext) IsValidDirective(name string) bool {
	return ctx.GetDirectiveInfo(name) != nil
}

// IsValidFunction checks if a name is a valid Kick Assembler function
func (ctx *ProcessorContext) IsValidFunction(name string) bool {
	return ctx.GetFunctionInfo(name) != nil
}

// IsValidConstant checks if a name is a valid Kick Assembler constant
func (ctx *ProcessorContext) IsValidConstant(name string) bool {
	return ctx.GetConstantInfo(name) != nil
}

// GetAddressingModeForMnemonic returns valid addressing modes for a mnemonic
func (ctx *ProcessorContext) GetAddressingModeForMnemonic(mnemonic string, operand string) *AddressingModeInfo {
	info := ctx.GetMnemonicInfo(mnemonic)
	if info == nil {
		return nil
	}

	// Analyze operand to determine addressing mode
	operand = strings.TrimSpace(operand)

	for _, mode := range info.AddressingModes {
		if ctx.matchesAddressingMode(operand, mode.Mode) {
			return mode
		}
	}

	return nil
}

// matchesAddressingMode checks if an operand matches a specific addressing mode
func (ctx *ProcessorContext) matchesAddressingMode(operand string, mode string) bool {
	switch mode {
	case "Immediate":
		return strings.HasPrefix(operand, "#")
	case "Zero Page":
		return ctx.isZeroPageAddress(operand)
	case "Zero Page,X":
		return strings.HasSuffix(operand, ",X") && ctx.isZeroPageAddress(strings.TrimSuffix(operand, ",X"))
	case "Zero Page,Y":
		return strings.HasSuffix(operand, ",Y") && ctx.isZeroPageAddress(strings.TrimSuffix(operand, ",Y"))
	case "Absolute":
		return ctx.isAbsoluteAddress(operand) && !strings.Contains(operand, ",")
	case "Absolute,X":
		return strings.HasSuffix(operand, ",X") && ctx.isAbsoluteAddress(strings.TrimSuffix(operand, ",X"))
	case "Absolute,Y":
		return strings.HasSuffix(operand, ",Y") && ctx.isAbsoluteAddress(strings.TrimSuffix(operand, ",Y"))
	case "Indirect":
		return strings.HasPrefix(operand, "(") && strings.HasSuffix(operand, ")")
	case "Indirect,X":
		return strings.HasPrefix(operand, "(") && strings.HasSuffix(operand, ",X)")
	case "Indirect,Y":
		return strings.HasPrefix(operand, "(") && strings.HasSuffix(operand, "),Y")
	case "Relative":
		return true // Relative addresses are context-dependent
	case "Implied":
		return operand == ""
	default:
		return false
	}
}

// isZeroPageAddress checks if an address is in zero page range ($00-$FF)
func (ctx *ProcessorContext) isZeroPageAddress(addr string) bool {
	addr = strings.TrimSpace(addr)

	// Parse numeric address
	if value, err := ctx.parseNumericValue(addr); err == nil {
		return value >= 0 && value <= 0xFF
	}

	// Could be a label - assume it might be zero page
	return true
}

// isAbsoluteAddress checks if an address is absolute (16-bit)
func (ctx *ProcessorContext) isAbsoluteAddress(addr string) bool {
	addr = strings.TrimSpace(addr)

	// Parse numeric address
	if value, err := ctx.parseNumericValue(addr); err == nil {
		return value >= 0 && value <= 0xFFFF
	}

	// Could be a label - assume it's absolute
	return true
}

// parseNumericValue parses a numeric value (hex, decimal, binary)
func (ctx *ProcessorContext) parseNumericValue(value string) (int, error) {
	value = strings.TrimSpace(value)

	// Remove leading #
	value = strings.TrimPrefix(value, "#")

	// Hex values
	if strings.HasPrefix(value, "$") {
		return strconv.Atoi("0x" + value[1:])
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		return strconv.Atoi(value)
	}

	// Binary values
	if strings.HasPrefix(value, "%") {
		val, err := strconv.ParseInt(value[1:], 2, 32)
		return int(val), err
	}

	// Decimal values
	if len(value) > 0 && unicode.IsDigit(rune(value[0])) {
		return strconv.Atoi(value)
	}

	return 0, fmt.Errorf("not a numeric value: %s", value)
}

// Integration with existing server initialization

// InitializeProcessorContext loads the processor context from config directory
func InitializeProcessorContext(configDir string) error {
	mnemonicPath := configDir + "/mnemonic.json"
	kickassPath := configDir + "/kickass.json"
	c64MemoryPath := configDir + "/c64memory.json"

	return LoadProcessorContext(mnemonicPath, kickassPath, c64MemoryPath)
}