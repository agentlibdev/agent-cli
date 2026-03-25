package install

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/registry"
)

type stubRegistry struct{}

func (stubRegistry) FetchVersion(_ context.Context, ref agentref.Ref) (registry.Version, error) {
	return registry.Version{
		Namespace:   ref.Namespace,
		Name:        ref.Name,
		Version:     ref.Version,
		Title:       "Code Reviewer",
		Description: "Reviews code changes.",
		License:     "MIT",
	}, nil
}

func (stubRegistry) FetchArtifacts(_ context.Context, _ agentref.Ref) ([]registry.Artifact, error) {
	return []registry.Artifact{
		{Path: "agent.yaml", MediaType: "application/yaml", SizeBytes: 12},
		{Path: "README.md", MediaType: "text/markdown", SizeBytes: 24},
	}, nil
}

func (stubRegistry) DownloadArtifact(_ context.Context, _ agentref.Ref, path string) ([]byte, string, error) {
	switch path {
	case "agent.yaml":
		return []byte("kind: Agent\n"), "application/yaml", nil
	case "README.md":
		return []byte("# Code Reviewer\n"), "text/markdown", nil
	default:
		return nil, "", os.ErrNotExist
	}
}

func TestInstallerWritesArtifactsAndLockfile(t *testing.T) {
	root := t.TempDir()
	ref, err := agentref.Parse("raul/code-reviewer@0.4.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	result, err := Run(t.Context(), stubRegistry{}, root, ref)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	agentPath := filepath.Join(root, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0", "README.md")
	if _, err := os.Stat(agentPath); err != nil {
		t.Fatalf("installed README missing: %v", err)
	}

	lockfilePath := filepath.Join(root, ".agentlib", "agent.lock.json")
	if _, err := os.Stat(lockfilePath); err != nil {
		t.Fatalf("lockfile missing: %v", err)
	}

	if result.Root != filepath.Join(root, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0") {
		t.Fatalf("Root = %q", result.Root)
	}
}

func TestRemoveDeletesVersionAndMatchingLockfile(t *testing.T) {
	root := t.TempDir()

	versionDir := filepath.Join(root, ".agentlib", "agents", "raul", "code-reviewer", "0.4.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	otherVersionDir := filepath.Join(root, ".agentlib", "agents", "raul", "code-reviewer", "0.3.0")
	if err := os.MkdirAll(otherVersionDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	lockfilePath := filepath.Join(root, ".agentlib", "agent.lock.json")
	lockfile := Lockfile{Version: 1}
	lockfile.Agent.Namespace = "raul"
	lockfile.Agent.Name = "code-reviewer"
	lockfile.Agent.Version = "0.4.0"
	lockfileBytes, err := json.Marshal(lockfile)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if err := os.WriteFile(lockfilePath, lockfileBytes, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	ref, err := agentref.Parse("raul/code-reviewer@0.4.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if err := Remove(root, ref); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Fatalf("removed version dir still exists: %v", err)
	}
	if _, err := os.Stat(otherVersionDir); err != nil {
		t.Fatalf("other version dir missing: %v", err)
	}
	if _, err := os.Stat(lockfilePath); !os.IsNotExist(err) {
		t.Fatalf("matching lockfile still exists: %v", err)
	}
}
