package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

const prReviewSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["verdict", "summary", "findings"],
  "properties": {
    "verdict": {
      "type": "string",
      "enum": ["pass", "needs_attention"]
    },
    "summary": {
      "type": "string",
      "minLength": 1
    },
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["path", "line_start", "line_end", "severity", "title", "body"],
        "properties": {
          "path": { "type": "string", "minLength": 1 },
          "line_start": { "type": "integer", "minimum": 1 },
          "line_end": { "type": "integer", "minimum": 1 },
          "severity": {
            "type": "string",
            "enum": ["low", "medium", "high"]
          },
          "title": { "type": "string", "minLength": 1 },
          "body": { "type": "string", "minLength": 1 }
        }
      }
    }
  }
}`

type PRReview struct {
	Verdict  string      `json:"verdict"`
	Summary  string      `json:"summary"`
	Findings []PRFinding `json:"findings"`
}

type PRFinding struct {
	Path      string `json:"path"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	Severity  string `json:"severity"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

func ParsePRReview(raw []byte) (PRReview, error) {
	var review PRReview
	if err := json.Unmarshal(raw, &review); err != nil {
		return PRReview{}, fmt.Errorf("decode pr review json: %w", err)
	}
	if review.Verdict != "pass" && review.Verdict != "needs_attention" {
		return PRReview{}, fmt.Errorf("invalid verdict %q", review.Verdict)
	}
	if strings.TrimSpace(review.Summary) == "" {
		return PRReview{}, fmt.Errorf("summary is required")
	}
	for index, finding := range review.Findings {
		if strings.TrimSpace(finding.Path) == "" {
			return PRReview{}, fmt.Errorf("findings[%d].path is required", index)
		}
		if finding.LineStart <= 0 || finding.LineEnd <= 0 {
			return PRReview{}, fmt.Errorf("findings[%d] line_start and line_end must be positive", index)
		}
		if finding.LineEnd < finding.LineStart {
			return PRReview{}, fmt.Errorf("findings[%d] line_end must be >= line_start", index)
		}
		switch finding.Severity {
		case "low", "medium", "high":
		default:
			return PRReview{}, fmt.Errorf("findings[%d] invalid severity %q", index, finding.Severity)
		}
		if strings.TrimSpace(finding.Title) == "" {
			return PRReview{}, fmt.Errorf("findings[%d].title is required", index)
		}
		if strings.TrimSpace(finding.Body) == "" {
			return PRReview{}, fmt.Errorf("findings[%d].body is required", index)
		}
	}
	return review, nil
}
