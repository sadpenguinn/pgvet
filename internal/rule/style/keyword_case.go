package style

import (
	"strings"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// KeywordCase enforces consistent casing of SQL keywords.
type KeywordCase struct {
	severity  config.Severity
	styleCase config.Case
}

// NewKeywordCase returns a KeywordCase rule with the given severity and style.
func NewKeywordCase(severity config.Severity, styleCase config.Case) *KeywordCase {
	return &KeywordCase{severity: severity, styleCase: styleCase}
}

func (r *KeywordCase) Name() string { return "keyword-case" }

func (r *KeywordCase) Check(file *rule.File) []rule.Issue {
	issues := make([]rule.Issue, 0, len(file.Statements))

	for _, statement := range file.Statements {
		issues = append(issues, r.checkStatement(statement)...)
	}

	return issues
}

func (r *KeywordCase) checkStatement(statement *sql.Statement) []rule.Issue {
	var issues []rule.Issue

	tokens := statement.Tokens

	for i, token := range tokens {
		if token.Type != sql.TokWord {
			continue
		}

		upper := strings.ToUpper(token.Value)
		if !sqlKeywords[upper] {
			continue
		}

		// skip the word if it is in identifier position
		if isIdentifierPosition(tokens, i) {
			continue
		}

		if !r.matchesStyle(token.Value) {
			issues = append(issues, rule.Issue{
				Line:     token.Line,
				Col:      token.Col,
				Rule:     r.Name(),
				Severity: r.severity,
				Message:  r.message(token.Value),
			})
		}
	}

	return issues
}

// isIdentifierPosition reports whether the token at position i is used as an
// identifier rather than a keyword, based on the surrounding token context.
func isIdentifierPosition(tokens []sql.Token, i int) bool {
	prev := prevMeaningful(tokens, i)

	// after AS — always an alias: SELECT ... AS text
	if prev.WordIs("as") {
		return true
	}

	// after dot — qualified name: t.text
	if prev.Type == sql.TokDot {
		return true
	}

	next := nextMeaningful(tokens, i)

	// before dot — schema or table alias: public.text
	if next.Type == sql.TokDot {
		return true
	}

	// before comparison/assignment operator — column name: WHERE text = $1
	if isComparisonOp(next) {
		return true
	}

	// after ( or , — column name or alias, unless it is a constraint declarator
	// or another word that must remain a keyword in that position:
	//   CREATE TABLE t (text TEXT, CONSTRAINT fk FOREIGN KEY ...)
	//   CREATE INDEX ON t(text text_pattern_ops)
	upper := strings.ToUpper(tokens[i].Value)
	if (prev.Type == sql.TokLParen || prev.Type == sql.TokComma) && !alwaysKeywordAfterParen[upper] {
		return true
	}

	return false
}

// alwaysKeywordAfterParen is the set of words that remain keywords even as the
// first token after ( or ,: constraint declarators and expression starters.
var alwaysKeywordAfterParen = map[string]bool{
	"CONSTRAINT": true,
	"PRIMARY":    true,
	"UNIQUE":     true,
	"FOREIGN":    true,
	"CHECK":      true,
	"EXCLUDE":    true,
	"LIKE":       true,
	"NULL":       true,
	"NOT":        true,
	"CASE":       true,
	"CAST":       true,
	"DEFAULT":    true,
	"GENERATED":  true,
	"USING":      true,
}

// prevMeaningful returns the last non-whitespace, non-comment token before position i.
func prevMeaningful(tokens []sql.Token, i int) sql.Token {
	for j := i - 1; j >= 0; j-- {
		if !tokens[j].IsSkippable() {
			return tokens[j]
		}
	}

	return sql.Token{}
}

// nextMeaningful returns the first non-whitespace, non-comment token after position i.
func nextMeaningful(tokens []sql.Token, i int) sql.Token {
	for j := i + 1; j < len(tokens); j++ {
		if !tokens[j].IsSkippable() {
			return tokens[j]
		}
	}

	return sql.Token{}
}

// isComparisonOp reports whether the token is an operator or keyword that
// appears immediately after a column name in an expression.
func isComparisonOp(t sql.Token) bool {
	switch t.Value {
	case "=", "!=", "<>", "<", ">", "<=", ">=":
		return true
	}

	// keyword operators that always follow a column name
	switch strings.ToUpper(t.Value) {
	case "IN", "LIKE", "ILIKE", "IS", "BETWEEN":
		return true
	}

	return false
}

func (r *KeywordCase) matchesStyle(word string) bool {
	if r.styleCase == config.CaseLower {
		return word == strings.ToLower(word)
	}

	return word == strings.ToUpper(word)
}

func (r *KeywordCase) message(word string) string {
	if r.styleCase == config.CaseLower {
		return "keyword '" + word + "' must be lowercase: " + strings.ToLower(word)
	}

	return "keyword '" + word + "' must be uppercase: " + strings.ToUpper(word)
}

// sqlKeywords is the set of SQL keywords that the rule enforces casing on.
var sqlKeywords = func() map[string]bool {
	words := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "INTO", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER", "TABLE", "INDEX", "VIEW", "SCHEMA",
		"DATABASE", "SEQUENCE", "FUNCTION", "PROCEDURE", "TRIGGER", "TYPE",
		"EXTENSION", "COLUMN", "CONSTRAINT", "ADD", "SET", "NOT", "NULL",
		"DEFAULT", "PRIMARY", "FOREIGN", "KEY", "REFERENCES", "ON", "CASCADE",
		"RESTRICT", "UNIQUE", "CONCURRENTLY", "EXISTS", "IF", "VALID", "RENAME",
		"TO", "DATA", "TRUNCATE", "LOCK", "RETURNING", "WITH", "AS", "IN",
		"UNION", "ALL", "DISTINCT", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER",
		"CROSS", "FULL", "ORDER", "BY", "GROUP", "HAVING", "LIMIT", "OFFSET",
		"AND", "OR", "LIKE", "ILIKE", "IS", "BEGIN", "COMMIT", "ROLLBACK",
		"TRANSACTION", "SERIAL", "BIGSERIAL", "INTEGER", "BIGINT", "SMALLINT",
		"INT", "INT2", "INT4", "INT8", "FLOAT", "FLOAT4", "FLOAT8", "REAL",
		"DOUBLE", "PRECISION", "VARCHAR", "CHARACTER", "VARYING", "CHAR", "TEXT",
		"BOOLEAN", "BOOL", "TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME", "TIMETZ",
		"INTERVAL", "NUMERIC", "DECIMAL", "MONEY", "BYTEA", "JSONB", "JSON",
		"UUID", "OID", "ARRAY", "GRANT", "REVOKE", "INHERITS", "PARTITION",
		"VALUES", "CASE", "WHEN", "THEN", "ELSE", "END", "COALESCE", "NULLIF",
		"CAST", "BETWEEN", "REPLACE", "TEMP", "TEMPORARY", "UNLOGGED",
		"MATERIALIZED", "REFRESH", "ANALYZE", "VACUUM", "REINDEX", "EXPLAIN",
		"COPY", "USING", "ENABLE", "DISABLE", "NO", "ONLY", "INHERIT",
		"CURSOR", "DECLARE", "PREPARE", "EXECUTE", "DEALLOCATE", "SAVEPOINT",
		"RELEASE", "COMMENT", "WINDOW", "OVER", "ROWS", "RANGE", "PRECEDING",
		"FOLLOWING", "UNBOUNDED", "CURRENT", "ROW", "RECURSIVE", "LANGUAGE",
		"VOLATILE", "STABLE", "IMMUTABLE", "SECURITY", "DEFINER", "INVOKER",
		"NOTHING", "DO", "RULE", "SIMILAR", "LOGGED", "OF", "ABSOLUTE",
		"RELATIVE", "FORWARD", "BACKWARD", "MOVE", "FETCH", "CLOSE", "OPEN",
		"LISTEN", "NOTIFY", "UNLISTEN", "LOAD", "CHECKPOINT", "FILTER",
		"EXCLUDE", "INCLUDE", "NULLS", "FIRST", "LAST", "LATERAL", "NATURAL",
		"EXCEPT", "INTERSECT", "SYMMETRIC", "ASYMMETRIC", "SOME", "ANY",
		"GLOBAL", "LOCAL", "ACTION", "DEFERRABLE", "DEFERRED",
		"IMMEDIATE", "INITIALLY", "MATCH", "PARTIAL", "SIMPLE",
		"WITHOUT", "TIMEZONE", "OVERLAY", "PLACING", "FOR",
	}

	m := make(map[string]bool, len(words))

	for _, w := range words {
		m[w] = true
	}

	return m
}()
