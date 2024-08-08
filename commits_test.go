package rp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeCommits(t *testing.T) {
	tests := []struct {
		name            string
		commits         []Commit
		expectedCommits []AnalyzedCommit
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "empty commits",
			commits:         []Commit{},
			expectedCommits: []AnalyzedCommit{},
			wantErr:         assert.NoError,
		},
		{
			name: "malformed commit message",
			commits: []Commit{
				{
					Message: "aksdjaklsdjka",
				},
			},
			expectedCommits: nil,
			wantErr:         assert.Error,
		},
		{
			name: "drops unreleasable",
			commits: []Commit{
				{
					Message: "chore: foobar",
				},
			},
			expectedCommits: []AnalyzedCommit{},
			wantErr:         assert.NoError,
		},
		{
			name: "highest bump (patch)",
			commits: []Commit{
				{
					Message: "chore: foobar",
				},
				{
					Message: "fix: blabla",
				},
			},
			expectedCommits: []AnalyzedCommit{
				{
					Commit:      Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "highest bump (minor)",
			commits: []Commit{
				{
					Message: "fix: blabla",
				},
				{
					Message: "feat: foobar",
				},
			},
			expectedCommits: []AnalyzedCommit{
				{
					Commit:      Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
				{
					Commit:      Commit{Message: "feat: foobar"},
					Type:        "feat",
					Description: "foobar",
				},
			},
			wantErr: assert.NoError,
		},

		{
			name: "highest bump (major)",
			commits: []Commit{
				{
					Message: "fix: blabla",
				},
				{
					Message: "feat!: foobar",
				},
			},
			expectedCommits: []AnalyzedCommit{
				{
					Commit:      Commit{Message: "fix: blabla"},
					Type:        "fix",
					Description: "blabla",
				},
				{
					Commit:         Commit{Message: "feat!: foobar"},
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
			analyzedCommits, err := NewConventionalCommitsParser().AnalyzeCommits(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.expectedCommits, analyzedCommits)
		})
	}
}
