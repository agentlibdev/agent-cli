package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/install"
	"github.com/agentlibdev/agent-cli/internal/registry"
	"github.com/agentlibdev/agent-cli/internal/targets"
)

func TestRunSearchFiltersAgents(t *testing.T) {
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{
				agents: []registry.AgentSummary{
					{Namespace: "raul", Name: "code-reviewer", LatestVersion: "0.4.0", Title: "Code Reviewer", Description: "Reviews code changes."},
					{Namespace: "raul", Name: "docs-writer", LatestVersion: "0.2.0", Title: "Docs Writer", Description: "Drafts documentation."},
				},
			}
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"search", "docs"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "raul/docs-writer@0.2.0") {
		t.Fatalf("stdout = %q, want docs-writer result", stdout.String())
	}
	if strings.Contains(stdout.String(), "raul/code-reviewer@0.4.0") {
		t.Fatalf("stdout = %q, did not expect code-reviewer result", stdout.String())
	}
}

func TestRunVersionPrintsVersion(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder

	exitCode := Run(context.Background(), []string{"version"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != "agentlib dev\n" {
		t.Fatalf("stdout = %q, want %q", got, "agentlib dev\n")
	}
}

func TestRunRemoveDeletesInstalledVersion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	installedPath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(installedPath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	lockfilePath := filepath.Join(home, ".agentlib", "agent.lock.json")
	if err := os.WriteFile(lockfilePath, []byte("{\"version\":1,\"agent\":{\"namespace\":\"raul\",\"name\":\"code-reviewer\",\"version\":\"0.4.0\"}}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"remove", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed: raul/code-reviewer@0.4.0") {
		t.Fatalf("stdout = %q, want remove confirmation", stdout.String())
	}
	if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
		t.Fatalf("installed path still exists: %v", err)
	}
	if _, err := os.Stat(lockfilePath); !os.IsNotExist(err) {
		t.Fatalf("lockfile still exists: %v", err)
	}
}

func TestRunStatusPrintsInstallAndActivationState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	targetPath := filepath.Join(home, ".agents", "skills", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".agentlib", "config.json"), []byte("{\"version\":1,\"activations\":[{\"targetId\":\"codex\",\"ref\":\"raul/code-reviewer@0.4.0\",\"path\":\""+targetPath+"\",\"activatedAt\":\"2026-04-14T10:00:00Z\"}]}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := (app{}).Run(context.Background(), []string{"status", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "installed: yes") {
		t.Fatalf("stdout = %q, want installed yes", output)
	}
	if !strings.Contains(output, "store: "+storePath) {
		t.Fatalf("stdout = %q, want store path", output)
	}
	if !strings.Contains(output, "active targets: 1") || !strings.Contains(output, "codex") {
		t.Fatalf("stdout = %q, want activation summary", output)
	}
}

func TestRunStatusPrintsMissingWhenPackageIsNotInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := (app{}).Run(context.Background(), []string{"status", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "installed: no") {
		t.Fatalf("stdout = %q, want installed no", output)
	}
	if !strings.Contains(output, "active targets: 0") {
		t.Fatalf("stdout = %q, want zero activations", output)
	}
}

func TestRunActivationsListPrintsPersistedRows(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".agentlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	configBody := `{"version":1,"activations":[{"targetId":"codex","ref":"raul/code-reviewer@0.4.0","path":"/tmp/codex","activatedAt":"2026-04-14T10:00:00Z"},{"targetId":"claude-code","ref":"raul/code-reviewer@0.4.0","path":"/tmp/claude","activatedAt":"2026-04-14T10:05:00Z"}]}`
	if err := os.WriteFile(filepath.Join(home, ".agentlib", "config.json"), []byte(configBody+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := (app{}).Run(context.Background(), []string{"activations", "list"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "codex\traul/code-reviewer@0.4.0\t/tmp/codex") {
		t.Fatalf("stdout = %q, want codex row", output)
	}
	if !strings.Contains(output, "claude-code\traul/code-reviewer@0.4.0\t/tmp/claude") {
		t.Fatalf("stdout = %q, want claude row", output)
	}
}

func TestRunInstallDefaultsToGlobalTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
	}
	exitCode := cli.Run(context.Background(), []string{"install", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	installedPath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if _, err := os.Stat(installedPath); err != nil {
		t.Fatalf("installed path missing: %v", err)
	}
}

func TestRunInstallLocalRejectsUninitializedProject(t *testing.T) {
	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
	}
	exitCode := cli.Run(context.Background(), []string{"install", "--local", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Run exitCode = %d, want 1, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "agentlib init") {
		t.Fatalf("stderr = %q, want init guidance", stderr.String())
	}
}

func TestRunInitCreatesProjectMarker(t *testing.T) {
	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"init"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	projectFile := filepath.Join(root, ".agentlib", "project.json")
	if _, err := os.Stat(projectFile); err != nil {
		t.Fatalf("project marker missing: %v", err)
	}
}

func TestRunInstallLocalWithInstallDirUsesOverride(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(projectDir, ".agentlib"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".agentlib", "project.json"), []byte("{\"version\":1}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
	}
	exitCode := cli.Run(context.Background(), []string{"install", "--local", "--install-dir", "vendor/agentlib", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	installedPath := filepath.Join(projectDir, "vendor/agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if _, err := os.Stat(installedPath); err != nil {
		t.Fatalf("installed path missing: %v", err)
	}
}

func TestRunInstallPromptsForDetectedRuntimesAndActivatesSelectedTargets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	enabled := make([]string, 0, 2)
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
		stdin: strings.NewReader("1,2\n"),
		isInteractiveInput: func() bool {
			return true
		},
		detectTargets: func(string) ([]targets.Detection, error) {
			return []targets.Detection{
				{
					Target:   targets.Target{ID: "codex", Name: "Codex", Type: targets.TypeBuiltIn, Format: "codex", Mode: "symlink", Enabled: true},
					Detected: true,
					Status:   "detected",
				},
				{
					Target:   targets.Target{ID: "claude-code", Name: "Claude Code", Type: targets.TypeBuiltIn, Format: "claude-code", Mode: "symlink", Enabled: true},
					Detected: true,
					Status:   "detected",
				},
			}, nil
		},
		enableTarget: func(_ string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error) {
			enabled = append(enabled, target.ID+":"+ref.String())
			return targets.EnableResult{Path: "/tmp/" + target.ID}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"install", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	if len(enabled) != 2 {
		t.Fatalf("len(enabled) = %d, want 2 (%v)", len(enabled), enabled)
	}
	if !strings.Contains(stdout.String(), "Select runtimes to activate") {
		t.Fatalf("stdout = %q, want prompt", stdout.String())
	}
	if !strings.Contains(stdout.String(), "activated: codex") || !strings.Contains(stdout.String(), "activated: claude-code") {
		t.Fatalf("stdout = %q, want activation summary", stdout.String())
	}
}

func TestRunInstallPromptFiltersDetectedRuntimesByVersionCompatibility(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	enabled := make([]string, 0, 2)
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{
				version: registry.Version{
					Namespace: "raul",
					Name:      "code-reviewer",
					Version:   "0.4.0",
					Compatibility: registry.Compatibility{
						Targets: []registry.TargetCompatibility{
							{TargetID: "codex", BuiltFor: true},
							{TargetID: "claude-code", BuiltFor: false, Tested: false, AdapterAvailable: false},
						},
					},
				},
			}
		},
		stdin: strings.NewReader("1\n"),
		isInteractiveInput: func() bool {
			return true
		},
		detectTargets: func(string) ([]targets.Detection, error) {
			return []targets.Detection{
				{
					Target:   targets.Target{ID: "codex", Name: "Codex", Type: targets.TypeBuiltIn, Format: "codex", Mode: "symlink", Enabled: true},
					Detected: true,
					Status:   "detected",
				},
				{
					Target:   targets.Target{ID: "claude-code", Name: "Claude Code", Type: targets.TypeBuiltIn, Format: "claude-code", Mode: "symlink", Enabled: true},
					Detected: true,
					Status:   "detected",
				},
			}, nil
		},
		enableTarget: func(_ string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error) {
			enabled = append(enabled, target.ID+":"+ref.String())
			return targets.EnableResult{Path: "/tmp/" + target.ID}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"install", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	if len(enabled) != 1 || enabled[0] != "codex:raul/code-reviewer@0.4.0" {
		t.Fatalf("enabled = %v, want only codex", enabled)
	}
	if !strings.Contains(stdout.String(), "1. Codex (codex)") {
		t.Fatalf("stdout = %q, want filtered codex prompt", stdout.String())
	}
	if strings.Contains(stdout.String(), "Claude Code") {
		t.Fatalf("stdout = %q, did not expect incompatible target in prompt", stdout.String())
	}
}

func TestRunInstallNoActivateSkipsPromptAndActivation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	prompted := false
	activated := false
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
		stdin: strings.NewReader("1\n"),
		isInteractiveInput: func() bool {
			prompted = true
			return true
		},
		detectTargets: func(string) ([]targets.Detection, error) {
			t.Fatal("detectTargets should not be called with --no-activate")
			return nil, nil
		},
		enableTarget: func(_ string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error) {
			activated = true
			return targets.EnableResult{}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"install", "--no-activate", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if activated {
		t.Fatal("activation ran, want skipped")
	}
	if strings.Contains(stdout.String(), "Select runtimes to activate") {
		t.Fatalf("stdout = %q, did not expect prompt", stdout.String())
	}
	if prompted {
		t.Fatal("interactive detection should not run with --no-activate")
	}
}

func TestRunInstallNonInteractiveSkipsPromptAndActivation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	activated := false
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
		isInteractiveInput: func() bool {
			return false
		},
		detectTargets: func(string) ([]targets.Detection, error) {
			t.Fatal("detectTargets should not be called for non-interactive install without explicit runtime")
			return nil, nil
		},
		enableTarget: func(_ string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error) {
			activated = true
			return targets.EnableResult{}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"install", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if activated {
		t.Fatal("activation ran, want skipped")
	}
	if strings.Contains(stdout.String(), "Select runtimes to activate") {
		t.Fatalf("stdout = %q, did not expect prompt", stdout.String())
	}
}

func TestRunInstallExplicitRuntimeActivatesWithoutPrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := t.TempDir()
	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWorkingDir)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	activated := ""
	cli := app{
		newRegistryClient: func(string) registryClient {
			return fakeRegistryClient{}
		},
		isInteractiveInput: func() bool {
			t.Fatal("interactive detection should not run for explicit runtime")
			return false
		},
		loadTargets: func(string) ([]targets.Target, error) {
			return []targets.Target{
				{ID: "codex", Name: "Codex", Type: targets.TypeBuiltIn, Format: "codex", Mode: "symlink", Enabled: true},
				{ID: "claude-code", Name: "Claude Code", Type: targets.TypeBuiltIn, Format: "claude-code", Mode: "symlink", Enabled: true},
			}, nil
		},
		enableTarget: func(_ string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error) {
			activated = target.ID
			return targets.EnableResult{Path: "/tmp/" + target.ID}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"install", "--runtime", "codex", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if activated != "codex" {
		t.Fatalf("activated = %q, want codex", activated)
	}
	if strings.Contains(stdout.String(), "Select runtimes to activate") {
		t.Fatalf("stdout = %q, did not expect prompt", stdout.String())
	}
}

type fakeRegistryClient struct {
	agents  []registry.AgentSummary
	version registry.Version
}

func TestRunTargetsListPrintsBuiltInsAndCustomTargets(t *testing.T) {
	cli := app{
		loadTargets: func(string) ([]targets.Target, error) {
			return []targets.Target{
				{ID: "codex", Type: targets.TypeBuiltIn, Format: "codex", Mode: "generate", Enabled: true},
				{ID: "custom-openclaw", Type: targets.TypeCustom, Format: "markdown-skill-dir", Mode: "symlink", Enabled: false},
			}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"targets", "list"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "codex") {
		t.Fatalf("stdout = %q, want codex", output)
	}
	if !strings.Contains(output, "custom-openclaw") {
		t.Fatalf("stdout = %q, want custom target", output)
	}
	if !strings.Contains(output, "disabled") {
		t.Fatalf("stdout = %q, want disabled marker", output)
	}
}

func TestRunTargetsDetectPrintsDetectedAndMissingTargets(t *testing.T) {
	cli := app{
		detectTargets: func(string) ([]targets.Detection, error) {
			return []targets.Detection{
				{
					Target:   targets.Target{ID: "codex", Type: targets.TypeBuiltIn, Format: "codex", Mode: "generate", Enabled: true},
					Detected: true,
					Status:   "detected",
					Path:     "/usr/local/bin/codex",
					Evidence: "command",
				},
				{
					Target:   targets.Target{ID: "openclaw", Type: targets.TypeBuiltIn, Format: "openclaw", Mode: "generate", Enabled: true},
					Detected: false,
					Status:   "missing",
				},
			}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"targets", "detect"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "codex") || !strings.Contains(output, "/usr/local/bin/codex") {
		t.Fatalf("stdout = %q, want detected codex", output)
	}
	if !strings.Contains(output, "openclaw") || !strings.Contains(output, "missing") {
		t.Fatalf("stdout = %q, want missing openclaw", output)
	}
}

func TestRunEnableUsesGlobalStoreAndConfiguredTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "target-skills")
	cli := app{
		loadTargets: func(string) ([]targets.Target, error) {
			return []targets.Target{
				{ID: "codex-local", Type: targets.TypeCustom, Format: "markdown-skill-dir", InstallRoot: targetDir, Mode: "symlink", Enabled: true},
			}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"enable", ref, "--target", "codex-local"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "enabled: raul/code-reviewer@0.4.0 -> codex-local") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunActivateUsesGlobalStoreAndConfiguredTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	targetDir := filepath.Join(t.TempDir(), "target-skills")
	cli := app{
		loadTargets: func(string) ([]targets.Target, error) {
			return []targets.Target{
				{ID: "codex-local", Type: targets.TypeCustom, Format: "markdown-skill-dir", InstallRoot: targetDir, Mode: "symlink", Enabled: true},
			}, nil
		},
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := cli.Run(context.Background(), []string{"activate", ref, "--target", "codex-local"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "activated: raul/code-reviewer@0.4.0 -> codex-local") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunEnableUsesBuiltInCodexWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "codex", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".agents", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}

	configPath := filepath.Join(home, ".agentlib", "config.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	var config struct {
		Version     int `json:"version"`
		Activations []struct {
			TargetID string `json:"targetId"`
			Ref      string `json:"ref"`
			Path     string `json:"path"`
		} `json:"activations"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if config.Version != 1 {
		t.Fatalf("config.Version = %d, want 1", config.Version)
	}
	if len(config.Activations) != 1 {
		t.Fatalf("len(config.Activations) = %d, want 1", len(config.Activations))
	}
	if config.Activations[0].TargetID != "codex" || config.Activations[0].Ref != ref || config.Activations[0].Path != targetPath {
		t.Fatalf("activation = %+v", config.Activations[0])
	}
}

func TestRunDeactivateRemovesBuiltInCodexActivationAndState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	enableStdout := strings.Builder{}
	enableStderr := strings.Builder{}
	if exitCode := (app{}).Run(context.Background(), []string{"enable", "--target", "codex", ref}, &enableStdout, &enableStderr); exitCode != 0 {
		t.Fatalf("enable exitCode = %d, stderr = %q", exitCode, enableStderr.String())
	}

	targetPath := filepath.Join(home, ".agents", "skills", "raul", "code-reviewer", "0.4.0")
	if _, err := os.Lstat(targetPath); err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"deactivate", "--target", "codex", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "deactivated: raul/code-reviewer@0.4.0 -> codex") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("target path still exists: %v", err)
	}

	configPath := filepath.Join(home, ".agentlib", "config.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	var config struct {
		Version     int `json:"version"`
		Activations []struct {
			TargetID string `json:"targetId"`
			Ref      string `json:"ref"`
			Path     string `json:"path"`
		} `json:"activations"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if len(config.Activations) != 0 {
		t.Fatalf("len(config.Activations) = %d, want 0", len(config.Activations))
	}
}

func TestRunEnableUsesBuiltInOpenClawWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "agent.yaml"), []byte("name: code-reviewer\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "openclaw", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".openclaw", "agents", "raul", "code-reviewer", "0.4.0")
	if got := stdout.String(); !strings.Contains(got, "enabled: raul/code-reviewer@0.4.0 -> openclaw") {
		t.Fatalf("stdout = %q", got)
	}
	content, err := os.ReadFile(filepath.Join(targetPath, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "hello\n" {
		t.Fatalf("README.md = %q", string(content))
	}

	metaPath := filepath.Join(targetPath, "agentlib-export.json")
	metaContent, err := os.ReadFile(metaPath)
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
	if meta.TargetID != "openclaw" || meta.SourceRef != ref || meta.SourceStorePath != storePath {
		t.Fatalf("metadata = %+v", meta)
	}
	if meta.ExportedAt == "" {
		t.Fatal("ExportedAt is empty")
	}
	if meta.FormatVersion != 1 {
		t.Fatalf("FormatVersion = %d, want 1", meta.FormatVersion)
	}
}

func TestRunEnableUsesBuiltInCrewAIWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "crewai", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".crewai", "agents", "raul", "code-reviewer", "0.4.0")
	starterContent, err := os.ReadFile(filepath.Join(targetPath, "crewai-agent.py"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if got := string(starterContent); !strings.Contains(got, "raul/code-reviewer@0.4.0") {
		t.Fatalf("starter = %q", got)
	}
}

func TestRunEnableUsesBuiltInLangChainWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "langchain", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".langchain", "agents", "raul", "code-reviewer", "0.4.0")
	starterContent, err := os.ReadFile(filepath.Join(targetPath, "langchain-agent.py"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if got := string(starterContent); !strings.Contains(got, "raul/code-reviewer@0.4.0") {
		t.Fatalf("starter = %q", got)
	}
}

func TestRunEnableResolvesClaudeAliasToBuiltInTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "claude", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".claude", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
	if !strings.Contains(stdout.String(), "-> claude-code") {
		t.Fatalf("stdout = %q, want canonical target id", stdout.String())
	}
}

func TestFindTargetResolvesGeminiAlias(t *testing.T) {
	target, ok := findTarget([]targets.Target{
		{ID: "gemini-cli", Type: targets.TypeBuiltIn},
	}, "gemini")
	if !ok {
		t.Fatal("findTarget returned ok = false, want true")
	}
	if target.ID != "gemini-cli" {
		t.Fatalf("target.ID = %q", target.ID)
	}
}

func TestRunEnableResolvesGeminiAliasToBuiltInTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "gemini", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".gemini", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
	if !strings.Contains(stdout.String(), "-> gemini-cli") {
		t.Fatalf("stdout = %q, want canonical target id", stdout.String())
	}
}

func TestRunEnableUsesBuiltInOpenCodeWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "opencode", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".config", "opencode", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func TestRunEnableUsesBuiltInCursorWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "cursor", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".cursor", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func TestRunEnableUsesBuiltInAntigravityWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "antigravity", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".gemini", "antigravity", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func TestRunEnableUsesBuiltInVSCodeWithoutCustomConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ref := "raul/code-reviewer@0.4.0"
	storePath := filepath.Join(home, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storePath, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := app{}.Run(context.Background(), []string{"enable", "--target", "vscode", ref}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}

	targetPath := filepath.Join(home, ".vscode", "agentlib", "skills", "raul", "code-reviewer", "0.4.0")
	info, err := os.Lstat(targetPath)
	if err != nil {
		t.Fatalf("Lstat returned error: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func (client fakeRegistryClient) FetchVersion(context.Context, agentref.Ref) (registry.Version, error) {
	if client.version.Version != "" {
		return client.version, nil
	}
	return registry.Version{
		Namespace: "raul",
		Name:      "code-reviewer",
		Version:   "0.4.0",
	}, nil
}

func (client fakeRegistryClient) FetchArtifacts(context.Context, agentref.Ref) ([]registry.Artifact, error) {
	return []registry.Artifact{
		{Path: "agent.yaml", MediaType: "application/yaml", SizeBytes: 12},
		{Path: "README.md", MediaType: "text/markdown", SizeBytes: 24},
	}, nil
}

func (client fakeRegistryClient) DownloadArtifact(context.Context, agentref.Ref, string) ([]byte, string, error) {
	return []byte("content\n"), "text/plain", nil
}

func (client fakeRegistryClient) FetchAgents(context.Context) ([]registry.AgentSummary, error) {
	return client.agents, nil
}

func TestProjectMarkerPath(t *testing.T) {
	root := t.TempDir()
	got := install.ProjectMarkerPath(root)
	want := filepath.Join(root, ".agentlib", "project.json")
	if got != want {
		t.Fatalf("ProjectMarkerPath = %q, want %q", got, want)
	}
}
