package plugin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ProviderRequest struct {
	Mode      Mode
	WorkDir   string
	Prompt    string
	Model     string
	Schema    string
	ExtraArgs []string
}

type Provider interface {
	Run(ctx context.Context, request ProviderRequest) ([]byte, error)
}

type commandRunner interface {
	Run(ctx context.Context, name string, args []string, dir string, stdin string) ([]byte, error)
}

func NewProvider(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case ProviderCodex:
		return NewCodexProvider(cfg.CLIBin, execRunner{}), nil
	case ProviderClaude:
		return NewClaudeProvider(cfg.CLIBin, execRunner{}), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", cfg.Provider)
	}
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args []string, dir string, stdin string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(bytes.TrimSpace(output)) == 0 {
			return nil, err
		}
		return nil, fmt.Errorf("%s", bytes.TrimSpace(output))
	}
	return output, nil
}

func writeSchemaFile(schema string) (string, func(), error) {
	file, err := os.CreateTemp("", "drone-ai-review-schema-*.json")
	if err != nil {
		return "", nil, fmt.Errorf("create schema file: %w", err)
	}
	if _, err := file.WriteString(schema); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("write schema file: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("close schema file: %w", err)
	}
	cleanup := func() {
		_ = os.Remove(file.Name())
	}
	return file.Name(), cleanup, nil
}
