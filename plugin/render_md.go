package plugin

import (
	"fmt"
	"strings"
)

func ValidateMarkdown(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("full review output is empty")
	}
	return nil
}
