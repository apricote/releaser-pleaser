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
