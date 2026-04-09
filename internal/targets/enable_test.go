package targets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

func TestEnableTargetCreatesSymlinkIntoInstallRoot(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "codex-skills")
	target := Target{
		ID:          "codex-local",
		Type:        TypeCustom,
		Format:      "markdown-skill-dir",
		InstallRoot: targetDir,
		Mode:        "symlink",
		Enabled:     true,
	}

	result, err := Enable(store, target, ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	info, err := os.Lstat(result.Path)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func TestEnableTargetCopiesPackageIntoInstallRoot(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(filepath.Join(sourceRoot, "assets"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "assets", "guide.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "claude-skills")
	target := Target{
		ID:          "claude-local",
		Type:        TypeCustom,
		Format:      "markdown-skill-dir",
		InstallRoot: targetDir,
		Mode:        "copy",
		Enabled:     true,
	}

	result, err := Enable(store, target, ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(result.Path, "assets", "guide.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "hello\n" {
		t.Fatalf("content = %q", string(content))
	}
}

func TestEnableBuiltInOpenClawGeneratesPackageExport(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	home := t.TempDir()
	target := Target{
		ID:          "openclaw",
		Type:        TypeBuiltIn,
		Format:      "package-export",
		InstallRoot: filepath.Join(home, ".openclaw", "agents"),
		Mode:        "generate",
		Enabled:     true,
	}

	result, err := Enable(store, target, ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	if result.Path != filepath.Join(home, ".openclaw", "agents", "raul", "code-reviewer", "0.4.0") {
		t.Fatalf("Path = %q", result.Path)
	}

	content, err := os.ReadFile(filepath.Join(result.Path, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "hello\n" {
		t.Fatalf("README.md = %q", string(content))
	}

	metaContent, err := os.ReadFile(filepath.Join(result.Path, "agentlib-export.json"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	var meta struct {
		TargetID        string `json:"targetId"`
		SourceRef       string `json:"sourceRef"`
		SourceStorePath string `json:"sourceStorePath"`
		ExportedAt      string `json:"exportedAt"`
		FormatVersion   int    `json:"formatVersion"`
	}
	if err := json.Unmarshal(metaContent, &meta); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if meta.TargetID != "openclaw" || meta.SourceRef != "raul/code-reviewer@0.4.0" || meta.SourceStorePath != sourceRoot {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta.ExportedAt == "" {
		t.Fatal("ExportedAt is empty")
	}
	if meta.FormatVersion != 1 {
		t.Fatalf("FormatVersion = %d, want 1", meta.FormatVersion)
	}
}

func TestEnableBuiltInCrewAIGeneratesPackageExportAndStarterFile(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	home := t.TempDir()
	target := Target{
		ID:          "crewai",
		Type:        TypeBuiltIn,
		Format:      "package-export",
		InstallRoot: filepath.Join(home, ".crewai", "agents"),
		Mode:        "generate",
		Enabled:     true,
	}

	result, err := Enable(store, target, ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(result.Path, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "hello\n" {
		t.Fatalf("README.md = %q", string(content))
	}

	starterContent, err := os.ReadFile(filepath.Join(result.Path, "crewai-agent.py"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	starter := string(starterContent)
	if !strings.Contains(starter, "raul/code-reviewer@0.4.0") || !strings.Contains(starter, "AgentLib CrewAI export") {
		t.Fatalf("starter = %q", starter)
	}
}

func TestEnableBuiltInLangChainGeneratesPackageExportAndStarterFile(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	home := t.TempDir()
	target := Target{
		ID:          "langchain",
		Type:        TypeBuiltIn,
		Format:      "package-export",
		InstallRoot: filepath.Join(home, ".langchain", "agents"),
		Mode:        "generate",
		Enabled:     true,
	}

	result, err := Enable(store, target, ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(result.Path, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "hello\n" {
		t.Fatalf("README.md = %q", string(content))
	}

	starterContent, err := os.ReadFile(filepath.Join(result.Path, "langchain-agent.py"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	starter := string(starterContent)
	if !strings.Contains(starter, "raul/code-reviewer@0.4.0") || !strings.Contains(starter, "AgentLib LangChain export") {
		t.Fatalf("starter = %q", starter)
	}
}

func TestEnableTargetRejectsMissingInstallRoot(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")

	_, err := Enable(store, Target{
		ID:      "codex",
		Type:    TypeBuiltIn,
		Format:  "codex",
		Mode:    "generate",
		Enabled: true,
	}, ref)
	if err == nil {
		t.Fatal("Enable returned nil error, want error")
	}
}

func TestEnableTargetRejectsMissingInstalledPackage(t *testing.T) {
	store := t.TempDir()
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	targetDir := filepath.Join(t.TempDir(), "codex-skills")

	_, err := Enable(store, Target{
		ID:          "codex-local",
		Type:        TypeCustom,
		Format:      "markdown-skill-dir",
		InstallRoot: targetDir,
		Mode:        "symlink",
		Enabled:     true,
	}, ref)
	if err == nil {
		t.Fatal("Enable returned nil error, want error")
	}
}

func TestEnableBuiltInCodexWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	items, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	store := filepath.Join(home, ".agentlib")
	ref := mustRef(t, "raul/code-reviewer@0.4.0")
	sourceRoot := filepath.Join(store, "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Enable(store, findByID(items, "codex"), ref)
	if err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}

	info, err := os.Lstat(result.Path)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
	if result.Path != filepath.Join(home, ".agents", "skills", "raul", "code-reviewer", "0.4.0") {
		t.Fatalf("Path = %q", result.Path)
	}
}

func mustRef(t *testing.T, value string) agentref.Ref {
	t.Helper()
	ref, err := agentref.Parse(value)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	return ref
}
