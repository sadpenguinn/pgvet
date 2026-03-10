package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/linter"
	"github.com/sadpenguinn/pgvet/internal/report"
	"github.com/sadpenguinn/pgvet/internal/rule"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		configPath  = flag.String("config", "", "path to config file (default: .pgvet.yaml)")
		dir         = flag.String("dir", ".", "directory containing .sql files")
		file        = flag.String("file", "", "lint a single file (overrides -dir)")
		format      = flag.String("format", "text", "output format: text, json")
		explain     = flag.Bool("explain", false, "print detailed explanation for each unique danger rule violation")
		noExitCode  = flag.Bool("no-exit-code", false, "do not exit with code 1 when errors are found")
		showVersion = flag.Bool("version", false, "print version and exit")
	)

	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, "pgvet version "+version)

		return 0
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)

		return 2
	}

	rules := linter.Build(cfg)
	l := linter.New(rules)
	formatter := report.New(report.Format(*format))

	issues, err := collectIssues(l, *file, *dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error linting: %v\n", err)

		return 2
	}

	err = formatter.Write(os.Stdout, issues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)

		return 2
	}

	if *format == "text" {
		if *explain {
			formatter.Explain(os.Stdout, issues)
		}

		formatter.Summary(os.Stdout, issues)
	}

	if !*noExitCode && hasErrors(issues) {
		return 1
	}

	return 0
}

// loadConfig loads config from the given path, falling back to well-known
// filenames in the current directory, then to built-in defaults.
func loadConfig(path string) (*config.Config, error) {
	if path == "" {
		candidates := []string{".pgvet.yaml", ".pgvet.yml", "pgvet.yaml", "pgvet.yml"}
		for _, c := range candidates {
			_, statErr := os.Stat(c)
			if statErr == nil {
				path = c

				break
			}
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	return cfg, nil
}

func collectIssues(l *linter.Linter, file, dir string) ([]rule.Issue, error) {
	if file != "" {
		issues, err := l.LintFile(file)
		if err != nil {
			return nil, fmt.Errorf("l.LintFile: %w", err)
		}

		return issues, nil
	}

	issues, err := l.LintDir(dir)
	if err != nil {
		return nil, fmt.Errorf("l.LintDir: %w", err)
	}

	return issues, nil
}

func hasErrors(issues []rule.Issue) bool {
	for _, issue := range issues {
		if issue.Severity == config.SeverityError {
			return true
		}
	}

	return false
}
