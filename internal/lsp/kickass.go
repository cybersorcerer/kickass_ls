package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"c64.nvim/internal/log"
)

type KickassDirective struct {
	Directive   string   `json:"directive"`
	Signature   string   `json:"signature"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

func LoadKickassDirectives(workspaceRoot string) ([]KickassDirective, error) {
	jsonPath := filepath.Join(workspaceRoot, "kickass.json")
	log.Debug("Loading kickass directives from %s", jsonPath)

	file, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var directives []KickassDirective

	// First try to parse as new structure with "directives" field
	var config struct {
		Directives []KickassDirective `json:"directives"`
	}
	err = json.Unmarshal(file, &config)
	if err == nil && len(config.Directives) > 0 {
		directives = config.Directives
	} else {
		// Fallback: try unmarshalling as array
		err = json.Unmarshal(file, &directives)
		if err != nil {
			// Try unmarshalling as a single object
			var singleDirective KickassDirective
			err2 := json.Unmarshal(file, &singleDirective)
			if err2 != nil {
				return nil, err // Return original error
			}
			directives = []KickassDirective{singleDirective}
		}
	}

	return directives, nil
}
