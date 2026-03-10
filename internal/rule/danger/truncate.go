package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// Truncate detects TRUNCATE TABLE statements.
// TRUNCATE
// for the duration of the operation.
type Truncate struct {
	severity config.Severity
}

// NewTruncate returns a Truncate rule with the given severity.
func NewTruncate(severity config.Severity) *Truncate {
	return &Truncate{severity: severity}
}

func (r *Truncate) Name() string { return "truncate" }

func (r *Truncate) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *Truncate) checkStatement(statement *sql.Statement) *rule.Issue {
	for _, token := range statement.Words() {
		if token.WordIs("TRUNCATE") {
			issue := issueAt(token, r.Name(), r.severity,
				"TRUNCATE")

			return &issue
		}
	}

	return nil
}
