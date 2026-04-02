package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTargetDefaultsToGlobalStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	target, err := ResolveTarget(TargetOptions{
		WorkingDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("ResolveTarget returned error: %v", err)
	}

	if target.Mode != ModeGlobal {
		t.Fatalf("Mode = %q, want %q", target.Mode, ModeGlobal)
	}
	if target.Root != filepath.Join(home, ".agentlib") {
		t.Fatalf("Root = %q", target.Root)
	}
}

func TestResolveTargetUsesLocalProjectRoot(t *testing.T) {
	project := t.TempDir()
	markerDir := filepath.Join(project, ".agentlib")
	if err := os.MkdirAll(markerDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(markerDir, "project.json"), []byte("{\"version\":1}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	target, err := ResolveTarget(TargetOptions{
		WorkingDir: project,
		Local:      true,
	})
	if err != nil {
		t.Fatalf("ResolveTarget returned error: %v", err)
	}

	if target.Mode != ModeLocal {
		t.Fatalf("Mode = %q, want %q", target.Mode, ModeLocal)
	}
	if target.Root != filepath.Join(project, ".agentlib") {
		t.Fatalf("Root = %q", target.Root)
	}
}

func TestResolveTargetUsesExplicitLocalInstallDir(t *testing.T) {
	project := t.TempDir()
	target, err := ResolveTarget(TargetOptions{
		WorkingDir: project,
		Local:      true,
		InstallDir: "vendor/agentlib",
	})
	if err != nil {
		t.Fatalf("ResolveTarget returned error: %v", err)
	}

	if target.Mode != ModeLocal {
		t.Fatalf("Mode = %q, want %q", target.Mode, ModeLocal)
	}
	if target.Root != filepath.Join(project, "vendor/agentlib") {
		t.Fatalf("Root = %q", target.Root)
	}
}

func TestResolveTargetRejectsLocalWithoutProject(t *testing.T) {
	_, err := ResolveTarget(TargetOptions{
		WorkingDir: t.TempDir(),
		Local:      true,
	})
	if err == nil {
		t.Fatal("ResolveTarget returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "agentlib init") {
		t.Fatalf("error = %q, want init guidance", err.Error())
	}
}

func TestResolveTargetRejectsConflictingModes(t *testing.T) {
	_, err := ResolveTarget(TargetOptions{
		WorkingDir: t.TempDir(),
		Local:      true,
		Global:     true,
	})
	if err == nil {
		t.Fatal("ResolveTarget returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "cannot use --local and --global together") {
		t.Fatalf("error = %q", err.Error())
	}
}
