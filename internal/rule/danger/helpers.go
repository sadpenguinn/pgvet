// Package dangerous contains rules that detect DDL operations capable of
// locking the database or causing data loss.
package dangerous

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// wordsAfterSeq returns the word tokens following the first occurrence of words.
func wordsAfterSeq(statement *sql.Statement, words ...string) []sql.Token {
	if idx := statement.IndexSeq(words...); idx >= 0 {
		return statement.Words()[idx+len(words):]
	}

	return nil
}

// issueAt builds an Issue positioned at the given token.
func issueAt(token sql.Token, ruleName string, severity config.Severity, msg string) rule.Issue {
	return rule.Issue{
		Line:     token.Line,
		Col:      token.Col,
		Rule:     ruleName,
		Severity: severity,
		Message:  msg,
	}
}
