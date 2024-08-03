package rp

import (
	"testing"

	"github.com/leodido/go-conventionalcommits"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeCommits(t *testing.T) {
	tests := []struct {
		name            string
		commits         []Commit
		expectedCommits []AnalyzedCommit
		expectedBump    conventionalcommits.VersionBump
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "empty commits",
			commits:         []Commit{},
			expectedCommits: []AnalyzedCommit{},
			expectedBump:    conventionalcommits.UnknownVersion,
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
			expectedBump:    conventionalcommits.UnknownVersion,
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
			expectedBump:    conventionalcommits.UnknownVersion,
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
			expectedBump: conventionalcommits.PatchVersion,
			wantErr:      assert.NoError,
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
			expectedBump: conventionalcommits.MinorVersion,
			wantErr:      assert.NoError,
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
			expectedBump: conventionalcommits.MajorVersion,
			wantErr:      assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzedCommits, versionBump, err := AnalyzeCommits(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.expectedCommits, analyzedCommits)
			assert.Equal(t, tt.expectedBump, versionBump)
		})
	}
}
