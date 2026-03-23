package agentref

import "testing"

func TestParseParsesExactVersionRef(t *testing.T) {
	ref, err := Parse("raul/code-reviewer@0.4.0")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if ref.Namespace != "raul" {
		t.Fatalf("Namespace = %q, want %q", ref.Namespace, "raul")
	}
	if ref.Name != "code-reviewer" {
		t.Fatalf("Name = %q, want %q", ref.Name, "code-reviewer")
	}
	if ref.Version != "0.4.0" {
		t.Fatalf("Version = %q, want %q", ref.Version, "0.4.0")
	}
}

func TestParseRejectsMalformedRefs(t *testing.T) {
	cases := []string{
		"raul/code-reviewer",
		"raul@0.4.0",
		"/code-reviewer@0.4.0",
		"raul/@0.4.0",
	}

	for _, value := range cases {
		if _, err := Parse(value); err == nil {
			t.Fatalf("Parse(%q) returned nil error", value)
		}
	}
}
