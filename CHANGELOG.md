# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.1.0] - 2025-01-01

### Added
- 12 danger rules: `create-index-no-concurrently`, `drop-index-no-concurrently`, `add-column-not-null`, `set-not-null`, `add-foreign-key-no-valid`, `drop-table`, `drop-column`, `truncate`, `lock-table`, `rename`, `change-column-type`, `redundant-index`
- 4 style rules: `keyword-case`, `require-semicolon`, `trailing-whitespace`, `max-line-length`
- Text and JSON output formats
- `-explain` flag for detailed rule explanations
- Inline disable directives: `-- sqlint:disable rule`, `-- sqlint:enable rule`
- YAML configuration via `.pgvet.yaml`
- Custom SQL tokenizer with dollar-quoted string support
- Concurrent file processing

[Unreleased]: https://github.com/sadpenguinn/pgvet/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/sadpenguinn/pgvet/releases/tag/v0.1.0
