package plugin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PRContext struct {
	Number       string
	SourceBranch string
	TargetBranch string
}

func LoadPRContext() (PRContext, error) {
	ctx := PRContext{
		Number:       strings.TrimSpace(os.Getenv("DRONE_PULL_REQUEST")),
		SourceBranch: strings.TrimSpace(os.Getenv("DRONE_SOURCE_BRANCH")),
		TargetBranch: strings.TrimSpace(os.Getenv("DRONE_TARGET_BRANCH")),
	}
	if ctx.Number == "" {
		return PRContext{}, fmt.Errorf("DRONE_PULL_REQUEST is required in pr mode")
	}
	if ctx.SourceBranch == "" {
		return PRContext{}, fmt.Errorf("DRONE_SOURCE_BRANCH is required in pr mode")
	}
	if ctx.TargetBranch == "" {
		return PRContext{}, fmt.Errorf("DRONE_TARGET_BRANCH is required in pr mode")
	}
	return ctx, nil
}

func EnsureGitRepoRoot(workDir string) (string, error) {
	root, err := runGitOutput(workDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("resolve git repo root: %w", err)
	}
	return strings.TrimSpace(root), nil
}

func BuildPRDiff(workDir, targetBranch string) (string, error) {
	if strings.TrimSpace(targetBranch) == "" {
		return "", fmt.Errorf("target branch is required")
	}
	baseRef := targetBranch
	if hasOriginRemote(workDir) {
		if err := runGit(workDir, "fetch", "--quiet", "origin", targetBranch); err != nil {
			return "", fmt.Errorf("fetch target branch: %w", err)
		}
		if refExists(workDir, "refs/remotes/origin/"+targetBranch) {
			baseRef = "origin/" + targetBranch
		}
	}
	if !refExists(workDir, baseRef) {
		return "", fmt.Errorf("target branch %q not found", targetBranch)
	}
	diff, err := runGitOutput(workDir, "diff", "--merge-base", baseRef, "HEAD")
	if err != nil {
		return "", fmt.Errorf("build diff: %w", err)
	}
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("git diff is empty")
	}
	return diff, nil
}

func hasOriginRemote(dir string) bool {
	return runGit(dir, "remote", "get-url", "origin") == nil
}

func refExists(dir, ref string) bool {
	return runGit(dir, "rev-parse", "--verify", "--quiet", ref) == nil
}

func runGit(dir string, args ...string) error {
	_, err := runGitCombined(dir, args...)
	return err
}

func runGitOutput(dir string, args ...string) (string, error) {
	output, err := runGitCombined(dir, args...)
	return strings.TrimSpace(string(output)), err
}

func runGitCombined(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", bytes.TrimSpace(output), err)
	}
	return output, nil
}
