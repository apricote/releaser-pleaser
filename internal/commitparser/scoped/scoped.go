// Package scoped implements a commit parser for Scoped Commits (https://scopedcommits.com/), messages in the
// format "scope: message".
//
// Every commit is considered releasable. Commits that do not follow the format are included with the whole subject
// line as the description and without a scope.
package scoped

import (
	"log/slog"
	"regexp"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

// TypeChange is the commit type assigned to all commits analyzed by this parser.
const TypeChange = "change"

var subjectRegex = regexp.MustCompile(`^([^\s:]+):\s+(.+)$`)

type Parser struct {
	logger *slog.Logger
}

func NewParser(logger *slog.Logger) *Parser {
	return &Parser{logger: logger}
}

func (p *Parser) Analyze(commits []git.Commit) ([]commitparser.AnalyzedCommit, error) {
	analyzedCommits := make([]commitparser.AnalyzedCommit, 0, len(commits))

	for _, commit := range commits {
		subject := commit.Subject()
		if subject == "" {
			p.logger.Warn("commit message is empty, skipping", "commit.hash", commit.Hash)
			continue
		}

		analyzed := commitparser.AnalyzedCommit{
			Commit:      commit,
			Type:        TypeChange,
			Description: subject,
		}

		if matches := subjectRegex.FindStringSubmatch(subject); matches != nil {
			scope := matches[1]
			analyzed.Scope = &scope
			analyzed.Description = matches[2]
		}

		analyzedCommits = append(analyzedCommits, analyzed)
	}

	return analyzedCommits, nil
}
