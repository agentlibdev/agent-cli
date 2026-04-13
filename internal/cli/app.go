package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/install"
	"github.com/agentlibdev/agent-cli/internal/manifest"
	"github.com/agentlibdev/agent-cli/internal/registry"
	"github.com/agentlibdev/agent-cli/internal/targets"
	"github.com/agentlibdev/agent-cli/internal/version"
)

type registryClient interface {
	FetchVersion(ctx context.Context, ref agentref.Ref) (registry.Version, error)
	FetchArtifacts(ctx context.Context, ref agentref.Ref) ([]registry.Artifact, error)
	DownloadArtifact(ctx context.Context, ref agentref.Ref, path string) ([]byte, string, error)
	FetchAgents(ctx context.Context) ([]registry.AgentSummary, error)
}

type app struct {
	newRegistryClient  func(baseURL string) registryClient
	loadTargets        func(projectDir string) ([]targets.Target, error)
	detectTargets      func(projectDir string) ([]targets.Detection, error)
	enableTarget       func(storeRoot string, target targets.Target, ref agentref.Ref) (targets.EnableResult, error)
	disableTarget      func(target targets.Target, ref agentref.Ref) (targets.DisableResult, error)
	stdin              io.Reader
	isInteractiveInput func() bool
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return app{
		newRegistryClient: func(baseURL string) registryClient {
			return registry.New(baseURL)
		},
		loadTargets:        targets.Load,
		detectTargets:      targets.Detect,
		enableTarget:       targets.Enable,
		disableTarget:      targets.Disable,
		stdin:              os.Stdin,
		isInteractiveInput: defaultInteractiveInput,
	}.Run(ctx, args, stdout, stderr)
}

func (a app) Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 1
	}

	switch args[0] {
	case "version":
		return a.runVersion(args[1:], stdout, stderr)
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "targets":
		return a.runTargets(args[1:], stdout, stderr)
	case "enable":
		return a.runEnable(args[1:], stdout, stderr)
	case "deactivate":
		return a.runDeactivate(args[1:], stdout, stderr)
	case "search":
		return a.runSearch(ctx, args[1:], stdout, stderr)
	case "show":
		return a.runShow(ctx, args[1:], stdout, stderr)
	case "install":
		return a.runInstall(ctx, args[1:], stdout, stderr)
	case "remove":
		return runRemove(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func (a app) runVersion(args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: agentlib version")
		return 1
	}

	fmt.Fprintf(stdout, "agentlib %s\n", version.Version)
	return 0
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib validate <path>")
		return 1
	}

	content, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "read manifest: %v\n", err)
		return 1
	}

	validated, err := manifest.ValidateYAML(content)
	if err != nil {
		fmt.Fprintf(stderr, "manifest invalid: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "manifest valid: %s/%s@%s\n", validated.Metadata.Namespace, validated.Metadata.Name, validated.Metadata.Version)
	return 0
}

func (a app) runSearch(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib search <query>")
		return 1
	}

	client := a.registryClient()
	agents, err := client.FetchAgents(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "search agents: %v\n", err)
		return 1
	}

	query := strings.ToLower(args[0])
	results := 0
	for _, agent := range agents {
		if !matchesSearch(agent, query) {
			continue
		}

		fmt.Fprintf(stdout, "%s/%s@%s\n", agent.Namespace, agent.Name, agent.LatestVersion)
		fmt.Fprintf(stdout, "title: %s\n", agent.Title)
		fmt.Fprintf(stdout, "description: %s\n", agent.Description)
		results++
	}

	if results == 0 {
		fmt.Fprintln(stdout, "no agents found")
	}

	return 0
}

