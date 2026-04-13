package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/targets"
)

type installActivationOptions struct {
	RuntimeIDs  []string
	AllDetected bool
	NoActivate  bool
}

type stringListFlag []string

func (items *stringListFlag) String() string {
	return strings.Join(*items, ",")
}

func (items *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("runtime id cannot be empty")
	}

	*items = append(*items, value)
	return nil
}

func (a app) maybeActivateInstall(
	storeRoot string,
	workingDir string,
	ref agentref.Ref,
	options installActivationOptions,
	stdin io.Reader,
	stdout io.Writer,
) error {
	if options.NoActivate {
		return nil
	}

	selectedTargets, err := a.selectInstallTargets(workingDir, options, stdin, stdout)
	if err != nil {
		return err
	}
	if len(selectedTargets) == 0 {
		return nil
	}

	enableTarget := a.enableTarget
	if enableTarget == nil {
		enableTarget = targets.Enable
	}

	for _, target := range selectedTargets {
		result, err := enableTarget(storeRoot, target, ref)
		if err != nil {
			return fmt.Errorf("activate %s: %w", target.ID, err)
		}

		fmt.Fprintf(stdout, "activated: %s\n", target.ID)
		fmt.Fprintf(stdout, "path: %s\n", result.Path)
	}

	return nil
}

func (a app) selectInstallTargets(
	workingDir string,
	options installActivationOptions,
	stdin io.Reader,
	stdout io.Writer,
) ([]targets.Target, error) {
	if len(options.RuntimeIDs) > 0 {
		return a.resolveTargetsByID(workingDir, options.RuntimeIDs)
	}

	if options.AllDetected {
		return a.detectedInstallTargets(workingDir)
	}

	isInteractiveInput := a.isInteractiveInput
	if isInteractiveInput == nil {
		isInteractiveInput = defaultInteractiveInput
	}
	if !isInteractiveInput() {
		return nil, nil
	}

	items, err := a.detectedInstallTargets(workingDir)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	return promptInstallTargets(stdin, stdout, items)
}

func (a app) resolveTargetsByID(workingDir string, ids []string) ([]targets.Target, error) {
	loadTargets := a.loadTargets
	if loadTargets == nil {
		loadTargets = targets.Load
	}

	items, err := loadTargets(workingDir)
	if err != nil {
		return nil, err
	}

	selected := make([]targets.Target, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		target, ok := findTarget(items, id)
		if !ok {
			return nil, fmt.Errorf("target %q not found", id)
		}
		if !target.Enabled {
			return nil, fmt.Errorf("target %q is disabled", target.ID)
		}
		if _, ok := seen[target.ID]; ok {
			continue
		}
		seen[target.ID] = struct{}{}
		selected = append(selected, target)
	}

	return selected, nil
}

func (a app) detectedInstallTargets(workingDir string) ([]targets.Target, error) {
	detectTargets := a.detectTargets
	if detectTargets == nil {
		detectTargets = targets.Detect
	}

	detections, err := detectTargets(workingDir)
	if err != nil {
		return nil, err
	}

	selected := make([]targets.Target, 0, len(detections))
	seen := map[string]struct{}{}
	for _, detection := range detections {
		if !detection.Detected || !detection.Target.Enabled {
			continue
		}
		if _, ok := seen[detection.Target.ID]; ok {
			continue
		}
		seen[detection.Target.ID] = struct{}{}
		selected = append(selected, detection.Target)
	}

	return selected, nil
}

func promptInstallTargets(stdin io.Reader, stdout io.Writer, items []targets.Target) ([]targets.Target, error) {
	fmt.Fprintln(stdout, "Select runtimes to activate:")
	for index, item := range items {
		name := item.Name
		if name == "" {
			name = item.ID
		}
		fmt.Fprintf(stdout, "%d. %s (%s)\n", index+1, name, item.ID)
	}
	fmt.Fprint(stdout, "Enter comma-separated numbers, or press Enter to skip: ")

	reader := bufio.NewReader(stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})

	selected := make([]targets.Target, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		index, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", part)
		}
		if index < 1 || index > len(items) {
			return nil, fmt.Errorf("selection %d out of range", index)
		}
		if _, ok := seen[index]; ok {
			continue
		}
		seen[index] = struct{}{}
		selected = append(selected, items[index-1])
	}

	return selected, nil
}
