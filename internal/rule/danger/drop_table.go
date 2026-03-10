package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// DropTable detects DROP TABLE statements.
// Dropping a table is irreversible and permanently destroys all of its data.
type DropTable struct {
	severity config.Severity
}

// NewDropTable returns a DropTable rule with the given severity.
func NewDropTable(severity config.Severity) *DropTable {
	return &DropTable{severity: severity}
}

func (r *DropTable) Name() string { return "drop-table" }

func (r *DropTable) Check(file *rule.File) []rule.Issue {
	var issues []rule.Issue

	for _, statement := range file.Statements {
		if issue := r.checkStatement(statement); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (r *DropTable) checkStatement(statement *sql.Statement) *rule.Issue {
	if statement.FirstWord() != "DROP" {
		return nil
	}

	words := statement.Words()

	for i, token := range words {
		if token.WordIs("DROP") && i+1 < len(words) && words[i+1].WordIs("TABLE") {
			issue := issueAt(token, r.Name(), r.severity,
				"DROP TABLE")

			return &issue
		}
	}

	return nil
}
