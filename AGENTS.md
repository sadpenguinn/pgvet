# AGENTS.md

## Purpose

`pgvet` is a static analyzer for PostgreSQL SQL migration files. It detects dangerous DDL operations that cause table locks or data loss, and enforces SQL style consistency. Designed to integrate naturally into Go developer workflows.

Binary: `pgvet`. Module: `github.com/sadpenguinn/pgvet`. Go 1.26.

---

## Architecture

```
cmd/pgvet/main.go          → CLI entry point, flag parsing, exit codes
internal/config/           → YAML config loading, severity model
internal/linter/           → Orchestration: file discovery, concurrency, directive parsing
internal/rule/             → Rule interface, Issue/File types
internal/rule/danger/      → 12 danger rules (DDL lock/data-loss detection)
internal/rule/style/       → 4 style rules (formatting, consistency)
internal/sql/              → Custom SQL tokenizer (no external parser)
internal/report/           → Text/JSON formatters, ANSI colors, explain mode
test/                      → Golden file tests + test data
```

---

## Key Files

| File | Description |
|------|-------------|
| `cmd/pgvet/main.go` | CLI flags: `-config`, `-dir`, `-file`, `-format`, `-explain`, `-no-exit-code`, `-version` |
| `internal/config/config.go` | YAML config, severity levels (error/warning/info), rule enable/disable |
| `internal/linter/linter.go` | Core engine: discovers `.sql` files, tokenizes, runs rules, collects issues |
| `internal/linter/registry.go` | Factory: builds active rule set from config |
| `internal/linter/disable.go` | Parses `-- sqlint:disable rule1,rule2` / `sqlint:enable` directives |
| `internal/rule/rule.go` | `Rule` interface (`Name() string`, `Check(*File) []Issue`), `Issue`, `File` types |
| `internal/rule/danger/` | One file per danger rule + `helpers.go` for shared pattern matching |
| `internal/rule/style/` | One file per style rule |
| `internal/sql/tokenizer.go` | Lexer: handles dollar-quoted strings, comments, quoted identifiers, position tracking |
| `internal/sql/statement.go` | Statement: `.Words()`, `.ContainsSeq()`, `.FirstWord()` |
| `internal/report/formatter.go` | Output: text (colorized, grouped by file) or JSON; summary counts |
| `internal/report/explains.go` | Markdown explanations for each danger rule (used with `-explain`) |
| `test/golden_test.go` | Golden file tests against `test/data/rules/` and `test/data/migrations/` |
| `.pgvet.yaml` | Default config — all rules enabled with default severities |
| `Makefile` | All build/test/lint commands |

---

## Data Flow

```
CLI flags
  → config.Load()                         # parse .pgvet.yaml
  → linter.Registry.Build(config)         # instantiate enabled rules
  → linter.Linter.Lint(dir|file)
      → filepath.Walk → collect .sql files
      → goroutine per file:
          → sql.Tokenize(content)
          → []Statement (split by semicolons/directives)
          → for each rule: rule.Check(file) → []Issue
          → filter by sqlint:disable directives
      → merge + sort issues
  → report.Formatter.Format(issues)       # text or JSON
  → exit 0 / 1 / 2
```

---

## External Dependencies

| Dependency | Role |
|-----------|------|
| `gopkg.in/yaml.v3` | Config file parsing |
| `github.com/stretchr/testify` | Test assertions (require/assert) |

No external SQL parser. All tokenization is custom (`internal/sql`).

---

## Domain Concepts

- **Danger rule** — detects a DDL operation that can lock tables, cause data loss, or break replication (e.g., `CREATE INDEX` without `CONCURRENTLY`)
- **Style rule** — detects formatting/consistency violation (e.g., mixed keyword case)
- **Issue** — a single violation: file path, line, column, rule name, severity, message
- **File** — parsed representation of a `.sql` file: path + list of `Statement`
- **Statement** — a sequence of SQL tokens between semicolons
- **Token** — lexical unit with type (Word, QuotedIdent, String, DollarStr, Number, Comment, Whitespace, Operator) and position
- **Directive** — inline SQL comment `-- sqlint:disable rule1,rule2` / `sqlint:enable` to suppress specific rules per statement
- **Severity** — error, warning, info; configurable per rule

---

## Development Notes

- **No AST** — rules operate on token sequences, not a parsed AST. Pattern matching via `ContainsSeq()` on `[]Token`.
- **Dollar-quoted strings** — tokenizer handles PostgreSQL `$$...$$` and named variants `$tag$...$tag$`. Rules must not flag content inside them.
- **Concurrent processing** — files processed in parallel with goroutines; results merged after all complete.
- **Exit codes** — `0` clean, `1` linting violations found, `2` execution error (config/IO). `-no-exit-code` forces `0` even with violations.
- **Color output** — auto-detected via TTY check; disabled by `NO_COLOR` env var or `TERM=dumb`.
- **Golden tests** — `test/data/rules/<rule>/error.sql`, `ok.sql`, `nolint.sql` per rule. Adding a new rule requires adding these test fixtures.
- **`redundant-index`** — the only cross-statement rule; it accumulates index definitions across the entire file.
- **Config defaults** — if no config file found, all rules enabled with default severities.

---

## Commands

```bash
make build        # build ./bin/pgvet
make test         # run all tests
make lint         # run golangci-lint
make run          # build and run with default flags
make clean        # remove ./bin/
make dependencies # install tool dependencies to ./bin/
```
