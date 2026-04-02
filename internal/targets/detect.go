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
			path, ok := detectBuiltIn(item.ID, lookPath)
			if ok {
				detection.Detected = true
				detection.Status = "detected"
				detection.Path = path
				detection.Evidence = "command"
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

func detectBuiltIn(id string, lookPath func(string) (string, error)) (string, bool) {
	for _, candidate := range builtInCommands(id) {
		path, err := lookPath(candidate)
		if err == nil {
			return path, true
		}
		if err != nil && !errors.Is(err, exec.ErrNotFound) {
			continue
		}
	}

	return "", false
}

func builtInCommands(id string) []string {
	switch id {
	case "codex":
		return []string{"codex"}
	case "claude":
		return []string{"claude"}
	case "gemini-cli":
		return []string{"gemini", "gemini-cli"}
	case "openclaw":
		return []string{"openclaw"}
	default:
		return []string{id}
	}
}
