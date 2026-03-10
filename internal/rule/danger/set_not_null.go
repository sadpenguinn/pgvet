package dangerous

import (
	"strings"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// SetNotNull detects ALTER COLUMN SET NOT NULL.
// This operation requires a full sequential scan of the table to validate
// existing rows. Consider using a CHECK constraint with NOT VALID followed
// by VALIDATE CONSTRAINT instead.
//
// Exception: the pattern add nullable → fill via UPDATE → set not null is
// considered safe because all rows are guaranteed to be filled before the
// nullable constraint is removed.
type SetNotNull struct {
	severity config.Severity
}

// NewSetNotNull returns a SetNotNull rule with the given severity.
func NewSetNotNull(severity config.Severity) *SetNotNull {
	return &SetNotNull{severity: severity}
}

func (r *SetNotNull) Name() string { return "set-not-null" }

func (r *SetNotNull) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for i, statement := range file.Statements {
		if issue := r.checkStatement(statement, file.Statements[:i]); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *SetNotNull) checkStatement(statement *sql.Statement, preceding []*sql.Statement) *rule.Issue {
	if statement.FirstWord() != keywordAlter {
		return nil
	}

	if !statement.ContainsSeq(keywordAlter, "TABLE") {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if !token.WordIs("SET") {
			continue
		}

		if i+2 < len(words) && words[i+1].WordIs("NOT") && words[i+2].WordIs("NULL") {
			table, col := alterColumnTarget(statement)
			if table != "" && col != "" && isSafeSetNotNull(table, col, preceding) {
				return nil
			}

			issue := issueAt(token, r.Name(), r.severity, "SET NOT NULL")

			return &issue
		}
	}

	return nil
}

// isSafeSetNotNull reports whether SET NOT NULL is preceded by:
// 1. ADD COLUMN col (without NOT NULL or DEFAULT) for the same table.
// 2. UPDATE table SET col = ... for the same table.
func isSafeSetNotNull(table, col string, preceding []*sql.Statement) bool {
	return hasAddColumnNullable(table, col, preceding) &&
		hasUpdateFill(table, col, preceding)
}

// hasAddColumnNullable reports whether stmts contain ADD COLUMN col without NOT NULL or DEFAULT.
func hasAddColumnNullable(table, col string, stmts []*sql.Statement) bool {
	for _, s := range stmts {
		if !s.ContainsSeq(keywordAlter, "TABLE") {
			continue
		}

		if !s.ContainsSeq("ADD", "COLUMN") {
			continue
		}

		t, c := addColumnTarget(s)
		if t != table || c != col {
			continue
		}

		// Column must not already have NOT NULL or DEFAULT — otherwise no fill is needed
		if s.ContainsSeq("NOT", "NULL") || sql.ContainsWord(s.Words(), "DEFAULT") {
			continue
		}

		return true
	}

	return false
}

// hasUpdateFill reports whether stmts contain UPDATE table SET col = ...
func hasUpdateFill(table, col string, stmts []*sql.Statement) bool {
	for _, s := range stmts {
		if s.FirstWord() != "UPDATE" {
			continue
		}

		words := s.Words()
		if len(words) < 2 {
			continue
		}

		if strings.ToLower(words[1].Value) != table {
			continue
		}

		// look for the target column after SET
		after := wordsAfterSeq(s, "SET")

		for _, w := range after {
			if strings.ToLower(w.Value) == col {
				return true
			}
		}
	}

	return false
}

// alterColumnTarget extracts (table, column) from ALTER TABLE t ALTER COLUMN col SET NOT NULL.
func alterColumnTarget(s *sql.Statement) (string, string) {
	after := wordsAfterSeq(s, "TABLE")
	if len(after) == 0 {
		return "", ""
	}

	table := strings.ToLower(after[0].Value)

	afterCol := wordsAfterSeq(s, "COLUMN")
	if len(afterCol) == 0 {
		return table, ""
	}

	col := strings.ToLower(afterCol[0].Value)

	return table, col
}

// addColumnTarget extracts (table, column) from ALTER TABLE t ADD COLUMN col TYPE.
func addColumnTarget(s *sql.Statement) (string, string) {
	after := wordsAfterSeq(s, "TABLE")
	if len(after) == 0 {
		return "", ""
	}

	table := strings.ToLower(after[0].Value)

	afterCol := wordsAfterSeq(s, "COLUMN")
	if len(afterCol) == 0 {
		return table, ""
	}

	col := strings.ToLower(afterCol[0].Value)

	return table, col
}
