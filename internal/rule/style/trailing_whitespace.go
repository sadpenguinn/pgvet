package style

import (
	"strings"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
)

// TrailingWhitespace detects lines that end with spaces or tabs.
type TrailingWhitespace struct {
	severity config.Severity
}

// NewTrailingWhitespace returns a TrailingWhitespace rule with the given severity.
func NewTrailingWhitespace(severity config.Severity) *TrailingWhitespace {
	return &TrailingWhitespace{severity: severity}
}

func (r *TrailingWhitespace) Name() string { return "trailing-whitespace" }

func (r *TrailingWhitespace) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for i, line := range file.Lines {
		trimmed := strings.TrimRight(line, " \t")
		if len(trimmed) == len(line) {
			continue
		}

		issues = append(issues, rule.Issue{
			Line:     i + 1,
			Col:      len(trimmed) + 1,
			Rule:     r.Name(),
			Severity: r.severity,
			Message:  "trailing whitespace",
		})
	}

	return issues
}
