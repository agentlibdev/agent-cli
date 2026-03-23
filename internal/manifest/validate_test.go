package manifest

import "testing"

const validManifest = `
apiVersion: agentlib.dev/v1alpha1
kind: Agent

metadata:
  namespace: raul
  name: code-reviewer
  version: 0.4.0
  title: Code Reviewer
  description: Reviews code changes.
`

func TestValidateAcceptsRequiredFields(t *testing.T) {
	if _, err := ValidateYAML([]byte(validManifest)); err != nil {
		t.Fatalf("ValidateYAML returned error: %v", err)
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	invalid := `
apiVersion: agentlib.dev/v1alpha1
kind: Agent

metadata:
  namespace: raul
`

	if _, err := ValidateYAML([]byte(invalid)); err == nil {
		t.Fatal("ValidateYAML returned nil error")
	}
}
