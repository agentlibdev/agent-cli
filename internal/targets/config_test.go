package targets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReturnsBuiltInsWhenNoConfigExists(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	items, err := Load(project)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(items) != 8 {
		t.Fatalf("len(items) = %d, want 8", len(items))
	}
	if !containsID(items, "antigravity") || !containsID(items, "codex") || !containsID(items, "windsurf") {
		t.Fatalf("items missing expected built-ins: %+v", items)
	}

	codex := findByID(items, "codex")
	if codex.InstallRoot != filepath.Join(home, ".agents", "skills") {
		t.Fatalf("codex.InstallRoot = %q", codex.InstallRoot)
	}
}

func TestLoadIncludesGlobalCustomTargets(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".agentlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".agentlib", "targets.json"), []byte(`{"targets":[{"id":"custom-openclaw","type":"custom","format":"markdown-skill-dir","installRoot":"/tmp/openclaw","mode":"symlink","enabled":true}]}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	items, err := Load(project)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !containsID(items, "custom-openclaw") {
		t.Fatalf("items missing custom-openclaw: %+v", items)
	}
}

func TestLoadIncludesProjectCustomTargets(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(project, ".agentlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agentlib", "targets.json"), []byte(`{"targets":[{"id":"project-openclaw","type":"custom","format":"json-manifest","installRoot":"./vendor/openclaw","mode":"generate","enabled":false}]}`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	items, err := Load(project)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !containsID(items, "project-openclaw") {
		t.Fatalf("items missing project-openclaw: %+v", items)
	}
}

func TestLoadReturnsInvalidJSONError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".agentlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".agentlib", "targets.json"), []byte(`{`), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := Load(project)
	if err == nil {
		t.Fatal("Load returned nil error, want error")
	}
}

func containsID(items []Target, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func findByID(items []Target, id string) Target {
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	return Target{}
}
