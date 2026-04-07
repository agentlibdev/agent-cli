package targets

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectBuiltInTargetFromPath(t *testing.T) {
	detections, err := DetectWithLookups(t.TempDir(), []Target{
		{ID: "codex", Type: TypeBuiltIn, Format: "codex", Mode: "generate", Enabled: true},
	}, func(name string) (string, error) {
		if name == "codex" {
			return "/usr/local/bin/codex", nil
		}
		return "", errors.New("not found")
	}, os.Stat)
	if err != nil {
		t.Fatalf("DetectWithLookups returned error: %v", err)
	}

	if len(detections) != 1 {
		t.Fatalf("len(detections) = %d, want 1", len(detections))
	}
	if !detections[0].Detected {
		t.Fatalf("Detected = false, want true")
	}
	if detections[0].Path != "/usr/local/bin/codex" {
		t.Fatalf("Path = %q", detections[0].Path)
	}
	if detections[0].Evidence != "command" {
		t.Fatalf("Evidence = %q", detections[0].Evidence)
	}
}

func TestDetectBuiltInTargetFromInstallRoot(t *testing.T) {
	root := t.TempDir()
	installRoot := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(installRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	detections, err := DetectWithLookups(root, []Target{
		{ID: "claude-code", Type: TypeBuiltIn, Format: "claude-code", InstallRoot: installRoot, Mode: "symlink", Enabled: true},
	}, func(string) (string, error) {
		return "", errors.New("not found")
	}, os.Stat)
	if err != nil {
		t.Fatalf("DetectWithLookups returned error: %v", err)
	}

	if len(detections) != 1 {
		t.Fatalf("len(detections) = %d, want 1", len(detections))
	}
	if !detections[0].Detected {
		t.Fatalf("Detected = false, want true")
	}
	if detections[0].Evidence != "installRoot" {
		t.Fatalf("Evidence = %q", detections[0].Evidence)
	}
	if detections[0].Path != installRoot {
		t.Fatalf("Path = %q", detections[0].Path)
	}
}

func TestDetectCustomTargetFromInstallRoot(t *testing.T) {
	root := t.TempDir()
	customRoot := filepath.Join(root, "openclaw")
	if err := os.MkdirAll(customRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	detections, err := DetectWithLookups(root, []Target{
		{ID: "openclaw-local", Type: TypeCustom, Format: "markdown-skill-dir", InstallRoot: customRoot, Mode: "symlink", Enabled: true},
	}, func(string) (string, error) {
		return "", errors.New("not found")
	}, os.Stat)
	if err != nil {
		t.Fatalf("DetectWithLookups returned error: %v", err)
	}

	if len(detections) != 1 {
		t.Fatalf("len(detections) = %d, want 1", len(detections))
	}
	if !detections[0].Detected {
		t.Fatalf("Detected = false, want true")
	}
	if detections[0].Evidence != "installRoot" {
		t.Fatalf("Evidence = %q", detections[0].Evidence)
	}
}

func TestDetectMarksMissingTargets(t *testing.T) {
	detections, err := DetectWithLookups(t.TempDir(), []Target{
		{ID: "openclaw", Type: TypeBuiltIn, Format: "openclaw", Mode: "generate", Enabled: true},
	}, func(string) (string, error) {
		return "", errors.New("not found")
	}, os.Stat)
	if err != nil {
		t.Fatalf("DetectWithLookups returned error: %v", err)
	}

	if len(detections) != 1 {
		t.Fatalf("len(detections) = %d, want 1", len(detections))
	}
	if detections[0].Detected {
		t.Fatalf("Detected = true, want false")
	}
	if !strings.Contains(detections[0].Status, "missing") {
		t.Fatalf("Status = %q", detections[0].Status)
	}
}
