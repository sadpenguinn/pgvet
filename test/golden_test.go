package golden_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/linter"
	"github.com/sadpenguinn/pgvet/internal/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// TestRules checks each rule in isolation through three cases:
// error (violation), ok (correct code), nolint (suppressed violation).
func TestRules(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("data", "rules"))
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ruleName := entry.Name()

		t.Run(ruleName, func(t *testing.T) {
			t.Parallel()

			cfg := singleRuleConfig(ruleName)
			runCases(t, cfg, filepath.Join("data", "rules", ruleName))
		})
	}
}

// TestMigrations checks real migration files from popular Go tools.
// Structure: test/data/migrations/<tool>/*.sql.
func TestMigrations(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join("data", "migrations"))
	require.NoError(t, err)

	cfg := config.Default()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		toolName := entry.Name()

		t.Run(toolName, func(t *testing.T) {
			t.Parallel()

			runDir(t, cfg, filepath.Join("data", "migrations", toolName))
		})
	}
}

func runCases(t *testing.T, cfg *config.Config, dir string) {
	t.Helper()

	for _, caseName := range []string{"error", "ok", "nolint"} {
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()

			sqlPath := filepath.Join(dir, caseName+".sql")
			goldenPath := filepath.Join(dir, caseName+".golden")
			runFile(t, cfg, sqlPath, goldenPath)
		})
	}
}

func runDir(t *testing.T, cfg *config.Config, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	rules := linter.Build(cfg)
	l := linter.New(rules)
	formatter := report.New(report.FormatText)

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}

		base := strings.TrimSuffix(name, ".sql")

		t.Run(base, func(t *testing.T) {
			t.Parallel()

			sqlPath := filepath.Join(dir, name)
			goldenPath := filepath.Join(dir, base+".golden")
			runFileWith(t, l, formatter, sqlPath, goldenPath)
		})
	}
}

func runFile(t *testing.T, cfg *config.Config, sqlPath, goldenPath string) {
	t.Helper()

	rules := linter.Build(cfg)
	l := linter.New(rules)
	formatter := report.New(report.FormatText)
	runFileWith(t, l, formatter, sqlPath, goldenPath)
}

func runFileWith(t *testing.T, l *linter.Linter, formatter *report.Formatter, sqlPath, goldenPath string) {
	t.Helper()

	issues, err := l.LintFile(sqlPath)
	require.NoError(t, err)

	for i := range issues {
		issues[i].File = filepath.Base(issues[i].File)
	}

	var sb strings.Builder

	err = formatter.Write(&sb, issues)
	require.NoError(t, err)

	got := sb.String()

	if *update {
		err = os.WriteFile(goldenPath, []byte(got), 0o600)

		require.NoError(t, err)

		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	assert.Equal(t, string(want), got)
}

// singleRuleConfig returns a config with only the specified rule enabled.
func singleRuleConfig(name string) *config.Config {
	cfg := &config.Config{}

	switch name {
	case "create-index-no-concurrently":
		cfg.Rules.Danger.CreateIndexNoConcurrently = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "drop-index-no-concurrently":
		cfg.Rules.Danger.DropIndexNoConcurrently = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "add-column-not-null":
		cfg.Rules.Danger.AddColumnNotNull = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "set-not-null":
		cfg.Rules.Danger.SetNotNull = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "add-foreign-key-no-valid":
		cfg.Rules.Danger.AddForeignKeyNoValid = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "drop-table":
		cfg.Rules.Danger.DropTable = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "drop-column":
		cfg.Rules.Danger.DropColumn = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "truncate":
		cfg.Rules.Danger.Truncate = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "lock-table":
		cfg.Rules.Danger.LockTable = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "rename":
		cfg.Rules.Danger.Rename = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "change-column-type":
		cfg.Rules.Danger.ChangeColumnType = config.Rule{Enabled: true, Severity: config.SeverityError}
	case "redundant-index":
		cfg.Rules.Danger.RedundantIndex = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "keyword-case":
		cfg.Rules.Style.KeywordCase = config.KeywordCaseRule{
			Rule: config.Rule{Enabled: true, Severity: config.SeverityWarning},
			Case: config.CaseUpper,
		}
	case "trailing-whitespace":
		cfg.Rules.Style.TrailingWhitespace = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	case "max-line-length":
		cfg.Rules.Style.MaxLineLength = config.MaxLineLengthRule{
			Rule: config.Rule{Enabled: true, Severity: config.SeverityWarning},
			Max:  120,
		}
	case "require-semicolon":
		cfg.Rules.Style.RequireSemicolon = config.Rule{Enabled: true, Severity: config.SeverityWarning}
	}

	return cfg
}
