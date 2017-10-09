package auth

import (
	"strings"
	"testing"
)

func TestIssueToken(t *testing.T) {
	token := IssueToken(123, "unit-test", false)
	if token == "" {
		t.Error("Token is empty")
	}
	vals := strings.Split(token, ".")
	if len(vals) != 2 {
		t.Errorf("Unexpected number of token parts: %d", len(vals))
	}
}
