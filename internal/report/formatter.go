package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
)

// Format selects the output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Formatter writes linting issues to an io.Writer.
type Formatter struct {
	format Format
}

// New returns a Formatter for the given output format.
func New(format Format) *Formatter {
	return &Formatter{format: format}
}

// Write sorts and writes all issues to w.
func (f *Formatter) Write(w io.Writer, issues []rule.Issue) error {
	sortIssues(issues)

	switch f.format {
	case FormatText:
		return f.writeText(w, issues)
	case FormatJSON:
		return f.writeJSON(w, issues)
	default:
		return f.writeText(w, issues)
	}
}

// Explain writes a detailed explanation for each unique danger rule found in issues.
func (f *Formatter) Explain(w io.Writer, issues []rule.Issue) {
	col := newColorizer(w)

	// collect unique rules in order of first appearance
	seen := make(map[string]bool)

	var unique []string

	for _, issue := range issues {
		if _, ok := explains[issue.Rule]; !ok {
			continue
		}

		if !seen[issue.Rule] {
			seen[issue.Rule] = true
			unique = append(unique, issue.Rule)
		}
	}

	if len(unique) == 0 {
		return
	}

	fmt.Fprintf(w, "\n  %s\n", col.bold("Explain"))

	for _, name := range unique {
		fmt.Fprintf(w, "\n  %s\n  %s\n", col.yellow(name), explains[name])
	}
}

// Summary writes a totals summary line to w.
func (f *Formatter) Summary(w io.Writer, issues []rule.Issue) {
	col := newColorizer(w)

	errors, warnings := 0, 0

	for _, issue := range issues {
		switch issue.Severity {
		case config.SeverityError:
			errors++
		case config.SeverityWarning:
			warnings++
		case config.SeverityInfo:
			// info severity does not count as error or warning
		}
	}

	if len(issues) == 0 {
		fmt.Fprintln(w, col.green("\n  No issues found."))

		return
	}

	total := fmt.Sprintf("\n  %d issue(s) found", len(issues))
	parts := fmt.Sprintf(": %d error(s), %d warning(s)", errors, warnings)

	if errors > 0 {
		fmt.Fprintln(w, col.red(total)+col.dim(parts))
	} else {
		fmt.Fprintln(w, col.green(total)+col.dim(parts))
	}
}

func (f *Formatter) writeText(w io.Writer, issues []rule.Issue) error {
	col := newColorizer(w)

	var currentFile string

	for _, issue := range issues {
		if issue.File != currentFile {
			currentFile = issue.File
			fmt.Fprintf(w, "\n  %s\n\n", col.bold(issue.File))
		}

		pos := fmt.Sprintf("%d:%d", issue.Line, issue.Col)
		posStr := fmt.Sprintf("%-8s", pos)
		ruleStr := fmt.Sprintf("%-30s", issue.Rule)

		var sevStr string

		switch issue.Severity {
		case config.SeverityError:
			sevStr = col.red(fmt.Sprintf("%-7s", "error"))
		case config.SeverityWarning:
			sevStr = col.yellow(fmt.Sprintf("%-7s", "warning"))
		case config.SeverityInfo:
			sevStr = col.blue(fmt.Sprintf("%-7s", "info"))
		default:
			sevStr = col.blue(fmt.Sprintf("%-7s", "info"))
		}

		fmt.Fprintf(w, "    %s  %s  %s  %s\n",
			col.dim(posStr),
			sevStr,
			col.cyan(ruleStr),
			issue.Message,
		)
	}

	return nil
}

// jsonIssue is the JSON representation of an Issue.
type jsonIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

func (f *Formatter) writeJSON(w io.Writer, issues []rule.Issue) error {
	output := make([]jsonIssue, len(issues))

	for i, issue := range issues {
		output[i] = jsonIssue{
			File:     issue.File,
			Line:     issue.Line,
			Col:      issue.Col,
			Severity: string(issue.Severity),
			Rule:     issue.Rule,
			Message:  issue.Message,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	err := enc.Encode(output)
	if err != nil {
		return fmt.Errorf("json.Encode: %w", err)
	}

	return nil
}

// sortIssues sorts issues by file, line, then column.
func sortIssues(issues []rule.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]

		if a.File != b.File {
			return a.File < b.File
		}

		if a.Line != b.Line {
			return a.Line < b.Line
		}

		return a.Col < b.Col
	})
}

// colorizer applies ANSI codes only when writing to a terminal.
type colorizer struct {
	enabled bool
}

func newColorizer(w io.Writer) colorizer {
	return colorizer{enabled: isTTY(w)}
}

func (c colorizer) wrap(code, s string) string {
	if !c.enabled {
		return s
	}

	return code + s + "\033[0m"
}

func (c colorizer) bold(s string) string   { return c.wrap("\033[1m", s) }
func (c colorizer) dim(s string) string    { return c.wrap("\033[2m", s) }
func (c colorizer) red(s string) string    { return c.wrap("\033[1;31m", s) }
func (c colorizer) yellow(s string) string { return c.wrap("\033[1;33m", s) }
func (c colorizer) blue(s string) string   { return c.wrap("\033[1;34m", s) }
func (c colorizer) green(s string) string  { return c.wrap("\033[1;32m", s) }
func (c colorizer) cyan(s string) string   { return c.wrap("\033[36m", s) }

// isTTY reports whether w is an interactive terminal with colors allowed.
func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}
