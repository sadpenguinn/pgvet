package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sadpenguinn/pgvet/internal/rule"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// Linter checks SQL files against a set of rules.
type Linter struct {
	rules []rule.Rule
}

// New returns a Linter that applies the given rules.
func New(rules []rule.Rule) *Linter {
	return &Linter{rules: rules}
}

// LintDir checks all .sql files in dir recursively.
func (l *Linter) LintDir(dir string) ([]rule.Issue, error) {
	paths, err := collectSQLFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("collectSQLFiles: %w", err)
	}

	return l.lintFiles(paths), nil
}

// LintFile checks a single SQL file.
func (l *Linter) LintFile(path string) ([]rule.Issue, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	return l.lintContent(path, string(content)), nil
}

// lintFiles checks multiple files concurrently.
func (l *Linter) lintFiles(paths []string) []rule.Issue {
	type result struct{ issues []rule.Issue }

	results := make([]result, len(paths))

	var wg sync.WaitGroup

	for i, path := range paths {
		wg.Add(1)

		go func(idx int, p string) {
			defer wg.Done()

			content, err := os.ReadFile(p)
			if err != nil {
				return
			}

			results[idx].issues = l.lintContent(p, string(content))
		}(i, path)
	}

	wg.Wait()

	var all []rule.Issue

	for _, r := range results {
		all = append(all, r.issues...)
	}

	return all
}

// lintContent checks the SQL content of a single file.
func (l *Linter) lintContent(path, content string) []rule.Issue {
	tokens := sql.Tokenize(content)
	statements := sql.SplitStatements(tokens)
	lines := strings.Split(content, "\n")

	file := &rule.File{
		Path:       path,
		Content:    content,
		Lines:      lines,
		Statements: statements,
	}

	directives := parseDirectives(tokens)

	var issues []rule.Issue

	for _, r := range l.rules {
		for _, issue := range r.Check(file) {
			if isDisabled(directives, issue.Rule, issue.Line) {
				continue
			}

			issue.File = path
			issues = append(issues, issue)
		}
	}

	return issues
}

// collectSQLFiles recursively collects paths to all .sql files under dir.
func collectSQLFiles(dir string) ([]string, error) {
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".sql" {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.WalkDir: %w", err)
	}

	return paths, nil
}
