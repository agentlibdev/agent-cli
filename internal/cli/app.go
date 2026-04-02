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
	newRegistryClient func(baseURL string) registryClient
	loadTargets       func(projectDir string) ([]targets.Target, error)
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return app{
		newRegistryClient: func(baseURL string) registryClient {
			return registry.New(baseURL)
		},
		loadTargets: targets.Load,
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
	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(stderr, "usage: agentlib install [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
		return 1
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: agentlib install [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
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
	if len(args) != 1 || args[0] != "list" {
		fmt.Fprintln(stderr, "usage: agentlib targets list")
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

	for _, item := range items {
		state := "enabled"
		if !item.Enabled {
			state = "disabled"
		}
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Type, item.Format, item.Mode, state)
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
	fmt.Fprintln(writer, "  install [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>")
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
