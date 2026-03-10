package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// ChangeColumnType detects ALTER COLUMN TYPE operations.
// Changing a column type triggers a full table rewrite and holds
// ACCESS EXCLUSIVE LOCK for the entire duration, blocking all queries.
type ChangeColumnType struct {
	severity config.Severity
}

// NewChangeColumnType returns a ChangeColumnType rule with the given severity.
func NewChangeColumnType(severity config.Severity) *ChangeColumnType {
	return &ChangeColumnType{severity: severity}
}

func (r *ChangeColumnType) Name() string { return "change-column-type" }

func (r *ChangeColumnType) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *ChangeColumnType) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	if !statement.ContainsSeq(keywordAlter, "TABLE") {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs(keywordAlter) {
			continue
		}

		j := i + 1

		// skip the outer ALTER TABLE
		if j < len(words) && words[j].WordIs("TABLE") {
			continue
		}

		if j < len(words) && words[j].WordIs("COLUMN") {
			j++
		}

		// column name
		j++
		if j >= len(words) {
			continue
		}

		// TYPE or SET DATA TYPE
		if words[j].WordIs("TYPE") {
			issue := issueAt(token, r.Name(), r.severity,
				"ALTER COLUMN TYPE")

			return &issue
		}

		if words[j].WordIs("SET") && j+2 < len(words) && words[j+1].WordIs("DATA") && words[j+2].WordIs("TYPE") {
			issue := issueAt(token, r.Name(), r.severity,
				"ALTER COLUMN SET DATA TYPE")

			return &issue
		}
	}

	return nil
}
