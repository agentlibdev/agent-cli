package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/install"
	"github.com/agentlibdev/agent-cli/internal/registry"
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

type fakeRegistryClient struct {
	agents []registry.AgentSummary
}

func (client fakeRegistryClient) FetchVersion(context.Context, agentref.Ref) (registry.Version, error) {
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
