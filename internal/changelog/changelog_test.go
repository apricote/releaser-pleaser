package changelog

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/testdata"
)

func Test_NewChangelogEntry(t *testing.T) {
	type args struct {
		analyzedCommits []commitparser.AnalyzedCommit
		version         string
		link            string
		compare         string
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
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:        "feat",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Features\n\n- Foobar! ([abc1234](https://example.com/commit/abc1234567890))\n",
			wantErr: assert.NoError,
		},
		{
			name: "single breaking change",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:           "feat",
						Description:    "Foobar!",
						BreakingChange: true,
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Features\n\n- **BREAKING**: Foobar! ([abc1234](https://example.com/commit/abc1234567890))\n",
			wantErr: assert.NoError,
		},
		{
			name: "single fix",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)\n\n### Bug Fixes\n\n- Foobar! ([abc1234](https://example.com/commit/abc1234567890))\n",
			wantErr: assert.NoError,
		},
		{
			name: "changes with and without scope",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "aaa1111111111", URL: "https://example.com/commit/aaa1111111111",
						Type:        "change",
						Scope:       new("api"),
						Description: "add endpoint",
					},
					{
						Hash: "bbb2222222222", URL: "https://example.com/commit/bbb2222222222",
						Type:        "change",
						Description: "fix typo in readme",
					},
				},
				version: "2026.47.0",
				link:    "https://example.com/2026.47.0",
			},
			want:    "## [2026.47.0](https://example.com/2026.47.0)\n\n### Changes\n\n- **api**: add endpoint ([aaa1111](https://example.com/commit/aaa1111111111))\n- fix typo in readme ([bbb2222](https://example.com/commit/bbb2222222222))\n",
			wantErr: assert.NoError,
		},
		{
			name: "multiple commits with scopes",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "aaa1111111111", URL: "https://example.com/commit/aaa1111111111",
						Type:        "feat",
						Description: "Blabla!",
					},
					{
						Hash: "bbb2222222222", URL: "https://example.com/commit/bbb2222222222",
						Type:        "feat",
						Description: "So awesome!",
						Scope:       new("awesome"),
					},
					{
						Hash: "ccc3333333333", URL: "https://example.com/commit/ccc3333333333",
						Type:        "fix",
						Description: "Foobar!",
					},
					{
						Hash: "ddd4444444444", URL: "https://example.com/commit/ddd4444444444",
						Type:        "fix",
						Description: "So sad!",
						Scope:       new("sad"),
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want: `## [1.0.0](https://example.com/1.0.0)

### Features

- Blabla! ([aaa1111](https://example.com/commit/aaa1111111111))
- **awesome**: So awesome! ([bbb2222](https://example.com/commit/bbb2222222222))

### Bug Fixes

- Foobar! ([ccc3333](https://example.com/commit/ccc3333333333))
- **sad**: So sad! ([ddd4444](https://example.com/commit/ddd4444444444))
`,
			wantErr: assert.NoError,
		},
		{
			name: "compare url",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
				compare: "https://example.com/compare/0.1.0/1.0.0/",
			},
			want:    testdata.MustReadFileString(t, "changelog-entry-compare-url.txt"),
			wantErr: assert.NoError,
		},
		{
			name: "prefix",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
				prefix:  testdata.MustReadFileString(t, "prefix.txt"),
			},
			want:    testdata.MustReadFileString(t, "changelog-entry-prefix.txt"),
			wantErr: assert.NoError,
		},
		{
			name: "suffix",
			args: args{
				analyzedCommits: []commitparser.AnalyzedCommit{
					{
						Hash: "abc1234567890", URL: "https://example.com/commit/abc1234567890",
						Type:        "fix",
						Description: "Foobar!",
					},
				},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
				suffix:  testdata.MustReadFileString(t, "suffix.txt"),
			},
			want:    testdata.MustReadFileString(t, "changelog-entry-suffix.txt"),
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := New(commitparser.ByType(tt.args.analyzedCommits), tt.args.version, tt.args.link, tt.args.compare, tt.args.prefix, tt.args.suffix)
			got, err := Entry(slog.Default(), DefaultTemplate(), data, Formatting{})
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
