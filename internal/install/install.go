package install

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/registry"
)

type Registry interface {
	FetchVersion(ctx context.Context, ref agentref.Ref) (registry.Version, error)
	FetchArtifacts(ctx context.Context, ref agentref.Ref) ([]registry.Artifact, error)
	DownloadArtifact(ctx context.Context, ref agentref.Ref, path string) ([]byte, string, error)
}

type Result struct {
	Root string
}

type Lockfile struct {
	Version int `json:"version"`
	Agent   struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Version   string `json:"version"`
	} `json:"agent"`
}

func Run(ctx context.Context, registryClient Registry, workingDir string, ref agentref.Ref) (Result, error) {
	_, err := registryClient.FetchVersion(ctx, ref)
	if err != nil {
		return Result{}, err
	}

	artifacts, err := registryClient.FetchArtifacts(ctx, ref)
	if err != nil {
		return Result{}, err
	}

	installRoot := filepath.Join(workingDir, ".agentlib", "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(installRoot, 0o755); err != nil {
		return Result{}, err
	}

	for _, artifact := range artifacts {
		content, _, err := registryClient.DownloadArtifact(ctx, ref, artifact.Path)
		if err != nil {
			return Result{}, err
		}

		targetPath := filepath.Join(installRoot, artifact.Path)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return Result{}, err
		}
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return Result{}, err
		}
	}

	lockfile := Lockfile{Version: 1}
	lockfile.Agent.Namespace = ref.Namespace
	lockfile.Agent.Name = ref.Name
	lockfile.Agent.Version = ref.Version

	lockfilePath := filepath.Join(workingDir, ".agentlib", "agent.lock.json")
	lockfileBytes, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(lockfilePath, append(lockfileBytes, '\n'), 0o644); err != nil {
		return Result{}, err
	}

	return Result{Root: installRoot}, nil
}

func Remove(workingDir string, ref agentref.Ref) error {
	installRoot := filepath.Join(workingDir, ".agentlib", "agents", ref.Namespace, ref.Name, ref.Version)
	if err := os.RemoveAll(installRoot); err != nil {
		return err
	}

	lockfilePath := filepath.Join(workingDir, ".agentlib", "agent.lock.json")
	lockfileBytes, err := os.ReadFile(lockfilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var lockfile Lockfile
	if err := json.Unmarshal(lockfileBytes, &lockfile); err != nil {
		return err
	}

	if lockfile.Agent.Namespace == ref.Namespace &&
		lockfile.Agent.Name == ref.Name &&
		lockfile.Agent.Version == ref.Version {
		if err := os.Remove(lockfilePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
