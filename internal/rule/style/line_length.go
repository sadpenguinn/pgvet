package style

import (
	"fmt"
	"unicode/utf8"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
)

// MaxLineLength detects lines that exceed the configured character limit.
type MaxLineLength struct {
	severity config.Severity
	maxLen   int
}

// NewMaxLineLength returns a MaxLineLength rule with the given severity and limit.
func NewMaxLineLength(severity config.Severity, maxLen int) *MaxLineLength {
	if maxLen <= 0 {
		maxLen = 120
	}

	return &MaxLineLength{severity: severity, maxLen: maxLen}
}

func (r *MaxLineLength) Name() string { return "max-line-length" }

func (r *MaxLineLength) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for i, line := range file.Lines {
		length := utf8.RuneCountInString(line)
		if length <= r.maxLen {
			continue
		}

		issues = append(issues, rule.Issue{
			Line:     i + 1,
			Col:      r.maxLen + 1,
			Rule:     r.Name(),
			Severity: r.severity,
			Message:  fmt.Sprintf("line length %d exceeds maximum %d", length, r.maxLen),
		})
	}

	return issues
}
