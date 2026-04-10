package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutePRModeWritesValidatedJSON(t *testing.T) {
	output := filepath.Join(t.TempDir(), "review.json")
	cfg := Config{
		Mode:     ModePR,
		Provider: ProviderCodex,
		Output:   output,
		WorkDir:  t.TempDir(),
	}

	err := execute(cfg, fakeDeps{
		repoRoot: "/repo",
		prContext: PRContext{
			Number:       "1",
			SourceBranch: "feature",
			TargetBranch: "main",
		},
		diff:           "diff --git a/main.go b/main.go",
		providerOutput: []byte(`{"verdict":"pass","summary":"clean","findings":[]}`),
	})
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	assertFileContains(t, output, `"verdict": "pass"`)
}

func TestExecuteFullModeWritesMarkdown(t *testing.T) {
	output := filepath.Join(t.TempDir(), "review.md")
	cfg := Config{
		Mode:     ModeFull,
		Provider: ProviderClaude,
		Output:   output,
		WorkDir:  t.TempDir(),
	}

	err := execute(cfg, fakeDeps{
		repoRoot:       "/repo",
		providerOutput: []byte("# Summary\n\nLooks good.\n"),
	})
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	assertFileContains(t, output, "# Summary")
}

func TestExecuteResolvesRelativeOutputFromRepoRoot(t *testing.T) {
	repoRoot := t.TempDir()
	cfg := Config{
		Mode:     ModeFull,
		Provider: ProviderClaude,
		Output:   "nested/review.md",
		WorkDir:  repoRoot,
	}

	err := execute(cfg, fakeDeps{
		repoRoot:       repoRoot,
		providerOutput: []byte("# Summary\n"),
	})
	if err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	assertFileContains(t, filepath.Join(repoRoot, "nested", "review.md"), "# Summary")
}

type fakeDeps struct {
	repoRoot       string
	prContext      PRContext
	diff           string
	override       string
	providerOutput []byte
}

func (f fakeDeps) EnsureGitRepoRoot(string) (string, error) { return f.repoRoot, nil }
func (f fakeDeps) LoadPRContext() (PRContext, error)        { return f.prContext, nil }
func (f fakeDeps) BuildPRDiff(string, string) (string, error) {
	return f.diff, nil
}
func (f fakeDeps) LoadPromptOverride(string) (string, error) { return f.override, nil }
func (f fakeDeps) NewProvider(Config) (Provider, error) {
	return fakeProvider{output: f.providerOutput}, nil
}
func (f fakeDeps) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

type fakeProvider struct {
	output []byte
}

func (f fakeProvider) Run(_ context.Context, request ProviderRequest) ([]byte, error) {
	if strings.TrimSpace(request.Prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}
	return f.output, nil
}

func assertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if !strings.Contains(string(body), expected) {
		t.Fatalf("file %s missing %q, got %s", path, expected, string(body))
	}
}
