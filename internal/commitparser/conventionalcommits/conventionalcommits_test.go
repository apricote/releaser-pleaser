package conventionalcommits

import (
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/versioning"
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
			analyzedCommits, err := NewParser(slog.Default(), nil).Analyze(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.expectedCommits, analyzedCommits)
		})
	}
}

func TestAnalyzeExtraPatchTypes(t *testing.T) {
	for _, commitType := range []string{"build", "ci", "chore", "docs", "perf", "refactor", "revert", "style", "test"} {
		t.Run(commitType, func(t *testing.T) {
			commit := git.Commit{Message: commitType + "(api): improve things"}
			commits, err := NewParser(slog.Default(), nil).Analyze([]git.Commit{commit})
			require.NoError(t, err)
			assert.Empty(t, commits)

			extraPatchTypes := regexp.MustCompile("^" + commitType + "$")
			commits, err = NewParser(slog.Default(), extraPatchTypes).Analyze([]git.Commit{commit, {Message: "fix: fix things"}})
			require.NoError(t, err)
			require.Len(t, commits, 2)
			assert.Equal(t, commitType, commits[0].Type)
			assert.Equal(t, "improve things", commits[0].Description)
			require.NotNil(t, commits[0].Scope)
			assert.Equal(t, "api", *commits[0].Scope)
			assert.Equal(t, versioning.PatchVersion, versioning.BumpFromCommits(commits[:1], extraPatchTypes))
		})
	}
}

func TestAnalyzeExtraPatchTypesFiltering(t *testing.T) {
	commits, err := NewParser(slog.Default(), regexp.MustCompile("^(?:docs|feat)$")).Analyze([]git.Commit{
		{Message: "chore: still ignored"},
		{Message: "not a conventional commit"},
		{Message: "feat: still a feature"},
		{Message: "docs!: breaking documentation"},
		{Message: "refactor: change API\n\nBREAKING CHANGE: removed API"},
	})
	require.NoError(t, err)
	require.Len(t, commits, 3)
	assert.Equal(t, "feat", commits[0].Type)
	assert.Equal(t, versioning.MinorVersion, versioning.BumpFromCommit(commits[0], regexp.MustCompile(".*")))
	assert.True(t, commits[1].BreakingChange)
	assert.True(t, commits[2].BreakingChange)
}
