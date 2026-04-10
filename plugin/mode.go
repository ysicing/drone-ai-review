package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type executorDeps interface {
	EnsureGitRepoRoot(string) (string, error)
	LoadPRContext() (PRContext, error)
	BuildPRDiff(string, string) (string, error)
	LoadPromptOverride(string) (string, error)
	NewProvider(Config) (Provider, error)
	WriteFile(string, []byte) error
}

type defaultExecutorDeps struct{}

func (defaultExecutorDeps) EnsureGitRepoRoot(workDir string) (string, error) {
	return EnsureGitRepoRoot(workDir)
}
func (defaultExecutorDeps) LoadPRContext() (PRContext, error) { return LoadPRContext() }
func (defaultExecutorDeps) BuildPRDiff(workDir, targetBranch string) (string, error) {
	return BuildPRDiff(workDir, targetBranch)
}
func (defaultExecutorDeps) LoadPromptOverride(path string) (string, error) {
	return LoadPromptOverride(path)
}
func (defaultExecutorDeps) NewProvider(cfg Config) (Provider, error) { return NewProvider(cfg) }
func (defaultExecutorDeps) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func (p Plugin) Exec() error {
	cfg, err := ParseConfig()
	if err != nil {
		return err
	}
	return execute(cfg, defaultExecutorDeps{})
}

func execute(cfg Config, deps executorDeps) error {
	repoRoot, err := deps.EnsureGitRepoRoot(cfg.WorkDir)
	if err != nil {
		return err
	}
	var override string
	if cfg.PromptFile != "" {
		override, err = deps.LoadPromptOverride(resolvePath(repoRoot, cfg.PromptFile))
		if err != nil {
			return err
		}
	}
	outputPath := resolvePath(repoRoot, cfg.Output)
	provider, err := deps.NewProvider(cfg)
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case ModePR:
		prCtx, err := deps.LoadPRContext()
		if err != nil {
			return err
		}
		diff, err := deps.BuildPRDiff(repoRoot, prCtx.TargetBranch)
		if err != nil {
			return err
		}
		output, err := provider.Run(context.Background(), ProviderRequest{
			Mode:    cfg.Mode,
			WorkDir: repoRoot,
			Prompt: BuildPRPrompt(PRPromptInput{
				RepoRoot:              repoRoot,
				Diff:                  diff,
				Include:               cfg.Include,
				Exclude:               cfg.Exclude,
				AdditionalInstruction: override,
			}),
			Model:     cfg.Model,
			Schema:    prReviewSchema,
			ExtraArgs: cfg.ExtraArgs,
		})
		if err != nil {
			return err
		}
		review, err := ParsePRReview(output)
		if err != nil {
			return err
		}
		body, err := json.MarshalIndent(review, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal pr review: %w", err)
		}
		return deps.WriteFile(outputPath, body)
	case ModeFull:
		output, err := provider.Run(context.Background(), ProviderRequest{
			Mode:    cfg.Mode,
			WorkDir: repoRoot,
			Prompt: BuildFullPrompt(FullPromptInput{
				RepoRoot:              repoRoot,
				Include:               cfg.Include,
				Exclude:               cfg.Exclude,
				AdditionalInstruction: override,
			}),
			Model:     cfg.Model,
			ExtraArgs: cfg.ExtraArgs,
		})
		if err != nil {
			return err
		}
		if err := ValidateMarkdown(string(output)); err != nil {
			return err
		}
		return deps.WriteFile(outputPath, output)
	default:
		return fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}

func resolvePath(repoRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(repoRoot, path)
}
