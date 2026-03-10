package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// AddForeignKey detects ADD FOREIGN KEY without NOT VALID.
// Without NOT VALID PostgreSQL scans the entire table to validate existing
// rows, holding an ACCESS EXCLUSIVE lock for the whole duration.
// Use NOT VALID to skip the scan and follow up with VALIDATE CONSTRAINT.
type AddForeignKey struct {
	severity config.Severity
}

// NewAddForeignKey returns an AddForeignKey rule with the given severity.
func NewAddForeignKey(severity config.Severity) *AddForeignKey {
	return &AddForeignKey{severity: severity}
}

func (r *AddForeignKey) Name() string { return "add-foreign-key-no-valid" }

func (r *AddForeignKey) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *AddForeignKey) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	if !statement.ContainsSeq("FOREIGN", "KEY") {
		return nil
	}

	if statement.ContainsSeq("NOT", "VALID") {
		return nil
	}

	firstTok, _ := statement.WordAt(0)
	issue := issueAt(firstTok, r.Name(), r.severity,
		"ADD FOREIGN KEY without NOT VALID")

	return &issue
}
