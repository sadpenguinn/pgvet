package linter

import (
	"slices"
	"strings"

	"github.com/sadpenguinn/pgvet/internal/sql"
)

// directive represents a single sqlint enable/disable instruction.
type directive struct {
	line    int
	rules   []string // empty means all rules
	disable bool     // true = disable, false = enable
}

// parseDirectives extracts sqlint directives from the token stream.
//
// Supported formats:
//
//	-- sqlint:disable rule1,rule2
//	-- sqlint:enable  rule1,rule2
//	-- sqlint:disable            (all rules)
//	-- sqlint:enable             (all rules)
func parseDirectives(tokens []sql.Token) []directive {
	var directives []directive

	for _, tok := range tokens {
		if tok.Type != sql.TokCommentLine {
			continue
		}

		d, ok := parseDirective(tok)
		if ok {
			directives = append(directives, d)
		}
	}

	return directives
}

// parseDirective parses a single sqlint directive from a comment token.
func parseDirective(tok sql.Token) (directive, bool) {
	body := strings.TrimSpace(strings.TrimPrefix(tok.Value, "--"))

	var (
		disable bool
		rest    string
	)

	switch {
	case strings.HasPrefix(body, "sqlint:disable"):
		disable = true
		rest = strings.TrimPrefix(body, "sqlint:disable")
	case strings.HasPrefix(body, "sqlint:enable"):
		disable = false
		rest = strings.TrimPrefix(body, "sqlint:enable")
	default:
		return directive{}, false
	}

	var rules []string

	rest = strings.TrimSpace(rest)
	if rest != "" {
		for r := range strings.SplitSeq(rest, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				rules = append(rules, r)
			}
		}
	}

	return directive{
		line:    tok.Line,
		rules:   rules,
		disable: disable,
	}, true
}

// isDisabled reports whether ruleName is suppressed at the given line.
// Directives must be ordered by line number (as they appear in the file).
// A directive on the same line as an issue also applies to that issue,
// which makes inline suppression work correctly.
func isDisabled(directives []directive, ruleName string, line int) bool {
	state := false

	for _, d := range directives {
		if d.line > line {
			break
		}

		if matchesDirective(d, ruleName) {
			state = d.disable
		}
	}

	return state
}

// matchesDirective reports whether the directive applies to the given rule.
func matchesDirective(d directive, ruleName string) bool {
	if len(d.rules) == 0 {
		return true // applies to all rules
	}

	return slices.Contains(d.rules, ruleName)
}
