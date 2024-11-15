package conventionalcommits

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

type Parser struct {
	machine conventionalcommits.Machine
	logger  *slog.Logger
}

func NewParser(logger *slog.Logger) *Parser {
	parserMachine := parser.NewMachine(
		parser.WithBestEffort(),
		parser.WithTypes(conventionalcommits.TypesConventional),
	)

	return &Parser{
		machine: parserMachine,
		logger:  logger,
	}
}

func (c *Parser) Analyze(commits []git.Commit) ([]commitparser.AnalyzedCommit, error) {
	analyzedCommits := make([]commitparser.AnalyzedCommit, 0, len(commits))

	for _, commit := range commits {
		msg, err := c.machine.Parse([]byte(strings.TrimSpace(commit.Message)))
		if err != nil {
			if msg == nil {
				c.logger.Warn("failed to parse message of commit, skipping", "commit.hash", commit.Hash, "err", err)
				continue
			}

			c.logger.Warn("failed to parse message of commit fully, trying to use as much as possible", "commit.hash", commit.Hash, "err", err)
		}

		conventionalCommit, ok := msg.(*conventionalcommits.ConventionalCommit)
		if !ok {
			return nil, fmt.Errorf("unable to get ConventionalCommit from parser result: %T", msg)
		}

		if conventionalCommit.Type == "" {
			// Parsing broke before getting the type, can not use the commit
			c.logger.Warn("commit type was not parsed, skipping", "commit.hash", commit.Hash, "err", err)
			continue
		}

		commitVersionBump := conventionalCommit.VersionBump(conventionalcommits.DefaultStrategy)
		if commitVersionBump > conventionalcommits.UnknownVersion {
			// We only care about releasable commits
			analyzedCommits = append(analyzedCommits, commitparser.AnalyzedCommit{
				Commit:         commit,
				Type:           conventionalCommit.Type,
				Description:    conventionalCommit.Description,
				Scope:          conventionalCommit.Scope,
				BreakingChange: conventionalCommit.IsBreakingChange(),
			})
		}

	}

	return analyzedCommits, nil
}
