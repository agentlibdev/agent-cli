package targets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

func TestDisableRemovesMaterializedTargetPath(t *testing.T) {
	targetRoot := filepath.Join(t.TempDir(), "codex-skills")
	ref := mustActivationRef(t, "raul/code-reviewer@0.4.0")
	targetPath := filepath.Join(targetRoot, ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetPath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Disable(Target{
		ID:          "codex",
		Type:        TypeBuiltIn,
		Format:      "codex",
		InstallRoot: targetRoot,
		Mode:        "symlink",
		Enabled:     true,
	}, ref)
	if err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}
	if result.Path != targetPath {
		t.Fatalf("result.Path = %q, want %q", result.Path, targetPath)
	}
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("target path still exists: %v", err)
	}
}

func mustActivationRef(t *testing.T, value string) agentref.Ref {
	t.Helper()
	ref, err := agentref.Parse(value)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	return ref
}
