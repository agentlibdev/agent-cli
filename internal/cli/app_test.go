package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/registry"
)

func TestRunSearchFiltersAgents(t *testing.T) {
	previousFactory := newRegistryClient
	newRegistryClient = func(string) registryClient {
		return fakeRegistryClient{
			agents: []registry.AgentSummary{
				{Namespace: "raul", Name: "code-reviewer", LatestVersion: "0.4.0", Title: "Code Reviewer", Description: "Reviews code changes."},
				{Namespace: "raul", Name: "docs-writer", LatestVersion: "0.2.0", Title: "Docs Writer", Description: "Drafts documentation."},
			},
		}
	}
	t.Cleanup(func() {
		newRegistryClient = previousFactory
	})

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := Run(context.Background(), []string{"search", "docs"}, &stdout, &stderr)
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

func TestRunRemoveDeletesInstalledVersion(t *testing.T) {
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

	installedPath := filepath.Join(root, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(installedPath, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	lockfilePath := filepath.Join(root, ".agentlib", "agent.lock.json")
	if err := os.WriteFile(lockfilePath, []byte("{\"version\":1,\"agent\":{\"namespace\":\"raul\",\"name\":\"code-reviewer\",\"version\":\"0.4.0\"}}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := Run(context.Background(), []string{"remove", "raul/code-reviewer@0.4.0"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run exitCode = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed: raul/code-reviewer@0.4.0") {
		t.Fatalf("stdout = %q, want remove confirmation", stdout.String())
	}
	if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
		t.Fatalf("installed path still exists: %v", err)
	}
}

type fakeRegistryClient struct {
	agents []registry.AgentSummary
}

func (client fakeRegistryClient) FetchVersion(context.Context, agentref.Ref) (registry.Version, error) {
	return registry.Version{}, nil
}

func (client fakeRegistryClient) FetchArtifacts(context.Context, agentref.Ref) ([]registry.Artifact, error) {
	return nil, nil
}

func (client fakeRegistryClient) DownloadArtifact(context.Context, agentref.Ref, string) ([]byte, string, error) {
	return nil, "", nil
}

func (client fakeRegistryClient) FetchAgents(context.Context) ([]registry.AgentSummary, error) {
	return client.agents, nil
}
