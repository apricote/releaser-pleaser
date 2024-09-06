package rp

import (
	"strings"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/markdown"
)

func parsePRBodyForCommitOverrides(commits []git.Commit) ([]git.Commit, error) {
	result := make([]git.Commit, 0, len(commits))

	for _, commit := range commits {
		singleResult, err := parseSinglePRBodyForCommitOverrides(commit)
		if err != nil {
			return nil, err
		}

		result = append(result, singleResult...)
	}

	return result, nil
}

func parseSinglePRBodyForCommitOverrides(commit git.Commit) ([]git.Commit, error) {
	if commit.PullRequest == nil {
		return []git.Commit{commit}, nil
	}

	source := []byte(commit.PullRequest.Description)
	var overridesText string
	var found bool
	err := markdown.WalkAST(source, markdown.GetCodeBlockText(source, "rp-commits", &overridesText, &found))
	if err != nil {
		return nil, err
	}

	if !found {
		return []git.Commit{commit}, nil
	}

	lines := strings.Split(overridesText, "\n")
	result := make([]git.Commit, 0, len(lines))
	for _, line := range lines {
		// Only consider lines with text
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		newCommit := commit
		newCommit.Message = line
		result = append(result, newCommit)
	}

	return result, nil
}
