package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/install"
	"github.com/agentlibdev/agent-cli/internal/manifest"
	"github.com/agentlibdev/agent-cli/internal/registry"
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
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return app{
		newRegistryClient: func(baseURL string) registryClient {
			return registry.New(baseURL)
		},
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
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib install <namespace/name@version>")
		return 1
	}

	ref, err := agentref.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	result, err := install.Run(ctx, a.registryClient(), workingDir, ref)
	if err != nil {
		fmt.Fprintf(stderr, "install agent: %v\n", err)
		return 1
	}

	lockfile := filepath.Join(workingDir, ".agentlib", "agent.lock.json")
	fmt.Fprintf(stdout, "installed: %s/%s@%s\n", ref.Namespace, ref.Name, ref.Version)
	fmt.Fprintf(stdout, "root: %s\n", result.Root)
	fmt.Fprintf(stdout, "lockfile: %s\n", lockfile)
	return 0
}

func runRemove(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib remove <namespace/name@version>")
		return 1
	}

	ref, err := agentref.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve working directory: %v\n", err)
		return 1
	}

	if err := install.Remove(workingDir, ref); err != nil {
		fmt.Fprintf(stderr, "remove agent: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "removed: %s/%s@%s\n", ref.Namespace, ref.Name, ref.Version)
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
	fmt.Fprintln(writer, "  search <query>")
	fmt.Fprintln(writer, "  show <namespace/name@version>")
	fmt.Fprintln(writer, "  install <namespace/name@version>")
	fmt.Fprintln(writer, "  remove <namespace/name@version>")
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
