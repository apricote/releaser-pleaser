package conventionalcommits

import (
	"fmt"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

type Parser struct {
	machine conventionalcommits.Machine
}

func NewParser() *Parser {
	parserMachine := parser.NewMachine(
		parser.WithBestEffort(),
		parser.WithTypes(conventionalcommits.TypesConventional),
	)

	return &Parser{
		machine: parserMachine,
	}
}

func (c *Parser) Analyze(commits []git.Commit) ([]commitparser.AnalyzedCommit, error) {
	analyzedCommits := make([]commitparser.AnalyzedCommit, 0, len(commits))

	for _, commit := range commits {
		msg, err := c.machine.Parse([]byte(commit.Message))
		if err != nil {
			return nil, fmt.Errorf("failed to parse message of commit %q: %w", commit.Hash, err)
		}
		conventionalCommit, ok := msg.(*conventionalcommits.ConventionalCommit)
		if !ok {
			return nil, fmt.Errorf("unable to get ConventionalCommit from parser result: %T", msg)
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
