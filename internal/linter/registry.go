package linter

import (
	"github.com/sadpenguinn/pgvet/internal/config"
	"github.com/sadpenguinn/pgvet/internal/rule"
	dangerous "github.com/sadpenguinn/pgvet/internal/rule/danger"
	"github.com/sadpenguinn/pgvet/internal/rule/style"
)

// Build constructs the list of enabled rules from the given configuration.
func Build(cfg *config.Config) []rule.Rule {
	var rules []rule.Rule

	d := &cfg.Rules.Danger
	rules = appendIfEnabled(rules, d.CreateIndexNoConcurrently, func() rule.Rule {
		return dangerous.NewCreateIndex(d.CreateIndexNoConcurrently.Severity)
	})
	rules = appendIfEnabled(rules, d.DropIndexNoConcurrently, func() rule.Rule {
		return dangerous.NewDropIndex(d.DropIndexNoConcurrently.Severity)
	})
	rules = appendIfEnabled(rules, d.AddColumnNotNull, func() rule.Rule {
		return dangerous.NewAddColumnNotNull(d.AddColumnNotNull.Severity)
	})
	rules = appendIfEnabled(rules, d.SetNotNull, func() rule.Rule {
		return dangerous.NewSetNotNull(d.SetNotNull.Severity)
	})
	rules = appendIfEnabled(rules, d.AddForeignKeyNoValid, func() rule.Rule {
		return dangerous.NewAddForeignKey(d.AddForeignKeyNoValid.Severity)
	})
	rules = appendIfEnabled(rules, d.DropTable, func() rule.Rule {
		return dangerous.NewDropTable(d.DropTable.Severity)
	})
	rules = appendIfEnabled(rules, d.DropColumn, func() rule.Rule {
		return dangerous.NewDropColumn(d.DropColumn.Severity)
	})
	rules = appendIfEnabled(rules, d.Truncate, func() rule.Rule {
		return dangerous.NewTruncate(d.Truncate.Severity)
	})
	rules = appendIfEnabled(rules, d.LockTable, func() rule.Rule {
		return dangerous.NewLockTable(d.LockTable.Severity)
	})
	rules = appendIfEnabled(rules, d.Rename, func() rule.Rule {
		return dangerous.NewRename(d.Rename.Severity)
	})
	rules = appendIfEnabled(rules, d.ChangeColumnType, func() rule.Rule {
		return dangerous.NewChangeColumnType(d.ChangeColumnType.Severity)
	})
	rules = appendIfEnabled(rules, d.RedundantIndex, func() rule.Rule {
		return dangerous.NewRedundantIndex(d.RedundantIndex.Severity)
	})

	s := &cfg.Rules.Style
	rules = appendIfEnabled(rules, s.KeywordCase.Rule, func() rule.Rule {
		return style.NewKeywordCase(s.KeywordCase.Severity, s.KeywordCase.Case)
	})
	rules = appendIfEnabled(rules, s.TrailingWhitespace, func() rule.Rule {
		return style.NewTrailingWhitespace(s.TrailingWhitespace.Severity)
	})
	rules = appendIfEnabled(rules, s.MaxLineLength.Rule, func() rule.Rule {
		return style.NewMaxLineLength(s.MaxLineLength.Severity, s.MaxLineLength.Max)
	})
	rules = appendIfEnabled(rules, s.RequireSemicolon, func() rule.Rule {
		return style.NewRequireSemicolon(s.RequireSemicolon.Severity)
	})

	return rules
}

func appendIfEnabled(rules []rule.Rule, cfg config.Rule, factory func() rule.Rule) []rule.Rule {
	if !cfg.Enabled {
		return rules
	}

	return append(rules, factory())
}
