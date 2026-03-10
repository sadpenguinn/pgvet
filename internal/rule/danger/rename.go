package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// Rename detects RENAME TABLE and RENAME COLUMN operations.
// Both require ACCESS EXCLUSIVE LOCK and break any existing code that
// references the old name.
type Rename struct {
	severity config.Severity
}

// NewRename returns a Rename rule with the given severity.
func NewRename(severity config.Severity) *Rename {
	return &Rename{severity: severity}
}

func (r *Rename) Name() string { return "rename" }

func (r *Rename) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *Rename) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs("RENAME") {
			continue
		}

		j := i + 1
		if j >= len(words) {
			continue
		}

		// ALTER TABLE … RENAME TO new_name
		if words[j].WordIs("TO") {
			issue := issueAt(token, r.Name(), r.severity,
				"RENAME TABLE")

			return &issue
		}

		// ALTER TABLE … RENAME COLUMN old TO new
		if words[j].WordIs("COLUMN") {
			issue := issueAt(token, r.Name(), r.severity,
				"RENAME COLUMN")

			return &issue
		}
	}

	return nil
}
