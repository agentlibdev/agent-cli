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

	if len(items) != 4 {
		t.Fatalf("len(items) = %d, want 4", len(items))
	}
	if items[0].ID != "claude" && items[0].ID != "codex" && items[0].ID != "gemini-cli" && items[0].ID != "openclaw" {
		t.Fatalf("unexpected first target: %+v", items[0])
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
