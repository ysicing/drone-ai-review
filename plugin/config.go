package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Mode string

const (
	ModePR   Mode = "pr"
	ModeFull Mode = "full"
)

type ProviderName string

const (
	ProviderCodex  ProviderName = "codex"
	ProviderClaude ProviderName = "claude"
)

type Config struct {
	Mode       Mode
	Provider   ProviderName
	Output     string
	WorkDir    string
	PromptFile string
	CLIBin     string
	Model      string
	ExtraArgs  []string
	Include    []string
	Exclude    []string
	Debug      bool
}

func ParseConfig() (Config, error) {
	cfg := Config{
		Mode:       Mode(strings.TrimSpace(os.Getenv("PLUGIN_MODE"))),
		Provider:   ProviderName(strings.TrimSpace(os.Getenv("PLUGIN_PROVIDER"))),
		Output:     strings.TrimSpace(os.Getenv("PLUGIN_OUTPUT")),
		PromptFile: strings.TrimSpace(os.Getenv("PLUGIN_PROMPT_FILE")),
		CLIBin:     strings.TrimSpace(os.Getenv("PLUGIN_CLI_BIN")),
		Model:      strings.TrimSpace(os.Getenv("PLUGIN_MODEL")),
		Debug:      os.Getenv("PLUGIN_DEBUG") != "",
	}

	wd := strings.TrimSpace(os.Getenv("PLUGIN_WORKDIR"))
	if wd == "" {
		current, err := os.Getwd()
		if err != nil {
			return Config{}, fmt.Errorf("get workdir: %w", err)
		}
		wd = current
	}
	cfg.WorkDir = filepath.Clean(wd)
	cfg.ExtraArgs = splitList(os.Getenv("PLUGIN_EXTRA_ARGS"))
	cfg.Include = splitList(os.Getenv("PLUGIN_INCLUDE"))
	cfg.Exclude = splitList(os.Getenv("PLUGIN_EXCLUDE"))

	if cfg.Mode != ModePR && cfg.Mode != ModeFull {
		return Config{}, fmt.Errorf("PLUGIN_MODE must be one of: pr, full")
	}
	if cfg.Provider != ProviderCodex && cfg.Provider != ProviderClaude {
		return Config{}, fmt.Errorf("PLUGIN_PROVIDER must be one of: codex, claude")
	}
	if cfg.Output == "" {
		return Config{}, fmt.Errorf("PLUGIN_OUTPUT is required")
	}
	if cfg.CLIBin == "" {
		cfg.CLIBin = string(cfg.Provider)
	}
	return cfg, nil
}

func splitList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		item := strings.TrimSpace(field)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}
