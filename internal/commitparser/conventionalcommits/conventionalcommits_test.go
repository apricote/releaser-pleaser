package conventionalcommits

import (
	"log/slog"
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
			name: "skips malformed commit message",
			commits: []git.Commit{
				{
					Message: "aksdjaklsdjka",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{},
			wantErr:         assert.NoError,
		},
		{
			// GitLab seems to create commits with pattern "scope: message\n" if no body is added.
			// This has previously caused a parser error "missing a blank line".
			// We added a workaround with `strings.TrimSpace()` and this test make sure that it does not break again.
			name: "handles title with new line",
			commits: []git.Commit{
				{
					Message: "aksdjaklsdjka",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{},
			wantErr:         assert.NoError,
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

		{
			name: "success with body",
			commits: []git.Commit{
				{
					Message: "feat: some thing (hz/fl!144)\n\nFixes #15\n\nDepends on !143",
				},
			},
			expectedCommits: []commitparser.AnalyzedCommit{
				{
					Commit:         git.Commit{Message: "feat: some thing (hz/fl!144)\n\nFixes #15\n\nDepends on !143"},
					Type:           "feat",
					Description:    "some thing (hz/fl!144)",
					BreakingChange: false,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzedCommits, err := NewParser(slog.Default()).Analyze(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.expectedCommits, analyzedCommits)
		})
	}
}
