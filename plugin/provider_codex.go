package plugin

import (
	"context"
	"fmt"
)

type CodexProvider struct {
	bin    string
	runner commandRunner
}

func NewCodexProvider(bin string, runner commandRunner) Provider {
	return CodexProvider{bin: bin, runner: runner}
}

func (p CodexProvider) Run(ctx context.Context, request ProviderRequest) ([]byte, error) {
	args := []string{"exec", "-C", request.WorkDir, "--skip-git-repo-check"}
	if request.Model != "" {
		args = append(args, "--model", request.Model)
	}
	args = append(args, request.ExtraArgs...)
	if request.Mode == ModePR {
		schemaFile, cleanup, err := writeSchemaFile(request.Schema)
		if err != nil {
			return nil, err
		}
		defer cleanup()
		args = append(args, "--output-schema", schemaFile, "-")
	} else {
		args = append(args, "-")
	}
	output, err := p.runner.Run(ctx, p.bin, args, request.WorkDir, request.Prompt)
	if err != nil {
		return nil, fmt.Errorf("codex exec %v: %w", args, err)
	}
	return output, nil
}
