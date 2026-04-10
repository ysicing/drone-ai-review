package plugin

import (
	"strings"
	"testing"
)

func TestBuildPRPromptContainsSchemaRules(t *testing.T) {
	prompt := BuildPRPrompt(PRPromptInput{
		RepoRoot: "/tmp/repo",
		Diff:     "diff --git a/a.go b/a.go",
	})

	if !strings.Contains(prompt, `"verdict"`) {
		t.Fatalf("prompt missing schema: %s", prompt)
	}
	if !strings.Contains(prompt, "line_start") {
		t.Fatalf("prompt missing line rules: %s", prompt)
	}
}

func TestBuildFullPromptContainsRepoRoot(t *testing.T) {
	prompt := BuildFullPrompt(FullPromptInput{
		RepoRoot: "/tmp/repo",
	})

	if !strings.Contains(prompt, "/tmp/repo") {
		t.Fatalf("prompt missing repo root: %s", prompt)
	}
}
