package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/agentlibdev/agent-cli/internal/agentref"
	"github.com/agentlibdev/agent-cli/internal/install"
	"github.com/agentlibdev/agent-cli/internal/manifest"
	"github.com/agentlibdev/agent-cli/internal/registry"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 1
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "show":
		return runShow(ctx, args[1:], stdout, stderr)
	case "install":
		return runInstall(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 1
	}
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

func runShow(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: agentlib show <namespace/name@version>")
		return 1
	}

	ref, err := agentref.Parse(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "parse ref: %v\n", err)
		return 1
	}

	client := registry.New(registryBaseURL())
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

func runInstall(ctx context.Context, args []string, stdout, stderr io.Writer) int {
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

	result, err := install.Run(ctx, registry.New(registryBaseURL()), workingDir, ref)
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

func registryBaseURL() string {
	if value := os.Getenv("AGENTLIB_BASE_URL"); value != "" {
		return value
	}

	return "https://agentlib.dev"
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "agentlib <command>")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Commands:")
	fmt.Fprintln(writer, "  validate <path>")
	fmt.Fprintln(writer, "  show <namespace/name@version>")
	fmt.Fprintln(writer, "  install <namespace/name@version>")
}
