package plugin

import (
	"fmt"
	"os"
	"strings"
)

type PRPromptInput struct {
	RepoRoot              string
	Diff                  string
	Include               []string
	Exclude               []string
	AdditionalInstruction string
}

type FullPromptInput struct {
	RepoRoot              string
	Include               []string
	Exclude               []string
	AdditionalInstruction string
}

func LoadPromptOverride(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read prompt file: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

func BuildPRPrompt(input PRPromptInput) string {
	return fmt.Sprintf(`Review the git diff for repository %s.
Return JSON only using this schema:
{
  "verdict": "pass" | "needs_attention",
  "summary": "short review summary",
  "findings": [
    {
      "path": "relative/file/path",
      "line_start": 1,
      "line_end": 1,
      "severity": "low" | "medium" | "high",
      "title": "short title",
      "body": "actionable explanation"
    }
  ]
}

Rules:
- Focus on correctness, bugs, security, and maintainability risks.
- Only report findings worth developer attention.
- path must be repository-relative.
- line_start and line_end must point to changed lines in the diff.
- Do not wrap the JSON in markdown fences.
- Include patterns: %s
- Exclude patterns: %s

Additional instructions:
%s

Diff:
%s`,
		input.RepoRoot,
		joinOrDefault(input.Include, "(all changed files)"),
		joinOrDefault(input.Exclude, "(none)"),
		defaultInstruction(input.AdditionalInstruction),
		input.Diff,
	)
}

func BuildFullPrompt(input FullPromptInput) string {
	return fmt.Sprintf(`Review the repository at %s.
Produce Markdown with these sections:
- Summary
- High Risk Findings
- Improvements
- Suggested Priorities

Rules:
- Focus on high-value issues instead of style nitpicks.
- Respect include patterns: %s
- Respect exclude patterns: %s

Additional instructions:
%s`,
		input.RepoRoot,
		joinOrDefault(input.Include, "(entire repository)"),
		joinOrDefault(input.Exclude, "(none)"),
		defaultInstruction(input.AdditionalInstruction),
	)
}

func defaultInstruction(value string) string {
	if strings.TrimSpace(value) == "" {
		return "No additional instructions."
	}
	return value
}

func joinOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ", ")
}
