package plugin

import (
	"context"
	"fmt"
)

type ClaudeProvider struct {
	bin    string
	runner commandRunner
}

func NewClaudeProvider(bin string, runner commandRunner) Provider {
	return ClaudeProvider{bin: bin, runner: runner}
}

func (p ClaudeProvider) Run(ctx context.Context, request ProviderRequest) ([]byte, error) {
	args := []string{"-p", "--permission-mode", "bypassPermissions", "--output-format", "text"}
	if request.Mode == ModePR {
		args = append(args, "--json-schema", request.Schema)
	}
	if request.Model != "" {
		args = append(args, "--model", request.Model)
	}
	args = append(args, request.ExtraArgs...)
	output, err := p.runner.Run(ctx, p.bin, args, request.WorkDir, request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("claude -p %v: %w", args, err)
	}
	return output, nil
}
