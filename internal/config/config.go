package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root linter configuration.
type Config struct {
	Rules Rules `yaml:"rules"`
}

// Rules groups all rule categories.
type Rules struct {
	Danger Danger `yaml:"danger"`
	Style  Style  `yaml:"style"`
}

// Severity configures error levels.
type Severity string

const (
	// SeverityError is error.
	SeverityError Severity = "error"
	// SeverityWarning is warning.
	SeverityWarning Severity = "warning"
	// SeverityInfo is info.
	SeverityInfo Severity = "info"
)

// Rule is the base configuration shared by every rule.
type Rule struct {
	Enabled  bool     `yaml:"enabled"`
	Severity Severity `yaml:"severity"`
}

// Danger holds configuration for dangerous DDL rules.
type Danger struct {
	CreateIndexNoConcurrently Rule `yaml:"create-index-no-concurrently"`
	DropIndexNoConcurrently   Rule `yaml:"drop-index-no-concurrently"`
	AddColumnNotNull          Rule `yaml:"add-column-not-null"`
	SetNotNull                Rule `yaml:"set-not-null"`
	AddForeignKeyNoValid      Rule `yaml:"add-foreign-key-no-valid"`
	DropTable                 Rule `yaml:"drop-table"`
	DropColumn                Rule `yaml:"drop-column"`
	Truncate                  Rule `yaml:"truncate"`
	LockTable                 Rule `yaml:"lock-table"`
	Rename                    Rule `yaml:"rename"`
	ChangeColumnType          Rule `yaml:"change-column-type"`
	RedundantIndex            Rule `yaml:"redundant-index"`
}

// Style holds configuration for style rules.
type Style struct {
	KeywordCase        KeywordCaseRule   `yaml:"keyword-case"`
	TrailingWhitespace Rule              `yaml:"trailing-whitespace"`
	MaxLineLength      MaxLineLengthRule `yaml:"max-line-length"`
	RequireSemicolon   Rule              `yaml:"require-semicolon"`
}

// Case configures style.
type Case string

const (
	// CaseUpper is upper.
	CaseUpper Case = "upper"
	// CaseLower is lower.
	CaseLower Case = "lower"
)

// KeywordCaseRule extends Rule with a case preference.
type KeywordCaseRule struct {
	Rule `yaml:",inline"`

	Case Case `yaml:"case"`
}

// MaxLineLengthRule extends Rule with a maximum length.
type MaxLineLengthRule struct {
	Rule `yaml:",inline"`

	Max int `yaml:"max"`
}

// Load reads configuration from the given file path.
// When path is empty the default configuration is returned.
func Load(path string) (*Config, error) {
	if path == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	config := Default()

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return config, nil
}

// Default returns the built-in configuration with all rules enabled.
func Default() *Config {
	return &Config{
		Rules: Rules{
			Danger: Danger{
				CreateIndexNoConcurrently: Rule{Enabled: true, Severity: SeverityError},
				DropIndexNoConcurrently:   Rule{Enabled: true, Severity: SeverityError},
				AddColumnNotNull:          Rule{Enabled: true, Severity: SeverityError},
				SetNotNull:                Rule{Enabled: true, Severity: SeverityError},
				AddForeignKeyNoValid:      Rule{Enabled: true, Severity: SeverityError},
				DropTable:                 Rule{Enabled: true, Severity: SeverityWarning},
				DropColumn:                Rule{Enabled: true, Severity: SeverityWarning},
				Truncate:                  Rule{Enabled: true, Severity: SeverityWarning},
				LockTable:                 Rule{Enabled: true, Severity: SeverityError},
				Rename:                    Rule{Enabled: true, Severity: SeverityWarning},
				ChangeColumnType:          Rule{Enabled: true, Severity: SeverityError},
				RedundantIndex:            Rule{Enabled: true, Severity: SeverityWarning},
			},
			Style: Style{
				KeywordCase:        KeywordCaseRule{Rule: Rule{Enabled: true, Severity: SeverityError}, Case: CaseUpper},
				TrailingWhitespace: Rule{Enabled: true, Severity: SeverityError},
				MaxLineLength:      MaxLineLengthRule{Rule: Rule{Enabled: false, Severity: SeverityError}, Max: 120},
				RequireSemicolon:   Rule{Enabled: true, Severity: SeverityError},
			},
		},
	}
}
