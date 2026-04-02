package targets

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

type EnableResult struct {
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
	default:
		return EnableResult{}, fmt.Errorf("target %s uses unsupported mode %q", target.ID, target.Mode)
	}

	return EnableResult{Path: targetPath}, nil
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
