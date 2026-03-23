package agentref

import (
	"fmt"
	"strings"
)

type Ref struct {
	Namespace string
	Name      string
	Version   string
}

func Parse(value string) (Ref, error) {
	versionParts := strings.Split(value, "@")
	if len(versionParts) != 2 || versionParts[0] == "" || versionParts[1] == "" {
		return Ref{}, fmt.Errorf("invalid agent ref %q", value)
	}

	nameParts := strings.Split(versionParts[0], "/")
	if len(nameParts) != 2 || nameParts[0] == "" || nameParts[1] == "" {
		return Ref{}, fmt.Errorf("invalid agent ref %q", value)
	}

	return Ref{
		Namespace: nameParts[0],
		Name:      nameParts[1],
		Version:   versionParts[1],
	}, nil
}
