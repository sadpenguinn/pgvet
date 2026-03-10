package sql

import "unicode"

// Tokenize splits SQL source text into a flat slice of tokens.
func Tokenize(src string) []Token {
	t := &tokenizer{src: []rune(src), line: 1, col: 1}

	var tokens []Token

	for {
		tok := t.next()
		tokens = append(tokens, tok)

		if tok.Type == TokEOF {
			break
		}
	}

	return tokens
}

type tokenizer struct {
	src  []rune
	pos  int
	line int
	col  int
}

var singleChar = map[rune]TokenType{
	';': TokSemicolon,
	'(': TokLParen,
	')': TokRParen,
	',': TokComma,
	'.': TokDot,
}

func (t *tokenizer) next() Token {
	if t.pos >= len(t.src) {
		return Token{TokEOF, "", t.line, t.col}
	}

	line, col, ch := t.line, t.col, t.src[t.pos]

	if tt, ok := singleChar[ch]; ok {
		t.advance()

		return Token{tt, string(ch), line, col}
	}

	return t.nextComplex(line, col, ch)
}

func (t *tokenizer) nextComplex(line, col int, ch rune) Token {
	switch {
	case ch == '-' && t.peek() == '-':
		return t.readLineComment(line, col)
	case ch == '/' && t.peek() == '*':
		return t.readBlockComment(line, col)
	case ch == '\'':
		return t.readQuoted(TokString, '\'', line, col)
	case ch == '"':
		return t.readQuoted(TokQuotedIdent, '"', line, col)
	case ch == '$' && (t.peek() == '$' || isWordStart(t.peek())):
		return t.readDollarStr(line, col)
	case ch == '\n' || ch == '\r':
		return t.readNewline(line, col)
	case ch == ' ' || ch == '\t':
		return Token{TokWhitespace, t.readWhile(func(r rune) bool { return r == ' ' || r == '\t' }), line, col}
	case isWordStart(ch):
		return Token{TokWord, t.readWhile(isWordPart), line, col}
	case isDigit(ch):
		return Token{TokNumber, t.readWhile(func(r rune) bool { return isDigit(r) || r == '.' }), line, col}
	default:
		t.advance()

		return Token{TokOther, string(ch), line, col}
	}
}

func (t *tokenizer) advance() {
	if t.pos >= len(t.src) {
		return
	}

	ch := t.src[t.pos]
	t.pos++

	if ch == '\n' {
		t.line++
		t.col = 1
	} else {
		t.col++
	}
}

func (t *tokenizer) peek() rune {
	if pos := t.pos + 1; pos < len(t.src) {
		return t.src[pos]
	}

	return 0
}

// readWhile advances while cond holds and returns the consumed text.
// Must only be used for characters that are never newlines.
func (t *tokenizer) readWhile(cond func(rune) bool) string {
	start := t.pos

	for t.pos < len(t.src) && cond(t.src[t.pos]) {
		t.pos++
		t.col++
	}

	return string(t.src[start:t.pos])
}

func (t *tokenizer) readNewline(line, col int) Token {
	if t.src[t.pos] == '\r' {
		t.pos++

		if t.pos < len(t.src) && t.src[t.pos] == '\n' {
			t.pos++
		}
	} else {
		t.pos++
	}

	t.line++
	t.col = 1

	return Token{TokNewline, "\n", line, col}
}

func (t *tokenizer) readLineComment(line, col int) Token {
	start := t.pos
	t.pos += 2
	t.col += 2

	for t.pos < len(t.src) && t.src[t.pos] != '\n' && t.src[t.pos] != '\r' {
		t.pos++
		t.col++
	}

	return Token{TokCommentLine, string(t.src[start:t.pos]), line, col}
}

func (t *tokenizer) readBlockComment(line, col int) Token {
	start := t.pos
	t.advance()
	t.advance() // /*

	for t.pos < len(t.src) {
		if t.src[t.pos] == '*' && t.peek() == '/' {
			t.advance()
			t.advance()

			break
		}

		t.advance()
	}

	return Token{TokCommentBlock, string(t.src[start:t.pos]), line, col}
}

// readQuoted reads a single-quoted string or double-quoted identifier.
// Handles escaped quotes: repeated quote character inside string ends quoting.
func (t *tokenizer) readQuoted(tt TokenType, quote rune, line, col int) Token {
	start := t.pos
	t.advance() // opening quote

	for t.pos < len(t.src) {
		ch := t.src[t.pos]
		t.advance()

		if ch == quote {
			if t.pos < len(t.src) && t.src[t.pos] == quote {
				t.advance() // escaped quote

				continue
			}

			break
		}
	}

	return Token{tt, string(t.src[start:t.pos]), line, col}
}

func (t *tokenizer) readDollarStr(line, col int) Token {
	start := t.pos
	t.advance() // $

	tagStart := t.pos

	for t.pos < len(t.src) && t.src[t.pos] != '$' {
		t.advance()
	}

	closing := []rune("$" + string(t.src[tagStart:t.pos]) + "$")

	if t.pos < len(t.src) {
		t.advance() // closing $ of opening tag
	}

	for t.pos < len(t.src) {
		if t.src[t.pos] == '$' && runesMatch(t.src[t.pos:], closing) {
			for range closing {
				t.advance()
			}

			break
		}

		t.advance()
	}

	return Token{TokDollarStr, string(t.src[start:t.pos]), line, col}
}

func runesMatch(src, pat []rune) bool {
	if len(src) < len(pat) {
		return false
	}

	for i, r := range pat {
		if src[i] != r {
			return false
		}
	}

	return true
}

func isWordStart(ch rune) bool { return unicode.IsLetter(ch) || ch == '_' }
func isWordPart(ch rune) bool  { return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' }
func isDigit(ch rune) bool     { return ch >= '0' && ch <= '9' }
