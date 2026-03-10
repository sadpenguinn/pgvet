package rule

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/sql"
)

// File holds all parsed information about a SQL source file.
type File struct {
	Path       string
	Content    string
	Lines      []string
	Statements []*sql.Statement
}

// Issue is a problem found in a SQL file.
type Issue struct {
	File     string
	Line     int
	Col      int
	Rule     string
	Severity config.Severity
	Message  string
}

// Rule is the interface implemented by all linter rules.
type Rule interface {
	Name() string
	Check(file *File) []Issue
}
