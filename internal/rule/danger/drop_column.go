package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// DropColumn detects ALTER TABLE … DROP COLUMN statements.
// Dropping a column is irreversible and permanently destroys the column data.
// On older PostgreSQL versions it also causes a full table rewrite.
type DropColumn struct {
	severity config.Severity
}

// NewDropColumn returns a DropColumn rule with the given severity.
func NewDropColumn(severity config.Severity) *DropColumn {
	return &DropColumn{severity: severity}
}

func (r *DropColumn) Name() string { return "drop-column" }

func (r *DropColumn) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *DropColumn) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	if !statement.ContainsSeq(keywordAlter, "TABLE") {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs("DROP") {
			continue
		}

		j := i + 1

		if j < len(words) && words[j].WordIs("COLUMN") {
			issue := issueAt(token, r.Name(), r.severity,
				"DROP COLUMN")

			return &issue
		}

		// DROP without COLUMN keyword but not DROP TABLE/INDEX/CONSTRAINT
		if j < len(words) && !words[j].WordIs("TABLE") && !words[j].WordIs("INDEX") && !words[j].WordIs("CONSTRAINT") {
			issue := issueAt(token, r.Name(), r.severity,
				"DROP COLUMN")

			return &issue
		}
	}

	return nil
}
