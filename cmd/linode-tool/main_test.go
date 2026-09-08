package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pixingzoudaiyuexing/linode-tool/internal/linode"
)

func TestTokenFromEnvironmentOrPromptUsesEnvironmentFirst(t *testing.T) {
	prompt := linode.NewPrompter(strings.NewReader("prompt-token\n"), &bytes.Buffer{})

	token, err := tokenFromEnvironmentOrPrompt("environment-token", prompt)
	if err != nil {
		t.Fatalf("tokenFromEnvironmentOrPrompt() error = %v", err)
	}
	if token != "environment-token" {
		t.Fatalf("token = %q, want environment-token", token)
	}
}

func TestTokenFromEnvironmentOrPromptReadsMissingTokenInteractively(t *testing.T) {
	var output bytes.Buffer
	prompt := linode.NewPrompter(strings.NewReader("prompt-token\n"), &output)

	token, err := tokenFromEnvironmentOrPrompt("", prompt)
	if err != nil {
		t.Fatalf("tokenFromEnvironmentOrPrompt() error = %v", err)
	}
	if token != "prompt-token" {
		t.Fatalf("token = %q, want prompt-token", token)
	}
	if !strings.Contains(output.String(), "Linode API Token:") {
		t.Fatalf("prompt output = %q, want token prompt", output.String())
	}
}