func (a app) runShow(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib show <namespace/name@version>")
		return 1
	}

	ref, err := agentref.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	client := a.registryClient()
	version, err := client.FetchVersion(ctx, ref)
	if err != nil {
		fmt.Fprintf(stderr, "fetch version: %v\n", err)
		return 1
	}

	artifacts, err := client.FetchArtifacts(ctx, ref)
	if err != nil {
		fmt.Fprintf(stderr, "fetch artifacts: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "%s/%s@%s\n", version.Namespace, version.Name, version.Version)
	fmt.Fprintf(stdout, "title: %s\n", version.Title)
	fmt.Fprintf(stdout, "description: %s\n", version.Description)
	fmt.Fprintf(stdout, "license: %s\n", version.License)
	fmt.Fprintf(stdout, "artifacts: %d\n", len(artifacts))
	for _, artifact := range artifacts {
		fmt.Fprintf(stdout, "- %s (%s, %d bytes)\n", artifact.Path, artifact.MediaType, artifact.SizeBytes)
	}

	return 0
}

func (a app) runInstall(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	local := flags.Bool("local", false, "")
	global := flags.Bool("global", false, "")
	flags.BoolVar(global, "g", false, "")
	installDir := flags.String("install-dir", "", "")
	noActivate := flags.Bool("no-activate", false, "")
	allDetected := flags.Bool("all-detected", false, "")
	var runtimes stringListFlag
	flags.Var(&runtimes, "runtime", "")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(stderr, "usage: agentlib install [--local|--global|-g] [--install-dir <dir>] [--runtime <id>] [--all-detected] [--no-activate] <namespace/name@version>")
		return 1
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: agentlib install [--local|--global|-g] [--install-dir <dir>] [--runtime <id>] [--all-detected] [--no-activate] <namespace/name@version>")
		return 1
	}

	resolvedTarget, err := resolveInstallTarget(*local, *global, *installDir)
	if err != nil {
		fmt.Fprintf(stderr, "resolve install target: %v\n", err)
		return 1
	}

	ref, err := agentref.Parse(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	result, err := install.Run(ctx, a.registryClient(), resolvedTarget.Root, ref)
	if err != nil {
		fmt.Fprintf(stderr, "install agent: %v\n", err)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	stdin := a.stdin
	if stdin == nil {
		stdin = os.Stdin
	}

	if err := a.maybeActivateInstall(
		resolvedTarget.Root,
		workingDir,
		ref,
		installActivationOptions{
			RuntimeIDs:  []string(runtimes),
			AllDetected: *allDetected,
			NoActivate:  *noActivate,
		},
		stdin,
		stdout,
	); err != nil {
		fmt.Fprintf(stderr, "activate runtimes: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "installed: %s/%s@%s\n", ref.Namespace, ref.Name, ref.Version)
	fmt.Fprintf(stdout, "root: %s\n", result.Root)
	fmt.Fprintf(stdout, "lockfile: %s\n", resolvedTarget.LockfilePath)
	return 0
}

func runRemove(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("remove", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	local := flags.Bool("local", false, "")
	global := flags.Bool("global", false, "")
	flags.BoolVar(global, "g", false, "")
	installDir := flags.String("install-dir", "", "")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(stderr, "usage: agentlib remove [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
		return 1
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: agentlib remove [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
		return 1
	}

	resolvedTarget, err := resolveInstallTarget(*local, *global, *installDir)
	if err != nil {
		fmt.Fprintf(stderr, "resolve install target: %v\n", err)
		return 1
	}

	ref, err := agentref.Parse(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	if err := install.Remove(resolvedTarget.Root, ref); err != nil {
		fmt.Fprintf(stderr, "remove agent: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "removed: %s/%s@%s\n", ref.Namespace, ref.Name, ref.Version)
	return 0
}

func (a app) runEnable(args []string, stdout, stderr io.Writer) int {
	local, global, installDir, targetID, refValue, err := parseEnableArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "usage: agentlib enable [--local|--global|-g] [--install-dir <dir>] --target <id> <namespace/name@version>")
		return 1
	}

	resolvedTarget, err := resolveInstallTarget(local, global, installDir)
	if err != nil {
		fmt.Fprintf(stderr, "resolve install target: %v\n", err)
		return 1
	}

	ref, err := agentref.Parse(refValue)
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	loadTargets := a.loadTargets
	if loadTargets == nil {
		loadTargets = targets.Load
	}

	items, err := loadTargets(workingDir)
	if err != nil {
		fmt.Fprintf(stderr, "load targets: %v\n", err)
		return 1
	}

	selected, ok := findTarget(items, targetID)
	if !ok {
		fmt.Fprintf(stderr, "target %q not found\n", targetID)
		return 1
	}

	enableTarget := a.enableTarget
	if enableTarget == nil {
		enableTarget = targets.Enable
	}

	result, err := enableTarget(resolvedTarget.Root, selected, ref)
	if err != nil {
		fmt.Fprintf(stderr, "enable target: %v\n", err)
		return 1
	}
	if err := targets.UpsertActivation(resolvedTarget.Root, selected.ID, ref, result.Path); err != nil {
		fmt.Fprintf(stderr, "persist activation: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "enabled: %s/%s@%s -> %s\n", ref.Namespace, ref.Name, ref.Version, selected.ID)
	fmt.Fprintf(stdout, "path: %s\n", result.Path)
	return 0
}

func (a app) runDeactivate(args []string, stdout, stderr io.Writer) int {
	local, global, installDir, targetID, refValue, err := parseEnableArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "usage: agentlib deactivate [--local|--global|-g] [--install-dir <dir>] --target <id> <namespace/name@version>")
		return 1
	}

	resolvedTarget, err := resolveInstallTarget(local, global, installDir)
	if err != nil {
		fmt.Fprintf(stderr, "resolve install target: %v\n", err)
		return 1
	}

	ref, err := agentref.Parse(refValue)
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	loadTargets := a.loadTargets
	if loadTargets == nil {
		loadTargets = targets.Load
	}

	items, err := loadTargets(workingDir)
	if err != nil {
		fmt.Fprintf(stderr, "load targets: %v\n", err)
		return 1
	}

	selected, ok := findTarget(items, targetID)
	if !ok {
		fmt.Fprintf(stderr, "target %q not found\n", targetID)
		return 1
	}

	disableTarget := a.disableTarget
	if disableTarget == nil {
		disableTarget = targets.Disable
	}

	result, err := disableTarget(selected, ref)
	if err != nil {
		fmt.Fprintf(stderr, "deactivate target: %v\n", err)
		return 1
	}
	if err := targets.RemoveActivation(resolvedTarget.Root, selected.ID, ref); err != nil {
		fmt.Fprintf(stderr, "persist activation removal: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "deactivated: %s/%s@%s -> %s\n", ref.Namespace, ref.Name, ref.Version, selected.ID)
	fmt.Fprintf(stdout, "path: %s\n", result.Path)
	return 0
}

func runInit(args []string, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: agentlib init")
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	projectFile, err := install.InitProject(workingDir)
	if err != nil {
		fmt.Fprintf(stderr, "init project: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "initialized: %s\n", projectFile)
	return 0
}

func (a app) runTargets(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib targets <list|detect>")
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	loadTargets := a.loadTargets
	if loadTargets == nil {
		loadTargets = targets.Load
	}

	switch args[0] {
	case "list":
		items, err := loadTargets(workingDir)
		if err != nil {
			fmt.Fprintf(stderr, "load targets: %v\n", err)
			return 1
		}

		for _, item := range items {
			state := "enabled"
			if !item.Enabled {
				state = "disabled"
			}
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Type, item.Format, item.Mode, state)
		}
	case "detect":
		detectTargets := a.detectTargets
		if detectTargets == nil {
			detectTargets = targets.Detect
		}

		items, err := detectTargets(workingDir)
		if err != nil {
			fmt.Fprintf(stderr, "detect targets: %v\n", err)
			return 1
		}

		for _, item := range items {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Target.ID, item.Target.Type, item.Target.Format, item.Status, item.Evidence, item.Path)
		}
	default:
		fmt.Fprintln(stderr, "usage: agentlib targets <list|detect>")
		return 1
	}

	return 0
}

func registryBaseURL() string {
	if value := os.Getenv("AGENTLIB_BASE_URL"); value != "" {
		return value
	}

	return "https://agentlib.dev"
}

func (a app) registryClient() registryClient {
	if a.newRegistryClient != nil {
		return a.newRegistryClient(registryBaseURL())
	}

	return registry.New(registryBaseURL())
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "agentlib <command>")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Commands:")
	fmt.Fprintln(writer, "  version")
	fmt.Fprintln(writer, "  validate <path>")
	fmt.Fprintln(writer, "  init")
	fmt.Fprintln(writer, "  search <query>")
	fmt.Fprintln(writer, "  show <namespace/name@version>")
	fmt.Fprintln(writer, "  targets list")
	fmt.Fprintln(writer, "  targets detect")
	fmt.Fprintln(writer, "  enable [--local|--global|-g] [--install-dir <dir>] --target <id> <namespace/name@version>")
	fmt.Fprintln(writer, "  deactivate [--local|--global|-g] [--install-dir <dir>] --target <id> <namespace/name@version>")
	fmt.Fprintln(writer, "  install [--local|--global|-g] [--install-dir <dir>] [--runtime <id>] [--all-detected] [--no-activate] <namespace/name@version>")
	fmt.Fprintln(writer, "  remove [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
}

func matchesSearch(agent registry.AgentSummary, query string) bool {
	candidate := strings.ToLower(strings.Join([]string{
		agent.Namespace,
		agent.Name,
		agent.Title,
		agent.Description,
	}, " "))

	return strings.Contains(candidate, query)
}

func resolveInstallTarget(local, global bool, installDir string) (install.Target, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return install.Target{}, err
	}

	return install.ResolveTarget(install.TargetOptions{
		WorkingDir: workingDir,
		Local:      local,
		Global:     global,
		InstallDir: installDir,
	})
}

func findTarget(items []targets.Target, id string) (targets.Target, bool) {
	for _, item := range items {
		if item.ID == id || targetAliasMatches(item.ID, id) {
			return item, true
		}
	}

	return targets.Target{}, false
}

func targetAliasMatches(canonicalID string, value string) bool {
	for _, alias := range targetAliases(canonicalID) {
		if alias == value {
			return true
		}
	}

	return false
}

func targetAliases(canonicalID string) []string {
	switch canonicalID {
	case "claude-code":
		return []string{"claude"}
	case "gemini-cli":
		return []string{"gemini"}
	case "github-copilot":
		return []string{"copilot", "github-copilot-chat"}
	default:
		return nil
	}
}

func parseEnableArgs(args []string) (bool, bool, string, string, string, error) {
	local := false
	global := false
	installDir := ""
	targetID := ""
	refValue := ""

	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--local":
			local = true
		case "--global", "-g":
			global = true
		case "--install-dir":
			index++
			if index >= len(args) {
				return false, false, "", "", "", fmt.Errorf("missing install dir")
			}
			installDir = args[index]
		case "--target":
			index++
			if index >= len(args) {
				return false, false, "", "", "", fmt.Errorf("missing target id")
			}
			targetID = args[index]
		default:
			if strings.HasPrefix(args[index], "-") {
				return false, false, "", "", "", fmt.Errorf("unknown flag %s", args[index])
			}
			if refValue != "" {
				return false, false, "", "", "", fmt.Errorf("multiple refs")
			}
			refValue = args[index]
		}
	}

	if refValue == "" || targetID == "" {
		return false, false, "", "", "", fmt.Errorf("missing required args")
	}

	return local, global, installDir, targetID, refValue, nil
}

func defaultInteractiveInput() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
