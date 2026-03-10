# pgvet - Lint your sql easily

[![Tests](https://github.com/sadpenguinn/pgvet/actions/workflows/test.yaml/badge.svg)](https://github.com/sadpenguinn/pgvet/actions/workflows/test.yaml)
[![Lint](https://github.com/sadpenguinn/pgvet/actions/workflows/lint.yaml/badge.svg)](https://github.com/sadpenguinn/pgvet/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sadpenguinn/pgvet)](https://goreportcard.com/report/github.com/sadpenguinn/pgvet)
[![Release](https://img.shields.io/github/v/release/sadpenguinn/pgvet)](https://github.com/sadpenguinn/pgvet/releases)
![go](https://img.shields.io/badge/go-%3E%3D1.26-blue)
![license](https://img.shields.io/badge/license-MIT-blue)

---

> ☔️ A Go-friendly linter for PostgreSQL migration SQL files.

Most SQL linters are built for DBAs — heavy and not designed for Go projects. **pgvet** is different: it's a lightweight static analyzer that fits naturally into Go workflows, catches dangerous DDL operations before they lock your production database, and enforces style consistency across your migration files.

## Why pgvet?

There's no SQL linter that works the way Go developers expect. `pgvet` runs like `golangci-lint`, integrates into CI/CD like any Go tool, and understands the specific dangers of PostgreSQL DDL — things like `CREATE INDEX` without `CONCURRENTLY`, or `ALTER COLUMN TYPE` that rewrites the entire table.

---

## Installation

```bash
go install github.com/sadpenguinn/pgvet/cmd/pgvet@latest
```

Or build from source:

```bash
git clone https://github.com/sadpenguinn/pgvet.git
cd pgvet
make build
```

## Usage

### Examples

```bash
# Lint all .sql files in current directory
pgvet

# Lint a specific directory
pgvet -dir ./migrations

# Lint a single file
pgvet -file ./migrations/001_users.sql

# JSON output (for CI/tooling)
pgvet -format json

# Show detailed explanations for violations
pgvet -explain

# Use a custom config file
pgvet -config ./my-pgvet.yaml
```

### Rules

**Danger rules** — catch operations that cause table locks, data loss, or replication lag:

| Rule | Description |
|------|-------------|
| `create-index-no-concurrently` | `CREATE INDEX` without `CONCURRENTLY` locks the table |
| `drop-index-no-concurrently` | `DROP INDEX` without `CONCURRENTLY` locks the table |
| `add-column-not-null` | Adding a `NOT NULL` column without a `DEFAULT` rewrites the table |
| `set-not-null` | `SET NOT NULL` requires a full table scan |
| `add-foreign-key-no-valid` | FK without `NOT VALID` locks both tables during validation |
| `drop-table` | Irreversible data loss |
| `drop-column` | Irreversible data loss |
| `truncate` | Irreversible data loss |
| `lock-table` | Explicit table lock — almost always avoidable |
| `rename` | Breaks dependent application code |
| `change-column-type` | Rewrites the entire table |
| `redundant-index` | Duplicate index wastes space and slows writes |

**Style rules** — keep migrations readable and consistent:

| Rule | Description |
|------|-------------|
| `keyword-case` | SQL keywords must be consistently upper or lowercase |
| `require-semicolon` | Every statement must end with `;` |
| `trailing-whitespace` | No trailing spaces |
| `max-line-length` | Optional line length limit |

### Configuration

Create a `.pgvet.yaml` in your project root:

```yaml
danger:
  create-index-no-concurrently:
    enabled: true
    severity: error
  drop-table:
    enabled: true
    severity: warning

style:
  keyword-case:
    enabled: true
    severity: warning
  require-semicolon:
    enabled: true
    severity: error
```

### Disabling rules inline

```sql
-- sqlint:disable drop-table
DROP TABLE legacy_users;
-- sqlint:enable drop-table
```

---

## 🌀 CI Integration

```yaml
- name: Lint SQL migrations
  run: pgvet -dir ./migrations -format json
```

## 🤝 Contributing

The project is under active development — new rules, edge cases, and integrations are being added regularly. PRs are very welcome.

If you want to contribute:
- Open an issue to discuss the idea first
- Add test fixtures under `test/data/rules/<rule-name>/` for new rules
- Run `make test` and `make lint` before submitting

## 📁 License

This project is under [MIT](./LICENSE) license.
