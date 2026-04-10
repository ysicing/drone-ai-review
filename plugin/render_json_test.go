package plugin

import "testing"

func TestParsePRReviewValidatesRequiredFields(t *testing.T) {
	raw := `{"verdict":"needs_attention","summary":"x","findings":[{"path":"","line_start":0,"line_end":0,"severity":"low","title":"","body":""}]}`

	_, err := ParsePRReview([]byte(raw))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParsePRReviewAcceptsValidPayload(t *testing.T) {
	raw := `{"verdict":"pass","summary":"clean","findings":[{"path":"main.go","line_start":8,"line_end":8,"severity":"medium","title":"Avoid panic","body":"return an error instead"}]}`

	review, err := ParsePRReview([]byte(raw))
	if err != nil {
		t.Fatalf("ParsePRReview() error = %v", err)
	}
	if review.Verdict != "pass" {
		t.Fatalf("Verdict = %q", review.Verdict)
	}
	if len(review.Findings) != 1 {
		t.Fatalf("Findings len = %d", len(review.Findings))
	}
}

func TestValidateMarkdownRejectsEmptyOutput(t *testing.T) {
	err := ValidateMarkdown("   ")
	if err == nil {
		t.Fatal("expected error")
	}
}
