package targets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

const activationConfigVersion = 1

type Activation struct {
	TargetID    string `json:"targetId"`
	Ref         string `json:"ref"`
	Path        string `json:"path"`
	ActivatedAt string `json:"activatedAt"`
}

type Config struct {
	Version     int          `json:"version"`
	Activations []Activation `json:"activations"`
}

func ConfigPath(storeRoot string) string {
	return filepath.Join(storeRoot, "config.json")
}

func LoadConfig(storeRoot string) (Config, error) {
	content, err := os.ReadFile(ConfigPath(storeRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{Version: activationConfigVersion, Activations: []Activation{}}, nil
		}
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return Config{}, err
	}
	if config.Version == 0 {
		config.Version = activationConfigVersion
	}
	if config.Activations == nil {
		config.Activations = []Activation{}
	}

	return config, nil
}

func UpsertActivation(storeRoot string, targetID string, ref agentref.Ref, path string) error {
	config, err := LoadConfig(storeRoot)
	if err != nil {
		return err
	}

	record := Activation{
		TargetID:    targetID,
		Ref:         ref.String(),
		Path:        path,
		ActivatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}

	replaced := false
	for index := range config.Activations {
		if config.Activations[index].TargetID == targetID && config.Activations[index].Ref == ref.String() {
			config.Activations[index] = record
			replaced = true
			break
		}
	}
	if !replaced {
		config.Activations = append(config.Activations, record)
	}

	return writeConfig(storeRoot, config)
}

func RemoveActivation(storeRoot string, targetID string, ref agentref.Ref) error {
	config, err := LoadConfig(storeRoot)
	if err != nil {
		return err
	}

	filtered := make([]Activation, 0, len(config.Activations))
	for _, item := range config.Activations {
		if item.TargetID == targetID && item.Ref == ref.String() {
			continue
		}
		filtered = append(filtered, item)
	}
	config.Activations = filtered

	return writeConfig(storeRoot, config)
}

func ActivationsForRef(storeRoot string, ref agentref.Ref) ([]Activation, error) {
	config, err := LoadConfig(storeRoot)
	if err != nil {
		return nil, err
	}

	filtered := make([]Activation, 0, len(config.Activations))
	for _, item := range config.Activations {
		if item.Ref == ref.String() {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func writeConfig(storeRoot string, config Config) error {
	config.Version = activationConfigVersion
	if config.Activations == nil {
		config.Activations = []Activation{}
	}

	if err := os.MkdirAll(storeRoot, 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(storeRoot), append(content, '\n'), 0o644)
}
