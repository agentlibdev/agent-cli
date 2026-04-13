package targets

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

const packageExportFormatVersion = 1

type EnableResult struct {
	Path string
}

type DisableResult struct {
	Path string
}

func Enable(storeRoot string, target Target, ref agentref.Ref) (EnableResult, error) {
	if target.InstallRoot == "" {
		return EnableResult{}, fmt.Errorf("target %s has no installRoot configured", target.ID)
	}

	sourceRoot := filepath.Join(storeRoot, "agents", ref.Namespace, ref.Name, ref.Version)
	if _, err := os.Stat(sourceRoot); err != nil {
		if os.IsNotExist(err) {
			return EnableResult{}, fmt.Errorf("agent %s/%s@%s is not installed in %s", ref.Namespace, ref.Name, ref.Version, storeRoot)
		}
		return EnableResult{}, err
	}

	targetPath := filepath.Join(target.InstallRoot, ref.Namespace, ref.Name, ref.Version)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return EnableResult{}, err
	}
	if err := os.RemoveAll(targetPath); err != nil {
		return EnableResult{}, err
	}

	switch target.Mode {
	case "copy":
		if err := copyDir(sourceRoot, targetPath); err != nil {
			return EnableResult{}, err
		}
	case "symlink", "":
		if err := os.Symlink(sourceRoot, targetPath); err != nil {
			return EnableResult{}, err
		}
	case "generate":
		if target.Format != "package-export" {
			return EnableResult{}, fmt.Errorf("target %s uses unsupported format %q for mode %q", target.ID, target.Format, target.Mode)
		}
		if err := copyDir(sourceRoot, targetPath); err != nil {
			return EnableResult{}, err
		}
		if err := writePackageExportMetadata(targetPath, target.ID, ref, sourceRoot); err != nil {
			return EnableResult{}, err
		}
		if err := writeGeneratedStarter(targetPath, target.ID, ref); err != nil {
			return EnableResult{}, err
		}
	default:
		return EnableResult{}, fmt.Errorf("target %s uses unsupported mode %q", target.ID, target.Mode)
	}

	return EnableResult{Path: targetPath}, nil
}

func Disable(target Target, ref agentref.Ref) (DisableResult, error) {
	if target.InstallRoot == "" {
		return DisableResult{}, fmt.Errorf("target %s has no installRoot configured", target.ID)
	}

	targetPath := filepath.Join(target.InstallRoot, ref.Namespace, ref.Name, ref.Version)
	if err := os.RemoveAll(targetPath); err != nil {
		return DisableResult{}, err
	}

	return DisableResult{Path: targetPath}, nil
}

func copyDir(sourceRoot, targetRoot string) error {
	return filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(targetRoot, 0o755)
		}

		targetPath := filepath.Join(targetRoot, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, sourceFile); err != nil {
			return err
		}

		return nil
	})
}

type packageExportMetadata struct {
	TargetID        string `json:"targetId"`
	SourceRef       string `json:"sourceRef"`
	SourceStorePath string `json:"sourceStorePath"`
	ExportedAt      string `json:"exportedAt"`
	FormatVersion   int    `json:"formatVersion"`
}

func writePackageExportMetadata(targetPath, targetID string, ref agentref.Ref, sourceStorePath string) error {
	content, err := json.MarshalIndent(packageExportMetadata{
		TargetID:        targetID,
		SourceRef:       fmt.Sprintf("%s/%s@%s", ref.Namespace, ref.Name, ref.Version),
		SourceStorePath: sourceStorePath,
		ExportedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		FormatVersion:   packageExportFormatVersion,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetPath, "agentlib-export.json"), append(content, '\n'), 0o644)
}

func writeGeneratedStarter(targetPath, targetID string, ref agentref.Ref) error {
	sourceRef := fmt.Sprintf("%s/%s@%s", ref.Namespace, ref.Name, ref.Version)

	switch targetID {
	case "crewai":
		content := fmt.Sprintf(
			"# AgentLib CrewAI export for %s\n# Generated from %s\n\nPACKAGE_DIR = %q\n",
			sourceRef,
			sourceRef,
			targetPath,
		)
		return os.WriteFile(filepath.Join(targetPath, "crewai-agent.py"), []byte(content), 0o644)
	case "langchain":
		content := fmt.Sprintf(
			"# AgentLib LangChain export for %s\n# Generated from %s\n\nPACKAGE_DIR = %q\n",
			sourceRef,
			sourceRef,
			targetPath,
		)
		return os.WriteFile(filepath.Join(targetPath, "langchain-agent.py"), []byte(content), 0o644)
	default:
		return nil
	}
}
