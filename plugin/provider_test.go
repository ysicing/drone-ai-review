package plugin

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCodexProviderBuildsExecCommandForPR(t *testing.T) {
	runner := &fakeRunner{err: fmt.Errorf("boom")}
	provider := NewCodexProvider("codex", runner)

	_, err := provider.Run(context.Background(), ProviderRequest{
		Mode:    ModePR,
		WorkDir: "/repo",
		Prompt:  "review this diff",
		Model:   "gpt-5",
		Schema:  prReviewSchema,
	})
	if err == nil {
		t.Fatal("expected fake runner error")
	}
	if !strings.Contains(err.Error(), "codex exec") {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.name != "codex" {
		t.Fatalf("name = %q", runner.name)
	}
	if !containsArg(runner.args, "--output-schema") {
		t.Fatalf("missing --output-schema in %v", runner.args)
	}
}

func TestClaudeProviderBuildsPrintCommandForPR(t *testing.T) {
	runner := &fakeRunner{err: fmt.Errorf("boom")}
	provider := NewClaudeProvider("claude", runner)

	_, err := provider.Run(context.Background(), ProviderRequest{
		Mode:    ModePR,
		WorkDir: "/repo",
		Prompt:  "review this diff",
		Model:   "sonnet",
		Schema:  prReviewSchema,
	})
	if err == nil {
		t.Fatal("expected fake runner error")
	}
	if !strings.Contains(err.Error(), "--json-schema") {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.name != "claude" {
		t.Fatalf("name = %q", runner.name)
	}
}

type fakeRunner struct {
	name  string
	args  []string
	dir   string
	stdin string
	err   error
}

func (r *fakeRunner) Run(_ context.Context, name string, args []string, dir string, stdin string) ([]byte, error) {
	r.name = name
	r.args = append([]string(nil), args...)
	r.dir = dir
	r.stdin = stdin
	if r.err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), r.err)
	}
	return []byte(`ok`), nil
}

func containsArg(args []string, expected string) bool {
	for _, arg := range args {
		if arg == expected {
			return true
		}
	}
	return false
}
