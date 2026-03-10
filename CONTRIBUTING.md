# Contributing to pgvet

Thank you for your interest in contributing!

## Before You Start

For non-trivial changes, open an issue first to discuss the idea. This avoids wasted effort if the direction doesn't fit the project.

## Development Setup

```bash
git clone https://github.com/sadpenguinn/pgvet.git
cd pgvet
make build
make test
```

## How to Add a New Rule

1. Create the rule file in `internal/rule/danger/` or `internal/rule/style/`
2. Register it in `internal/linter/registry.go`
3. Add config support in `internal/config/config.go`
4. Add test fixtures:
   - `test/data/rules/<rule-name>/error.sql` — SQL that should trigger the rule
   - `test/data/rules/<rule-name>/ok.sql` — SQL that should pass
   - `test/data/rules/<rule-name>/nolint.sql` — SQL with inline disable directive
5. Run `make test` — golden tests will validate your fixtures automatically

## Code Standards

- Run `make lint` before submitting — CI will reject linting failures
- Each function should do one thing; keep functions under 30 lines
- Wrap errors with context: `fmt.Errorf("repo.GetUser: %w", err)`
- Comments explain *why*, not *what*; write them in Russian
- No external dependencies without prior discussion

## Testing

- All rule logic must be covered by golden file tests in `test/data/`
- `make test` runs the full test suite
- Do not submit PRs with failing tests

## Pull Request Process

1. Fork the repository and create a feature branch from `master`
2. Make your changes and ensure `make test` and `make lint` both pass
3. Fill in the PR template
4. Link the relevant issue in your PR description

## Commit Style

Use clear, imperative commit messages:

```
add redundant-index rule
fix tokenizer handling of dollar-quoted strings
improve error message for add-column-not-null
```

No prefixes (`feat:`, `fix:` etc.) — just plain English describing what was done.
