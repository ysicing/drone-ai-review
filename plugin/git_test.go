package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPRContextRequiresDronePREnv(t *testing.T) {
	t.Setenv("DRONE_PULL_REQUEST", "")
	t.Setenv("DRONE_SOURCE_BRANCH", "")
	t.Setenv("DRONE_TARGET_BRANCH", "")

	_, err := LoadPRContext()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "DRONE_PULL_REQUEST") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPRDiffUsesTargetBranch(t *testing.T) {
	repoDir := initGitRepoWithFeatureBranch(t)

	diff, err := BuildPRDiff(repoDir, "main")
	if err != nil {
		t.Fatalf("BuildPRDiff() error = %v", err)
	}
	if !strings.Contains(diff, "+++ b/app.txt") {
		t.Fatalf("unexpected diff: %s", diff)
	}
}

func initGitRepoWithFeatureBranch(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()
	runGitTest(t, repoDir, "init", "-b", "main")
	runGitTest(t, repoDir, "config", "user.email", "ci@example.com")
	runGitTest(t, repoDir, "config", "user.name", "CI")
	writeTestFile(t, filepath.Join(repoDir, "app.txt"), "base\n")
	runGitTest(t, repoDir, "add", "app.txt")
	runGitTest(t, repoDir, "commit", "-m", "base")
	runGitTest(t, repoDir, "checkout", "-b", "feature")
	writeTestFile(t, filepath.Join(repoDir, "app.txt"), "base\nfeature\n")
	runGitTest(t, repoDir, "commit", "-am", "feature")
	return repoDir
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	if err := runGit(dir, args...); err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
