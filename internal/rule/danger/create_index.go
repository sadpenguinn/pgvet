package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// CreateIndex detects CREATE INDEX statements that omit CONCURRENTLY.
// Without CONCURRENTLY the index build holds a lock that blocks all writes
// for the duration of the operation.
type CreateIndex struct {
	severity config.Severity
}

// NewCreateIndex returns a CreateIndex rule with the given severity.
func NewCreateIndex(severity config.Severity) *CreateIndex {
	return &CreateIndex{severity: severity}
}

func (r *CreateIndex) Name() string { return "create-index-no-concurrently" }

func (r *CreateIndex) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *CreateIndex) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != "CREATE" {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs("CREATE") {
			continue
		}

		j := i + 1

		// skip optional UNIQUE
		if j < len(words) && words[j].WordIs("UNIQUE") {
			j++
		}

		if j >= len(words) || !words[j].WordIs("INDEX") {
			continue
		}

		// the word after INDEX must be CONCURRENTLY
		j++
		if j < len(words) && words[j].WordIs("CONCURRENTLY") {
			return nil
		}

		issue := issueAt(token, r.Name(), r.severity,
			"CREATE INDEX without CONCURRENTLY")

		return &issue
	}

	return nil
}
