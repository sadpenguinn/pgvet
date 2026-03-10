package dangerous

import (
	"fmt"
	"strings"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// RedundantIndex detects redundant CREATE INDEX statements within a single file:
//   - exact duplicates (same columns, same method)
//   - index on (a) when index on (a, b) exists — btree prefix rule
//   - non-unique index when UNIQUE on the same columns exists
//   - partial index (WHERE) covered by a full index on the same columns
//   - index with INCLUDE covered by an index with a superset INCLUDE on the same keys
type RedundantIndex struct {
	severity config.Severity
}

// NewRedundantIndex returns a RedundantIndex rule with the given severity.
func NewRedundantIndex(severity config.Severity) *RedundantIndex {
	return &RedundantIndex{severity: severity}
}

func (r *RedundantIndex) Name() string { return "redundant-index" }

func (r *RedundantIndex) Check(file *rule.File) []rule.Issue {
	var defs []*indexDef

	for _, stmt := range file.Statements {
		if def := parseIndexDef(stmt); def != nil {
			defs = append(defs, def)
		}
	}

	var issues []rule.Issue

	for i, candidate := range defs {
		for j, superseder := range defs {
			if i == j {
				continue
			}

			if !isRedundantBy(candidate, superseder) {
				continue
			}

			// for exact duplicates flag only the later occurrence
			exactDup := len(candidate.keyCols) == len(superseder.keyCols) && candidate.unique == superseder.unique
			if exactDup &&
				(superseder.line > candidate.line ||
					(superseder.line == candidate.line && superseder.col >= candidate.col)) {
				continue
			}

			issues = append(issues, rule.Issue{
				Line:     candidate.line,
				Col:      candidate.col,
				Rule:     r.Name(),
				Severity: r.severity,
				Message:  buildRedundancyMessage(candidate, superseder),
			})

			break
		}
	}

	return issues
}

// indexDef describes a single CREATE INDEX within a file.
type indexDef struct {
	schema    string
	table     string
	method    string // "btree" by default
	unique    bool
	keyCols   []indexCol
	inclCols  map[string]bool // INCLUDE columns (unordered set)
	partial   bool            // has a WHERE predicate
	whereToks []sql.Token     // normalized WHERE tokens
	line      int
	col       int
}

// indexCol describes a single column in the key part of an index.
type indexCol struct {
	name       string // lowercase; empty for expressions
	isExpr     bool
	expr       []sql.Token // normalized expression tokens
	desc       bool        // true = DESC (default ASC)
	nullsFirst bool        // resolved value taking defaults into account
}

// isRedundantBy reports whether candidate is covered by superseder.
func isRedundantBy(candidate, superseder *indexDef) bool {
	if !sameTable(candidate, superseder) {
		return false
	}

	if candidate.method != superseder.method {
		return false
	}

	// a UNIQUE index is not covered by a non-unique one
	if candidate.unique && !superseder.unique {
		return false
	}

	// superseder is partial but candidate is not: does not cover
	if superseder.partial && !candidate.partial {
		return false
	}

	// both partial: WHERE predicates must match
	if candidate.partial && superseder.partial {
		if !tokenSliceEqual(candidate.whereToks, superseder.whereToks) {
			return false
		}
	}

	nC, nS := len(candidate.keyCols), len(superseder.keyCols)

	if !keyColsCompatible(candidate, nC, nS) {
		return false
	}

	for i := range nC {
		if !colsEqual(candidate.keyCols[i], superseder.keyCols[i]) {
			return false
		}
	}

	// INCLUDE of superseder must be a superset of INCLUDE of candidate
	for col := range candidate.inclCols {
		if !superseder.inclCols[col] {
			return false
		}
	}

	return true
}

// keyColsCompatible reports whether the key column counts are compatible for redundancy.
func keyColsCompatible(candidate *indexDef, nC, nS int) bool {
	if candidate.unique {
		// UNIQUE is only covered by an exact UNIQUE duplicate
		return nC == nS
	}

	// non-unique: for btree a prefix is enough; for others an exact match is required
	if candidate.method == "btree" {
		return nS >= nC
	}

	return nS == nC
}

func colsEqual(a, b indexCol) bool {
	if a.isExpr != b.isExpr {
		return false
	}

	if a.desc != b.desc {
		return false
	}

	if a.nullsFirst != b.nullsFirst {
		return false
	}

	if a.isExpr {
		return tokenSliceEqual(a.expr, b.expr)
	}

	return a.name == b.name
}

func sameTable(a, b *indexDef) bool {
	// if both have a schema — compare fully
	// if one has a schema and the other does not — skip (different search_path)
	if a.schema != b.schema {
		return false
	}

	return a.table == b.table
}

func buildRedundancyMessage(candidate, superseder *indexDef) string {
	candStr := "(" + strings.Join(colDisplayNames(candidate.keyCols), ", ") + ")"
	supStr := "(" + strings.Join(colDisplayNames(superseder.keyCols), ", ") + ")"

	exactDup := len(candidate.keyCols) == len(superseder.keyCols) && candidate.unique == superseder.unique
	if exactDup {
		if candidate.unique {
			return "duplicate unique index on " + candStr
		}

		return "duplicate index on " + candStr
	}

	superLabel := "index"
	if superseder.unique {
		superLabel = "unique index"
	}

	return fmt.Sprintf("index on %s is redundant: covered by %s on %s", candStr, superLabel, supStr)
}

func colDisplayNames(cols []indexCol) []string {
	names := make([]string, len(cols))

	for i, c := range cols {
		if c.isExpr {
			names[i] = "<expr>"
		} else {
			names[i] = c.name
		}
	}

	return names
}

// ─── parsing ──────────────────────────────────────────────────────────────────

func parseIndexDef(stmt *sql.Statement) *indexDef {
	if stmt.FirstWord() != "CREATE" {
		return nil
	}

	words := stmt.Words()

	// find INDEX and check for UNIQUE
	unique := false
	indexWordPos := -1

	for i, w := range words {
		if w.WordIs("UNIQUE") {
			unique = true
		}

		if w.WordIs("INDEX") {
			indexWordPos = i

			break
		}
	}

	if indexWordPos < 0 {
		return nil
	}

	// find ON
	onPos := -1

	for i := indexWordPos + 1; i < len(words); i++ {
		if words[i].WordIs("ON") {
			onPos = i

			break
		}
	}

	if onPos < 0 || onPos+1 >= len(words) {
		return nil
	}

	// table name
	schema, table := parseTableRef(stmt.Tokens, words[onPos+1])

	// access method
	method := "btree"

	for i := onPos + 1; i < len(words); i++ {
		if words[i].WordIs("USING") && i+1 < len(words) {
			method = strings.ToLower(words[i+1].Value)

			break
		}
	}

	// position of the last token of the table name in the raw token stream
	tableEndPos := findTableEndPos(stmt.Tokens, words[onPos+1])
	if tableEndPos < 0 {
		return nil
	}

	// key column block — first () after the table name
	keyOpen, keyClose := findParenBlock(stmt.Tokens, tableEndPos+1)
	if keyOpen < 0 {
		return nil
	}

	// parse key columns
	entries := splitAtTopLevelCommas(stmt.Tokens[keyOpen+1 : keyClose])
	keyCols := make([]indexCol, 0, len(entries))

	for _, entry := range entries {
		col, ok := parseIndexCol(entry)
		if ok {
			keyCols = append(keyCols, col)
		}
	}

	if len(keyCols) == 0 {
		return nil
	}

	// INCLUDE columns
	inclCols := parseIncludeCols(stmt.Tokens, keyClose+1)

	// partial index (WHERE)
	afterInclude := keyClose + 1

	if inclCols != nil {
		// look for WHERE after INCLUDE (...)
		_, inclClose := findParenBlock(stmt.Tokens, keyClose+1)
		if inclClose >= 0 {
			afterInclude = inclClose + 1
		}
	}

	whereToks := extractWhereToks(stmt.Tokens, afterInclude)
	partial := len(whereToks) > 0

	createTok := words[0]

	return &indexDef{
		schema:    schema,
		table:     table,
		method:    method,
		unique:    unique,
		keyCols:   keyCols,
		inclCols:  inclCols,
		partial:   partial,
		whereToks: whereToks,
		line:      createTok.Line,
		col:       createTok.Col,
	}
}

// parseTableRef extracts (schema, table) from the token immediately after ON.
func parseTableRef(tokens []sql.Token, firstWord sql.Token) (string, string) {
	pos := tokenPos(tokens, firstWord)
	if pos < 0 {
		return "", normalizeIdent(firstWord)
	}

	name := normalizeIdent(tokens[pos])
	j := skipWS(tokens, pos+1)

	if j < len(tokens) && tokens[j].Type == sql.TokDot {
		j = skipWS(tokens, j+1)

		if j < len(tokens) && isIdentToken(tokens[j]) {
			return name, normalizeIdent(tokens[j])
		}
	}

	return "", name
}

// findTableEndPos returns the position of the last token of the table name.
func findTableEndPos(tokens []sql.Token, firstWord sql.Token) int {
	pos := tokenPos(tokens, firstWord)
	if pos < 0 {
		return -1
	}

	j := skipWS(tokens, pos+1)

	if j < len(tokens) && tokens[j].Type == sql.TokDot {
		j = skipWS(tokens, j+1)

		if j < len(tokens) && isIdentToken(tokens[j]) {
			return j
		}
	}

	return pos
}

// findParenBlock finds the first ( starting from startAfter and returns positions of ( and ).
func findParenBlock(tokens []sql.Token, startAfter int) (int, int) {
	for i := startAfter; i < len(tokens); i++ {
		if tokens[i].Type != sql.TokLParen {
			continue
		}

		depth := 0

		for j := i; j < len(tokens); j++ {
			switch tokens[j].Type { //nolint:exhaustive
			case sql.TokLParen:
				depth++
			case sql.TokRParen:
				depth--

				if depth == 0 {
					return i, j
				}
			}
		}
	}

	return -1, -1
}

// splitAtTopLevelCommas splits a token slice at commas at depth 0.
func splitAtTopLevelCommas(tokens []sql.Token) [][]sql.Token {
	var entries [][]sql.Token

	var cur []sql.Token

	depth := 0

	for _, t := range tokens {
		switch t.Type { //nolint:exhaustive
		case sql.TokLParen:
			depth++

			cur = append(cur, t)
		case sql.TokRParen:
			depth--

			cur = append(cur, t)
		case sql.TokComma:
			if depth == 0 {
				entries = append(entries, cur)
				cur = nil
			} else {
				cur = append(cur, t)
			}
		default:
			cur = append(cur, t)
		}
	}

	if len(cur) > 0 {
		entries = append(entries, cur)
	}

	return entries
}

// parseIndexCol parses one element of a column list: expr [ASC|DESC] [NULLS FIRST|LAST].
func parseIndexCol(entry []sql.Token) (indexCol, bool) {
	// positions of TokWord tokens in the slice
	var wordIdxs []int

	for i, t := range entry {
		if t.Type == sql.TokWord {
			wordIdxs = append(wordIdxs, i)
		}
	}

	exprEnd := len(entry) // exclusive end of expression (before modifiers)
	desc := false
	nullsFirst := false
	nullsExplicit := false
	n := len(wordIdxs)

	// NULLS FIRST/LAST at the end
	if n >= 2 {
		last := entry[wordIdxs[n-1]]
		prev := entry[wordIdxs[n-2]]

		if prev.WordIs("NULLS") && (last.WordIs("FIRST") || last.WordIs("LAST")) {
			nullsExplicit = true
			nullsFirst = last.WordIs("FIRST")
			exprEnd = wordIdxs[n-2]
			n -= 2
			wordIdxs = wordIdxs[:n]
		}
	}

	// ASC / DESC direction
	if n >= 1 {
		last := entry[wordIdxs[n-1]]

		if last.WordIs("ASC") {
			exprEnd = wordIdxs[n-1]
			n--
			wordIdxs = wordIdxs[:n]
		} else if last.WordIs("DESC") {
			desc = true
			exprEnd = wordIdxs[n-1]
			n--
			wordIdxs = wordIdxs[:n]
		}
	}

	_ = wordIdxs // n and wordIdxs are used only for modifier detection above

	// resolve NULLS default: ASC → NULLS LAST, DESC → NULLS FIRST
	if !nullsExplicit {
		nullsFirst = desc
	}

	// trim tokens to exprEnd, strip surrounding whitespace
	expr := trimWS(entry[:exprEnd])
	if len(expr) == 0 {
		return indexCol{}, false
	}

	// simple identifier?
	if len(expr) == 1 && isIdentToken(expr[0]) {
		return indexCol{
			name:       normalizeIdent(expr[0]),
			desc:       desc,
			nullsFirst: nullsFirst,
		}, true
	}

	// expression: normalize tokens
	return indexCol{
		isExpr:     true,
		expr:       normalizeExprTokens(expr),
		desc:       desc,
		nullsFirst: nullsFirst,
	}, true
}

// parseIncludeCols extracts INCLUDE columns as a set of lowercase names.
func parseIncludeCols(tokens []sql.Token, startAfter int) map[string]bool {
	for i := startAfter; i < len(tokens); i++ {
		if tokens[i].Type != sql.TokWord || !strings.EqualFold(tokens[i].Value, "INCLUDE") {
			continue
		}

		openParen, closeParen := findParenBlock(tokens, i+1)
		if openParen < 0 {
			return nil
		}

		cols := make(map[string]bool)
		depth := 0

		for j := openParen; j <= closeParen; j++ {
			switch tokens[j].Type { //nolint:exhaustive
			case sql.TokLParen:
				depth++
			case sql.TokRParen:
				depth--
			case sql.TokWord:
				if depth == 1 {
					cols[strings.ToLower(tokens[j].Value)] = true
				}
			case sql.TokQuotedIdent:
				if depth == 1 {
					cols[strings.Trim(tokens[j].Value, "\"")] = true
				}
			}
		}

		return cols
	}

	return nil
}

// extractWhereToks returns normalized tokens of the WHERE predicate.
func extractWhereToks(tokens []sql.Token, startAfter int) []sql.Token {
	wherePos := -1

	for i := startAfter; i < len(tokens); i++ {
		if tokens[i].Type == sql.TokWord && strings.EqualFold(tokens[i].Value, "WHERE") {
			wherePos = i

			break
		}
	}

	if wherePos < 0 {
		return nil
	}

	var result []sql.Token

	for i := wherePos + 1; i < len(tokens); i++ {
		t := tokens[i]

		if t.Type == sql.TokEOF || t.Type == sql.TokSemicolon {
			break
		}

		if t.Type == sql.TokWhitespace || t.Type == sql.TokNewline ||
			t.Type == sql.TokCommentLine || t.Type == sql.TokCommentBlock {
			continue
		}

		if t.Type == sql.TokWord {
			t.Value = strings.ToLower(t.Value)
		}

		result = append(result, t)
	}

	return result
}

// normalizeExprTokens strips whitespace and comments, lowercases TokWord values.
func normalizeExprTokens(tokens []sql.Token) []sql.Token {
	result := make([]sql.Token, 0, len(tokens))

	for _, t := range tokens {
		if t.Type == sql.TokWhitespace || t.Type == sql.TokNewline ||
			t.Type == sql.TokCommentLine || t.Type == sql.TokCommentBlock {
			continue
		}

		if t.Type == sql.TokWord {
			t.Value = strings.ToLower(t.Value)
		}

		result = append(result, t)
	}

	return result
}

func tokenSliceEqual(a, b []sql.Token) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Type != b[i].Type || a[i].Value != b[i].Value {
			return false
		}
	}

	return true
}

// ─── utilities ────────────────────────────────────────────────────────────────

func tokenPos(tokens []sql.Token, tok sql.Token) int {
	for i, t := range tokens {
		if t.Line == tok.Line && t.Col == tok.Col {
			return i
		}
	}

	return -1
}

func skipWS(tokens []sql.Token, from int) int {
	for from < len(tokens) && (tokens[from].Type == sql.TokWhitespace || tokens[from].Type == sql.TokNewline) {
		from++
	}

	return from
}

func isIdentToken(t sql.Token) bool {
	return t.Type == sql.TokWord || t.Type == sql.TokQuotedIdent
}

func normalizeIdent(t sql.Token) string {
	if t.Type == sql.TokQuotedIdent {
		return strings.Trim(t.Value, "\"")
	}

	return strings.ToLower(t.Value)
}

func trimWS(tokens []sql.Token) []sql.Token {
	start := 0

	for start < len(tokens) && (tokens[start].Type == sql.TokWhitespace || tokens[start].Type == sql.TokNewline) {
		start++
	}

	end := len(tokens)

	for end > start && (tokens[end-1].Type == sql.TokWhitespace || tokens[end-1].Type == sql.TokNewline) {
		end--
	}

	return tokens[start:end]
}
