package targets

import (
	"os"
	"path/filepath"
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

func mustRef(t *testing.T, value string) agentref.Ref {
	t.Helper()
	ref, err := agentref.Parse(value)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	return ref
}
