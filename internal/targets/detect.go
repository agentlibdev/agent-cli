package targets

import (
	"errors"
	"os"
	"os/exec"
)

type Detection struct {
	Target   Target
	Detected bool
	Status   string
	Path     string
	Evidence string
}

func Detect(projectDir string) ([]Detection, error) {
	items, err := Load(projectDir)
	if err != nil {
		return nil, err
	}

	return DetectWithLookups(projectDir, items, exec.LookPath, os.Stat)
}

func DetectWithLookups(
	projectDir string,
	items []Target,
	lookPath func(string) (string, error),
	stat func(string) (os.FileInfo, error),
) ([]Detection, error) {
	detections := make([]Detection, 0, len(items))

	for _, item := range items {
		detection := Detection{
			Target: item,
			Status: "missing",
		}

		if item.Type == TypeBuiltIn {
			path, evidence, ok := detectBuiltIn(item, lookPath, stat)
			if ok {
				detection.Detected = true
				detection.Status = "detected"
				detection.Path = path
				detection.Evidence = evidence
			}
			detections = append(detections, detection)
			continue
		}

		if item.InstallRoot != "" {
			if _, err := stat(item.InstallRoot); err == nil {
				detection.Detected = true
				detection.Status = "detected"
				detection.Path = item.InstallRoot
				detection.Evidence = "installRoot"
				detections = append(detections, detection)
				continue
			}
		}

		if item.ManifestPath != "" {
			if _, err := stat(item.ManifestPath); err == nil {
				detection.Detected = true
				detection.Status = "detected"
				detection.Path = item.ManifestPath
				detection.Evidence = "manifestPath"
			}
		}

		detections = append(detections, detection)
	}

	return detections, nil
}

func detectBuiltIn(target Target, lookPath func(string) (string, error), stat func(string) (os.FileInfo, error)) (string, string, bool) {
	id := target.ID
	for _, candidate := range builtInCommands(id) {
		path, err := lookPath(candidate)
		if err == nil {
			return path, "command", true
		}
		if err != nil && !errors.Is(err, exec.ErrNotFound) {
			continue
		}
	}

	if target.InstallRoot != "" {
		if _, err := stat(target.InstallRoot); err == nil {
			return target.InstallRoot, "installRoot", true
		}
	}

	if target.ManifestPath != "" {
		if _, err := stat(target.ManifestPath); err == nil {
			return target.ManifestPath, "manifestPath", true
		}
	}

	return "", "", false
}

func builtInCommands(id string) []string {
	switch id {
	case "antigravity":
		return []string{"antigravity"}
	case "claude-code":
		return []string{"claude"}
	case "cursor":
		return []string{"cursor"}
	case "codex":
		return []string{"codex"}
	case "gemini-cli":
		return []string{"gemini", "gemini-cli"}
	case "github-copilot":
		return []string{"github-copilot", "copilot"}
	case "opencode":
		return []string{"opencode"}
	case "windsurf":
		return []string{"windsurf"}
	default:
		return []string{id}
	}
}
