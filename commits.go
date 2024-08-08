package rp

import (
	"fmt"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
)

type AnalyzedCommit struct {
	Commit
	Type           string
	Description    string
	Scope          *string
	BreakingChange bool
}

type CommitParser interface {
	Analyze(commits []Commit) ([]AnalyzedCommit, error)
}

type ConventionalCommitsParser struct {
	machine conventionalcommits.Machine
}

func NewConventionalCommitsParser() *ConventionalCommitsParser {
	parserMachine := parser.NewMachine(
		parser.WithBestEffort(),
		parser.WithTypes(conventionalcommits.TypesConventional),
	)

	return &ConventionalCommitsParser{
		machine: parserMachine,
	}
}

func (c *ConventionalCommitsParser) AnalyzeCommits(commits []Commit) ([]AnalyzedCommit, error) {
	analyzedCommits := make([]AnalyzedCommit, 0, len(commits))

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
			analyzedCommits = append(analyzedCommits, AnalyzedCommit{
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
