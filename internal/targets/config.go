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
	Name         string `json:"name,omitempty"`
	Type         Type   `json:"type"`
	Format       string `json:"format"`
	RelativePath string `json:"relativePath,omitempty"`
	InstallRoot  string `json:"installRoot,omitempty"`
	ManifestPath string `json:"manifestPath,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Enabled      bool   `json:"enabled"`
	baseDir      string `json:"-"`
}

type fileConfig struct {
	Targets []Target `json:"targets"`
}

func Load(projectDir string) ([]Target, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return nil, fmt.Errorf("HOME is not set")
	}

	items := append([]Target{}, builtIns(home)...)

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

func builtIns(home string) []Target {
	return []Target{
		{
			ID:           "antigravity",
			Name:         "Antigravity",
			Type:         TypeBuiltIn,
			Format:       "antigravity",
			RelativePath: ".gemini/antigravity/skills",
			InstallRoot:  filepath.Join(home, ".gemini", "antigravity", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "claude-code",
			Name:         "Claude Code",
			Type:         TypeBuiltIn,
			Format:       "claude-code",
			RelativePath: ".claude/skills",
			InstallRoot:  filepath.Join(home, ".claude", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "cursor",
			Name:         "Cursor",
			Type:         TypeBuiltIn,
			Format:       "cursor",
			RelativePath: ".cursor/skills",
			InstallRoot:  filepath.Join(home, ".cursor", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "crewai",
			Name:         "CrewAI",
			Type:         TypeBuiltIn,
			Format:       "package-export",
			RelativePath: ".crewai/agents",
			InstallRoot:  filepath.Join(home, ".crewai", "agents"),
			Mode:         "generate",
			Enabled:      true,
		},
		{
			ID:           "codex",
			Name:         "Codex",
			Type:         TypeBuiltIn,
			Format:       "codex",
			RelativePath: ".agents/skills",
			InstallRoot:  filepath.Join(home, ".agents", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "openclaw",
			Name:         "OpenClaw",
			Type:         TypeBuiltIn,
			Format:       "package-export",
			RelativePath: ".openclaw/agents",
			InstallRoot:  filepath.Join(home, ".openclaw", "agents"),
			Mode:         "generate",
			Enabled:      true,
		},
		{
			ID:           "gemini-cli",
			Name:         "Gemini CLI",
			Type:         TypeBuiltIn,
			Format:       "gemini-cli",
			RelativePath: ".gemini/skills",
			InstallRoot:  filepath.Join(home, ".gemini", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "github-copilot",
			Name:         "GitHub Copilot",
			Type:         TypeBuiltIn,
			Format:       "github-copilot",
			RelativePath: ".copilot/skills",
			InstallRoot:  filepath.Join(home, ".copilot", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "langchain",
			Name:         "LangChain",
			Type:         TypeBuiltIn,
			Format:       "package-export",
			RelativePath: ".langchain/agents",
			InstallRoot:  filepath.Join(home, ".langchain", "agents"),
			Mode:         "generate",
			Enabled:      true,
		},
		{
			ID:           "opencode",
			Name:         "OpenCode",
			Type:         TypeBuiltIn,
			Format:       "opencode",
			RelativePath: ".config/opencode/skills",
			InstallRoot:  filepath.Join(home, ".config", "opencode", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "windsurf",
			Name:         "Windsurf",
			Type:         TypeBuiltIn,
			Format:       "windsurf",
			RelativePath: ".codeium/windsurf/skills",
			InstallRoot:  filepath.Join(home, ".codeium", "windsurf", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
		{
			ID:           "vscode",
			Name:         "VS Code",
			Type:         TypeBuiltIn,
			Format:       "vscode",
			RelativePath: ".vscode/agentlib/skills",
			InstallRoot:  filepath.Join(home, ".vscode", "agentlib", "skills"),
			Mode:         "symlink",
			Enabled:      true,
		},
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

	baseDir := filepath.Dir(path)
	for index := range config.Targets {
		config.Targets[index].baseDir = baseDir
		if config.Targets[index].InstallRoot != "" && !filepath.IsAbs(config.Targets[index].InstallRoot) {
			config.Targets[index].InstallRoot = filepath.Join(baseDir, config.Targets[index].InstallRoot)
		}
		if config.Targets[index].ManifestPath != "" && !filepath.IsAbs(config.Targets[index].ManifestPath) {
			config.Targets[index].ManifestPath = filepath.Join(baseDir, config.Targets[index].ManifestPath)
		}
	}

	return config.Targets, nil
}
