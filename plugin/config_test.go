package plugin

import (
	"strings"
	"testing"
)

func TestParseConfigRequiresMode(t *testing.T) {
	t.Setenv("PLUGIN_MODE", "")
	t.Setenv("PLUGIN_PROVIDER", "codex")
	t.Setenv("PLUGIN_OUTPUT", "review.json")

	_, err := ParseConfig()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "PLUGIN_MODE") {
		t.Fatalf("expected PLUGIN_MODE error, got %v", err)
	}
}

func TestParseConfigRequiresProvider(t *testing.T) {
	t.Setenv("PLUGIN_MODE", "full")
	t.Setenv("PLUGIN_PROVIDER", "")
	t.Setenv("PLUGIN_OUTPUT", "review.md")

	_, err := ParseConfig()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "PLUGIN_PROVIDER") {
		t.Fatalf("expected PLUGIN_PROVIDER error, got %v", err)
	}
}

func TestParseConfigAppliesDefaults(t *testing.T) {
	t.Setenv("PLUGIN_MODE", "full")
	t.Setenv("PLUGIN_PROVIDER", "codex")
	t.Setenv("PLUGIN_OUTPUT", "review.md")
	t.Setenv("PLUGIN_WORKDIR", "")
	t.Setenv("PLUGIN_INCLUDE", "go.mod,README.md")
	t.Setenv("PLUGIN_EXCLUDE", ".git,vendor")

	cfg, err := ParseConfig()
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if cfg.Mode != ModeFull {
		t.Fatalf("Mode = %q", cfg.Mode)
	}
	if cfg.WorkDir == "" {
		t.Fatal("expected default workdir")
	}
	if cfg.CLIBin != "codex" {
		t.Fatalf("CLIBin = %q", cfg.CLIBin)
	}
	if len(cfg.Include) != 2 {
		t.Fatalf("Include len = %d", len(cfg.Include))
	}
	if len(cfg.Exclude) != 2 {
		t.Fatalf("Exclude len = %d", len(cfg.Exclude))
	}
}
