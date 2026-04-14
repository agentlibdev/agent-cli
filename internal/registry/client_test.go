package registry

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestClientFetchesVersionAndArtifacts(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			switch request.URL.Path {
			case "/api/v1/agents/raul/code-reviewer/versions/0.4.0":
				return jsonResponse(`{"version":{"namespace":"raul","name":"code-reviewer","version":"0.4.0","title":"Code Reviewer","description":"Reviews code changes.","license":"MIT","manifestJson":"{}","publishedAt":"2026-03-23T00:00:00Z","compatibility":{"targets":[{"targetId":"codex","builtFor":true,"tested":true,"adapterAvailable":true},{"targetId":"claude-code","builtFor":false,"tested":false,"adapterAvailable":true}]}}}`), nil
			case "/api/v1/agents/raul/code-reviewer/versions/0.4.0/artifacts":
				return jsonResponse(`{"items":[{"path":"agent.yaml","mediaType":"application/yaml","sizeBytes":12},{"path":"README.md","mediaType":"text/markdown","sizeBytes":24}]}`), nil
			default:
				return notFoundResponse(), nil
			}
		}),
	}

	client := NewWithHTTPClient("https://agentlib.dev", httpClient)
	ref, err := agentref.Parse("raul/code-reviewer@0.4.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	version, err := client.FetchVersion(t.Context(), ref)
	if err != nil {
		t.Fatalf("FetchVersion returned error: %v", err)
	}
	if version.Title != "Code Reviewer" {
		t.Fatalf("Title = %q, want %q", version.Title, "Code Reviewer")
	}
	if len(version.Compatibility.Targets) != 2 {
		t.Fatalf("len(version.Compatibility.Targets) = %d, want 2", len(version.Compatibility.Targets))
	}
	if version.Compatibility.Targets[0].TargetID != "codex" || !version.Compatibility.Targets[0].BuiltFor {
		t.Fatalf("compatibility target = %+v", version.Compatibility.Targets[0])
	}

	artifacts, err := client.FetchArtifacts(t.Context(), ref)
	if err != nil {
		t.Fatalf("FetchArtifacts returned error: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("len(artifacts) = %d, want 2", len(artifacts))
	}
}

func TestClientFetchesAgents(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			switch request.URL.Path {
			case "/api/v1/agents":
				return jsonResponse(`{"items":[{"namespace":"raul","name":"code-reviewer","latestVersion":"0.4.0","title":"Code Reviewer","description":"Reviews code changes."},{"namespace":"raul","name":"docs-writer","latestVersion":"0.2.0","title":"Docs Writer","description":"Drafts documentation."}],"nextCursor":null}`), nil
			default:
				return notFoundResponse(), nil
			}
		}),
	}

	client := NewWithHTTPClient("https://agentlib.dev", httpClient)

	agents, err := client.FetchAgents(t.Context())
	if err != nil {
		t.Fatalf("FetchAgents returned error: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("len(agents) = %d, want 2", len(agents))
	}
	if agents[0].Name != "code-reviewer" {
		t.Fatalf("agents[0].Name = %q, want %q", agents[0].Name, "code-reviewer")
	}
}

func TestClientReturnsErrorForUpstreamFailures(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return notFoundResponse(), nil
		}),
	}

	client := NewWithHTTPClient("https://agentlib.dev", httpClient)
	ref, err := agentref.Parse("raul/code-reviewer@0.4.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if _, err := client.FetchVersion(t.Context(), ref); err == nil {
		t.Fatal("FetchVersion returned nil error")
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"content-type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

func notFoundResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header: http.Header{
			"content-type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(`{"error":{"code":"not_found","message":"missing"}}`)),
	}
}
