package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/agentlibdev/agent-cli/internal/agentref"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Version struct {
	Namespace     string
	Name          string
	Version       string
	Title         string
	Description   string
	License       string
	ManifestJSON  string
	PublishedAt   string
	Compatibility Compatibility
}

type Compatibility struct {
	Targets []TargetCompatibility `json:"targets"`
}

type TargetCompatibility struct {
	TargetID         string `json:"targetId"`
	BuiltFor         bool   `json:"builtFor"`
	Tested           bool   `json:"tested"`
	AdapterAvailable bool   `json:"adapterAvailable"`
}

type AgentSummary struct {
	Namespace     string
	Name          string
	LatestVersion string
	Title         string
	Description   string
}

type Artifact struct {
	Path      string
	MediaType string
	SizeBytes int64
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func New(baseURL string) Client {
	return NewWithHTTPClient(baseURL, http.DefaultClient)
}

func NewWithHTTPClient(baseURL string, httpClient *http.Client) Client {
	return Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c Client) FetchVersion(ctx context.Context, ref agentref.Ref) (Version, error) {
	var body struct {
		Version Version `json:"version"`
	}

	path := fmt.Sprintf("%s/api/v1/agents/%s/%s/versions/%s", c.baseURL, ref.Namespace, ref.Name, ref.Version)
	if err := c.getJSON(ctx, path, &body); err != nil {
		return Version{}, err
	}

	return body.Version, nil
}

func (c Client) FetchArtifacts(ctx context.Context, ref agentref.Ref) ([]Artifact, error) {
	var body struct {
		Items []Artifact `json:"items"`
	}

	path := fmt.Sprintf("%s/api/v1/agents/%s/%s/versions/%s/artifacts", c.baseURL, ref.Namespace, ref.Name, ref.Version)
	if err := c.getJSON(ctx, path, &body); err != nil {
		return nil, err
	}

	return body.Items, nil
}

func (c Client) FetchAgents(ctx context.Context) ([]AgentSummary, error) {
	var body struct {
		Items []AgentSummary `json:"items"`
	}

	path := fmt.Sprintf("%s/api/v1/agents", c.baseURL)
	if err := c.getJSON(ctx, path, &body); err != nil {
		return nil, err
	}

	return body.Items, nil
}

func (c Client) DownloadArtifact(ctx context.Context, ref agentref.Ref, path string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/%s/versions/%s/artifacts/%s", c.baseURL, ref.Namespace, ref.Name, ref.Version, path)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, "", readAPIError(response)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	return content, response.Header.Get("content-type"), nil
}

func (c Client) getJSON(ctx context.Context, url string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return readAPIError(response)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func readAPIError(response *http.Response) error {
	var body apiError
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return fmt.Errorf("registry request failed with status %d", response.StatusCode)
	}

	if body.Error.Message == "" {
		return fmt.Errorf("registry request failed with status %d", response.StatusCode)
	}

	return fmt.Errorf("%s", body.Error.Message)
}
