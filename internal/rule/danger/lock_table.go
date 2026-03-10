package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// LockTable detects explicit LOCK TABLE statements.
// Explicit table locks block other transactions and can cause deadlocks
// or service downtime in production environments.
type LockTable struct {
	severity config.Severity
}

// NewLockTable returns a LockTable rule with the given severity.
func NewLockTable(severity config.Severity) *LockTable {
	return &LockTable{severity: severity}
}

func (r *LockTable) Name() string { return "lock-table" }

func (r *LockTable) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *LockTable) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != "LOCK" {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if token.WordIs("LOCK") && i+1 < len(words) && words[i+1].WordIs("TABLE") {
			issue := issueAt(token, r.Name(), r.severity,
				"LOCK TABLE")

			return &issue
		}
	}

	return nil
}
