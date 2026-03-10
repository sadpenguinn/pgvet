package linter

import (
	"slices"
	"strings"

	"github.com/sadpenguinn/pgvet/internal/sql"
)

// directive представляет инструкцию подавления nolint на конкретной строке.
type directive struct {
	line  int
	rules []string // пустой — все правила
}

// parseDirectives извлекает nolint-директивы из потока токенов.
//
// Поддерживаемые форматы:
//
//	-- nolint:rule1,rule2
//	-- nolint              (все правила)
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

// parseDirective разбирает одну nolint-директиву из токена комментария.
func parseDirective(tok sql.Token) (directive, bool) {
	body := strings.TrimSpace(strings.TrimPrefix(tok.Value, "--"))

	if !strings.HasPrefix(body, "nolint") {
		return directive{}, false
	}

	rest := strings.TrimPrefix(body, "nolint")

	var rules []string

	if rest, ok := strings.CutPrefix(rest, ":"); ok {
		for r := range strings.SplitSeq(rest, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				rules = append(rules, r)
			}
		}
	}

	return directive{
		line:  tok.Line,
		rules: rules,
	}, true
}

// isDisabled сообщает, подавлено ли ruleName на данной строке.
// Директива применяется только к строке, на которой она находится.
func isDisabled(directives []directive, ruleName string, line int) bool {
	for _, d := range directives {
		if d.line == line && matchesDirective(d, ruleName) {
			return true
		}
	}

	return false
}

// matchesDirective сообщает, применяется ли директива к данному правилу.
func matchesDirective(d directive, ruleName string) bool {
	if len(d.rules) == 0 {
		return true
	}

	return slices.Contains(d.rules, ruleName)
}
