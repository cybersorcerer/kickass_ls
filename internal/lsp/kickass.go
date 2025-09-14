package lsp

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"c64.nvim/internal/log"
)

type KickassDirective struct {
	Directive   string   `json:"directive"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

func LoadKickassDirectives(workspaceRoot string) ([]KickassDirective, error) {
	jsonPath := filepath.Join(workspaceRoot, "kickass.json")
	log.Debug("Loading kickass directives from %s", jsonPath)

	file, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var directives []KickassDirective
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

	return directives, nil
}
