package sql

import "strings"

// Statement represents a single SQL statement.
type Statement struct {
	Tokens        []Token
	StartLine     int
	EndLine       int
	HasTerminator bool
}

// Text returns the full source text of the statement.
func (s *Statement) Text() string {
	var sb strings.Builder

	for _, tok := range s.Tokens {
		sb.WriteString(tok.Value)
	}

	return sb.String()
}

// Words returns only TokWord tokens, skipping whitespace and comments.
func (s *Statement) Words() []Token {
	return filterTokens(s.Tokens, func(t Token) bool { return t.Type == TokWord })
}

// Comments returns all comment tokens in the statement.
func (s *Statement) Comments() []Token {
	return filterTokens(s.Tokens, func(t Token) bool { return t.IsComment() })
}

// ContainsSeq reports whether the statement contains the given word sequence (case-insensitive).
func (s *Statement) ContainsSeq(words ...string) bool {
	return s.IndexSeq(words...) >= 0
}

// IndexSeq returns the index in Words() of the first occurrence of the word sequence.
// Returns -1 if not found.
func (s *Statement) IndexSeq(words ...string) int {
	return SeqIndex(s.Words(), words...)
}

// FirstWord returns the first word of the statement in upper case.
func (s *Statement) FirstWord() string {
	if ws := s.Words(); len(ws) > 0 {
		return strings.ToUpper(ws[0].Value)
	}

	return ""
}

// WordAt returns the word token at position i within Words().
func (s *Statement) WordAt(i int) (Token, bool) {
	ws := s.Words()

	if i < 0 || i >= len(ws) {
		return Token{}, false
	}

	return ws[i], true
}

// SeqIndex returns the index of the first occurrence of the word sequence in tokens.
// Returns -1 if not found.
func SeqIndex(tokens []Token, words ...string) int {
	for i := range tokens {
		if i+len(words) > len(tokens) {
			break
		}

		match := true

		for j, w := range words {
			if !tokens[i+j].WordIs(w) {
				match = false

				break
			}
		}

		if match {
			return i
		}
	}

	return -1
}

// ContainsWord reports whether any token in the slice matches word (case-insensitive).
func ContainsWord(tokens []Token, word string) bool {
	for _, tok := range tokens {
		if tok.WordIs(word) {
			return true
		}
	}

	return false
}

// SplitStatements splits a token slice into individual statements separated by semicolons.
func SplitStatements(tokens []Token) []*Statement {
	var stmts []*Statement

	var cur []Token

	for _, tok := range tokens {
		if tok.Type == TokEOF {
			break
		}

		cur = append(cur, tok)

		if tok.Type == TokSemicolon {
			if s := buildStatement(cur, true); s != nil {
				stmts = append(stmts, s)
			}

			cur = nil
		}
	}

	if s := buildStatement(cur, false); s != nil {
		stmts = append(stmts, s)
	}

	return stmts
}

func buildStatement(tokens []Token, hasTerminator bool) *Statement {
	for _, tok := range tokens {
		if !tok.IsSkippable() && tok.Type != TokSemicolon {
			return &Statement{tokens, tokens[0].Line, tokens[len(tokens)-1].Line, hasTerminator}
		}
	}

	return nil
}

func filterTokens(tokens []Token, f func(Token) bool) []Token {
	var out []Token

	for _, t := range tokens {
		if f(t) {
			out = append(out, t)
		}
	}

	return out
}
