package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Mode string

const (
	ModeGlobal Mode = "global"
	ModeLocal  Mode = "local"
)

type Target struct {
	Mode         Mode
	Root         string
	LockfilePath string
}

type TargetOptions struct {
	WorkingDir string
	Local      bool
	Global     bool
	InstallDir string
}

func ResolveTarget(options TargetOptions) (Target, error) {
	if options.Local && options.Global {
		return Target{}, errors.New("cannot use --local and --global together")
	}

	if options.InstallDir != "" && !options.Local {
		return Target{}, errors.New("--install-dir requires --local")
	}

	if options.Local {
		root := ""
		if options.InstallDir != "" {
			if filepath.IsAbs(options.InstallDir) {
				root = options.InstallDir
			} else {
				root = filepath.Join(options.WorkingDir, options.InstallDir)
			}
		} else {
			projectRoot, err := FindProjectRoot(options.WorkingDir)
			if err != nil {
				return Target{}, err
			}
			root = filepath.Join(projectRoot, ".agentlib")
		}

		return Target{
			Mode:         ModeLocal,
			Root:         root,
			LockfilePath: filepath.Join(root, "agent.lock.json"),
		}, nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		return Target{}, errors.New("HOME is not set")
	}

	root := filepath.Join(home, ".agentlib")
	return Target{
		Mode:         ModeGlobal,
		Root:         root,
		LockfilePath: filepath.Join(root, "agent.lock.json"),
	}, nil
}

func ProjectMarkerPath(root string) string {
	return filepath.Join(root, ".agentlib", "project.json")
}

func FindProjectRoot(workingDir string) (string, error) {
	current := workingDir
	for {
		if _, err := os.Stat(ProjectMarkerPath(current)); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", errors.New("no AgentLib project found; run agentlib init or use --global")
}

func InitProject(workingDir string) (string, error) {
	markerPath := ProjectMarkerPath(workingDir)
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o755); err != nil {
		return "", err
	}

	body, err := json.MarshalIndent(struct {
		Version int `json:"version"`
	}{Version: 1}, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(markerPath, append(body, '\n'), 0o644); err != nil {
		return "", err
	}

	return markerPath, nil
}
