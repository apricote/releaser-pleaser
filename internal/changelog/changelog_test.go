package changelog

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

func ptr[T any](input T) *T {
	return &input
}

func Test_NewChangelogEntry(t *testing.T) {
	type args struct {
		analyzedCommits []commitparser.AnalyzedCommit
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
				analyzedCommits: []commitparser.AnalyzedCommit{},
				version:         "1.0.0",
				link:            "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n",
			wantErr: assert.NoError,
		},
		{
			name: "single feature",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:      git.Commit{},
						Type:        "feat",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Features\n\n- Foobar!\n",
			wantErr: assert.NoError,
		},
		{
			name: "single breaking change",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:         git.Commit{},
						Type:           "feat",
						Description:    "Foobar!",
						BreakingChange: true,
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Features\n\n- **BREAKING**: Foobar!\n",
			wantErr: assert.NoError,
		},
		{
			name: "single fix",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:      git.Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Bug Fixes\n\n- Foobar!\n",
			wantErr: assert.NoError,
		},
		{
			name: "multiple commits with scopes",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:      git.Commit{},
						Type:        "feat",
						Description: "Blabla!",
					},
					{
						Commit:      git.Commit{},
						Type:        "feat",
						Description: "So awesome!",
						Scope:       ptr("awesome"),
					},
					{
						Commit:      git.Commit{},
						Type:        "fix",
						Description: "Foobar!",
					},
					{
						Commit:      git.Commit{},
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
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:      git.Commit{},
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
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Commit:      git.Commit{},
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
			data := New(commitparser.ByType(tt.args.analyzedCommits), tt.args.version, tt.args.link, tt.args.prefix, tt.args.suffix)
			got, err := Entry(slog.Default(), DefaultTemplate(), data, Formatting{})
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
