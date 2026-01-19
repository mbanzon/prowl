package main

import (
	"os"
	"strings"
	"testing"
)

func TestReadmeMentionsCurrentFlags(t *testing.T) {
	content, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}

	readme := string(content)
	if !strings.Contains(readme, "`-protect`") {
		t.Fatalf("README.md should mention the -protect flag")
	}
	if !strings.Contains(readme, "Default is `5001`") {
		t.Fatalf("README.md should mention default port 5001")
	}
}
