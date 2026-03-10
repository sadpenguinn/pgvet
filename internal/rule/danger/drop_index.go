package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// DropIndex detects DROP INDEX statements that omit CONCURRENTLY.
// Without CONCURRENTLY the operation acquires a lock that blocks writes.
type DropIndex struct {
	severity config.Severity
}

// NewDropIndex returns a DropIndex rule with the given severity.
func NewDropIndex(severity config.Severity) *DropIndex {
	return &DropIndex{severity: severity}
}

func (r *DropIndex) Name() string { return "drop-index-no-concurrently" }

func (r *DropIndex) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *DropIndex) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != "DROP" {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs("DROP") {
			continue
		}

		j := i + 1

		if j >= len(words) || !words[j].WordIs("INDEX") {
			continue
		}

		// the word after INDEX must be CONCURRENTLY
		j++
		if j < len(words) && words[j].WordIs("CONCURRENTLY") {
			return nil
		}

		issue := issueAt(token, r.Name(), r.severity,
			"DROP INDEX without CONCURRENTLY")

		return &issue
	}

	return nil
}
