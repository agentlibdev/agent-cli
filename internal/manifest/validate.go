package manifest

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Namespace   string `yaml:"namespace"`
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	} `yaml:"metadata"`
}

func ValidateYAML(source []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(source, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest yaml: %w", err)
	}

	required := map[string]string{
		"apiVersion":           manifest.APIVersion,
		"kind":                 manifest.Kind,
		"metadata.namespace":   manifest.Metadata.Namespace,
		"metadata.name":        manifest.Metadata.Name,
		"metadata.version":     manifest.Metadata.Version,
		"metadata.title":       manifest.Metadata.Title,
		"metadata.description": manifest.Metadata.Description,
	}

	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return Manifest{}, fmt.Errorf("missing required field %s", field)
		}
	}

	return manifest, nil
}
