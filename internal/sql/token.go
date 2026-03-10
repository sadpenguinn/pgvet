package sql

import "strings"

// TokenType identifies the lexical category of a token.
type TokenType uint8

const (
	TokEOF          TokenType = iota
	TokWord                   // unquoted identifier or keyword
	TokQuotedIdent            // "quoted identifier"
	TokString                 // 'string literal'
	TokDollarStr              // $$dollar quoted string$$
	TokNumber                 // numeric literal
	TokCommentLine            // -- single line comment
	TokCommentBlock           // /* block comment */
	TokWhitespace             // space or tab
	TokNewline                // \n or \r\n
	TokSemicolon              // ;
	TokLParen                 // (
	TokRParen                 // )
	TokComma                  // ,
	TokDot                    // .
	TokOther                  // any other character
)

// Token is a single lexical unit from a SQL source file.
type Token struct {
	Type  TokenType
	Value string
	Line  int // 1-based line number
	Col   int // 1-based column number
}

// WordIs reports whether the token is a specific word (case-insensitive).
func (t Token) WordIs(word string) bool {
	return t.Type == TokWord && strings.EqualFold(t.Value, word)
}

// IsComment reports whether the token is a comment of any kind.
func (t Token) IsComment() bool {
	return t.Type == TokCommentLine || t.Type == TokCommentBlock
}

// IsSkippable reports whether the token carries no semantic meaning
// (whitespace or comment).
func (t Token) IsSkippable() bool {
	return t.Type == TokWhitespace || t.Type == TokNewline || t.IsComment()
}
