package rp

import (
	"fmt"

	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
)

type AnalyzedCommit struct {
	Commit
	Type        string
	Description string
	Scope       *string
}

func AnalyzeCommits(commits []Commit) ([]AnalyzedCommit, conventionalcommits.VersionBump, error) {
	parserMachine := parser.NewMachine(
		parser.WithBestEffort(),
		parser.WithTypes(conventionalcommits.TypesConventional),
	)

	analyzedCommits := make([]AnalyzedCommit, 0, len(commits))

	highestVersionBump := conventionalcommits.UnknownVersion

	for _, commit := range commits {
		msg, err := parserMachine.Parse([]byte(commit.Message))
		if err != nil {
			return nil, conventionalcommits.UnknownVersion, fmt.Errorf("failed to parse message of commit %q: %w", commit.Hash, err)
		}
		conventionalCommit, ok := msg.(*conventionalcommits.ConventionalCommit)
		if !ok {
			return nil, conventionalcommits.UnknownVersion, fmt.Errorf("unable to get ConventionalCommit from parser result: %T", msg)
		}

		commitVersionBump := conventionalCommit.VersionBump(conventionalcommits.DefaultStrategy)
		if commitVersionBump > conventionalcommits.UnknownVersion {
			// We only care about releasable commits
			analyzedCommits = append(analyzedCommits, AnalyzedCommit{
				Commit:      commit,
				Type:        conventionalCommit.Type,
				Description: conventionalCommit.Description,
				Scope:       conventionalCommit.Scope,
			})
		}

		if commitVersionBump > highestVersionBump {
			// Get max version bump from all releasable commits
			highestVersionBump = commitVersionBump
		}
	}

	return analyzedCommits, highestVersionBump, nil
}
