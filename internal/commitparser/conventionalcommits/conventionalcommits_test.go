package conventionalcommits

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

func TestAnalyzeCommits(t *testing.T) {
	tests := []struct {
		name            string
		commits         []git.Commit
		expectedCommits []commitparser.AnalyzedCommit
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "empty commits",
			commits:         []git.Commit{},
			expectedCommits: []commitparser.AnalyzedCommit{},
			wantErr:         assert.NoError,
		},
		{
			name: "malformed commit message",
			commits: []git.Commit{
				{
					Message: "aksdjaklsdjka",
				},
			},
			expectedCommits: nil,
			wantErr:         assert.Error,
		},
		{
			name: "drops unreleasable",
			commits: []git.Commit{
				{
					Message: "chore: foobar",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{},
			wantErr:         assert.NoError,
		},
		{
			name: "highest bump (patch)",
			commits: []git.Commit{
				{
					Message: "chore: foobar",
				},
				{
					Message: "fix: blabla",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{
				{
					Commit:      git.Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "highest bump (minor)",
			commits: []git.Commit{
				{
					Message: "fix: blabla",
				},
				{
					Message: "feat: foobar",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{
				{
					Commit:      git.Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
				{
					Commit:      git.Commit{Message: "feat: foobar"},
					Type:        "feat",
					Description: "foobar",
				},
			},
			wantErr: assert.NoError,
		},

		{
			name: "highest bump (major)",
			commits: []git.Commit{
				{
					Message: "fix: blabla",
				},
				{
					Message: "feat!: foobar",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{
				{
					Commit:      git.Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
				{
					Commit:         git.Commit{Message: "feat!: foobar"},
					Type:           "feat",
					Description:    "foobar",
					BreakingChange: true,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzedCommits, err := NewParser().Analyze(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.expectedCommits, analyzedCommits)
		})
	}
}
