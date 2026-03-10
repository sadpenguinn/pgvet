package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

const keywordAlter = "ALTER"

// AddColumnNotNull detects ADD COLUMN NOT NULL without a DEFAULT value.
// In PostgreSQL < 11 this triggers a full table rewrite.
// In PostgreSQL >= 11 a non-volatile DEFAULT avoids the rewrite, but
// the two-step migration pattern is still recommended for large tables.
type AddColumnNotNull struct {
	severity config.Severity
}

// NewAddColumnNotNull returns an AddColumnNotNull rule with the given severity.
func NewAddColumnNotNull(severity config.Severity) *AddColumnNotNull {
	return &AddColumnNotNull{severity: severity}
}

func (r *AddColumnNotNull) Name() string { return "add-column-not-null" }

func (r *AddColumnNotNull) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *AddColumnNotNull) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	if !statement.ContainsSeq(keywordAlter, "TABLE") {
		return nil
	}

	words := statement.Words()
	addIdx := -1

	for i, token := range words {
		if !token.WordIs("ADD") {
			continue
		}

		j := i + 1

		// skip optional COLUMN keyword
		if j < len(words) && words[j].WordIs("COLUMN") {
			j++
		}

		addIdx = j

		break
	}

	if addIdx < 0 || addIdx >= len(words) {
		return nil
	}

	// analyse column definition tokens that follow ADD [COLUMN]
	rest := words[addIdx:]
	hasNotNull := sql.SeqIndex(rest, "NOT", "NULL") >= 0
	hasDefault := sql.ContainsWord(rest, "DEFAULT")

	if !hasNotNull || hasDefault {
		return nil
	}

	firstTok, _ := statement.WordAt(0)
	issue := issueAt(firstTok, r.Name(), r.severity,
		"ADD COLUMN NOT NULL without DEFAULT")

	return &issue
}
