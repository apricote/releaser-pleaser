package commitparser

import (
	"github.com/apricote/releaser-pleaser/internal/git"
)

type CommitParser interface {
	Analyze(commits []git.Commit) ([]AnalyzedCommit, error)
}

type AnalyzedCommit struct {
	git.Commit
	Type           string
	Description    string
	Scope          *string
	BreakingChange bool
}

// ByType groups the Commits by the type field. Used by the Changelog.
func ByType(in []AnalyzedCommit) map[string][]AnalyzedCommit {
	out := map[string][]AnalyzedCommit{}

	for _, commit := range in {
		if out[commit.Type] == nil {
			out[commit.Type] = make([]AnalyzedCommit, 0, 1)
		}

		out[commit.Type] = append(out[commit.Type], commit)
	}

	return out
}
