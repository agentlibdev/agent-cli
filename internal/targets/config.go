package targets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Type string

const (
	TypeBuiltIn Type = "built-in"
	TypeCustom  Type = "custom"
)

type Target struct {
	ID           string `json:"id"`
	Type         Type   `json:"type"`
	Format       string `json:"format"`
	InstallRoot  string `json:"installRoot,omitempty"`
	ManifestPath string `json:"manifestPath,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Enabled      bool   `json:"enabled"`
}

type fileConfig struct {
	Targets []Target `json:"targets"`
}

func Load(projectDir string) ([]Target, error) {
	items := append([]Target{}, builtIns()...)

	home := os.Getenv("HOME")
	if home == "" {
		return nil, fmt.Errorf("HOME is not set")
	}

	globalItems, err := loadFile(filepath.Join(home, ".agentlib", "targets.json"))
	if err != nil {
		return nil, err
	}
	items = append(items, globalItems...)

	projectItems, err := loadFile(filepath.Join(projectDir, ".agentlib", "targets.json"))
	if err != nil {
		return nil, err
	}
	items = append(items, projectItems...)

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

func builtIns() []Target {
	return []Target{
		{ID: "claude", Type: TypeBuiltIn, Format: "claude", Mode: "generate", Enabled: true},
		{ID: "codex", Type: TypeBuiltIn, Format: "codex", Mode: "generate", Enabled: true},
		{ID: "gemini-cli", Type: TypeBuiltIn, Format: "gemini-cli", Mode: "generate", Enabled: true},
		{ID: "openclaw", Type: TypeBuiltIn, Format: "openclaw", Mode: "generate", Enabled: true},
	}
}

func loadFile(path string) ([]Target, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config fileConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parse targets config %s: %w", path, err)
	}

	return config.Targets, nil
}
