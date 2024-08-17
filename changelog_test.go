package rp

import (
	"io"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/testutils"
)

func ptr[T any](input T) *T {
	return &input
}

func TestUpdateChangelogFile(t *testing.T) {
	tests := []struct {
		name            string
		repoFn          testutils.Repo
		entry           string
		expectedContent string
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "empty repo",
			repoFn:          testutils.WithTestRepo(),
			entry:           "## v1.0.0\n",
			expectedContent: "# Changelog\n\n## v1.0.0\n",
			wantErr:         assert.NoError,
		},
		{
			name: "repo with well-formatted changelog",
			repoFn: testutils.WithTestRepo(testutils.WithCommit("feat: add changelog", testutils.WithFile(ChangelogFile, `# Changelog

## v0.0.1

- Bazzle

## v0.1.0

### Bazuuum
`))),
			entry: "## v1.0.0\n\n- Version 1, juhu.\n",
			expectedContent: `# Changelog

## v1.0.0

- Version 1, juhu.

## v0.0.1

- Bazzle

## v0.1.0

### Bazuuum
`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repoFn(t)
			wt, err := repo.Worktree()
			require.NoError(t, err, "failed to get worktree")

			err = UpdateChangelogFile(wt, tt.entry)
			if !tt.wantErr(t, err) {
				return
			}

			wtStatus, err := wt.Status()
			require.NoError(t, err, "failed to get worktree status")

			assert.Len(t, wtStatus, 1, "worktree status does not have the expected entry number")

			changelogFileStatus := wtStatus.File(ChangelogFile)

			assert.Equal(t, git.Unmodified, changelogFileStatus.Worktree, "unexpected file status in worktree")
			assert.Equal(t, git.Added, changelogFileStatus.Staging, "unexpected file status in staging")

			changelogFile, err := wt.Filesystem.Open(ChangelogFile)
			require.NoError(t, err)
			defer changelogFile.Close()

			changelogFileContent, err := io.ReadAll(changelogFile)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedContent, string(changelogFileContent))
		})
	}
}

func Test_NewChangelogEntry(t *testing.T) {
	type args struct {
		analyzedCommits []AnalyzedCommit
		version         string
		link            string
		prefix          string
		suffix          string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty",
			args: args{
				analyzedCommits: []AnalyzedCommit{},
				version:         "1.0.0",
				link:            "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)",
			wantErr: assert.NoError,
		},
		{
			name: "single feature",
			args: args{
				analyzedCommits: []AnalyzedCommit{
					{
						Commit:      Commit{},
						Type:        "feat",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n### Features\n\n- Foobar!\n",
			wantErr: assert.NoError,
		},
		{
			name: "single fix",
			args: args{
				analyzedCommits: []AnalyzedCommit{
					{
						Commit:      Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n### Bug Fixes\n\n- Foobar!\n",
			wantErr: assert.NoError,
		},
		{
			name: "multiple commits with scopes",
			args: args{
				analyzedCommits: []AnalyzedCommit{
					{
						Commit:      Commit{},
						Type:        "feat",
						Description: "Blabla!",
					},
					{
						Commit:      Commit{},
						Type:        "feat",
						Description: "So awesome!",
						Scope:       ptr("awesome"),
					},
					{
						Commit:      Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
					{
						Commit:      Commit{},
						Type:        "fix",
						Description: "So sad!",
						Scope:       ptr("sad"),
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want: `## [1.0.0](https://example.com/1.0.0)
### Features

- Blabla!
- **awesome**: So awesome!

### Bug Fixes

- Foobar!
- **sad**: So sad!
`,
			wantErr: assert.NoError,
		},
		{
			name: "prefix",
			args: args{
				analyzedCommits: []AnalyzedCommit{
					{
						Commit:      Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
				prefix:  "### Breaking Changes",
			},
			want: `## [1.0.0](https://example.com/1.0.0)
### Breaking Changes

### Bug Fixes

- Foobar!
`,
			wantErr: assert.NoError,
		},
		{
			name: "suffix",
			args: args{
				analyzedCommits: []AnalyzedCommit{
					{
						Commit:      Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
				suffix:  "### Compatibility\n\nThis version is compatible with flux-compensator v2.2 - v2.9.",
			},
			want: `## [1.0.0](https://example.com/1.0.0)
### Bug Fixes

- Foobar!

### Compatibility

This version is compatible with flux-compensator v2.2 - v2.9.
`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewChangelogEntry(tt.args.analyzedCommits, tt.args.version, tt.args.link, tt.args.prefix, tt.args.suffix)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
