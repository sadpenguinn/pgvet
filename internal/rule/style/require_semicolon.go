package style

import (
	"strings"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// RequireSemicolon detects statements that are not terminated with a semicolon.
type RequireSemicolon struct {
	severity config.Severity
}

// NewRequireSemicolon returns a RequireSemicolon rule with the given severity.
func NewRequireSemicolon(severity config.Severity) *RequireSemicolon {
	return &RequireSemicolon{severity: severity}
}

func (r *RequireSemicolon) Name() string { return "require-semicolon" }

func (r *RequireSemicolon) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if !statement.HasTerminator && !hasImplicitTerminator(statement.Tokens) {
			issues = append(issues, rule.Issue{
				Line:     statement.EndLine,
				Col:      1,
				Rule:     r.Name(),
				Severity: r.severity,
				Message:  "statement is not terminated with a semicolon",
			})
		}
	}

	return issues
}

// hasImplicitTerminator reports whether the statement ends with a framework
// terminator comment (e.g. goose StatementEnd). Such blocks do not require
// an explicit semicolon.
func hasImplicitTerminator(tokens []sql.Token) bool {
	for i := len(tokens) - 1; i >= 0; i-- {
		tok := tokens[i]

		if tok.Type == sql.TokWhitespace || tok.Type == sql.TokNewline {
			continue
		}

		if tok.Type == sql.TokCommentLine {
			return strings.Contains(tok.Value, "StatementEnd")
		}

		return false
	}

	return false
}
