package rp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](input T) *T {
	return &input
}

func Test_formatChangelog(t *testing.T) {
	type args struct {
		commits []AnalyzedCommit
		version string
		link    string
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
				commits: []AnalyzedCommit{},
				version: "1.0.0",
				link:    "https://example.com/1.0.0",
			},
			want:    "## [1.0.0](https://example.com/1.0.0)",
			wantErr: assert.NoError,
		},
		{
			name: "single feature",
			args: args{
				commits: []AnalyzedCommit{
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
				commits: []AnalyzedCommit{
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
				commits: []AnalyzedCommit{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatChangelog(tt.args.commits, tt.args.version, tt.args.link)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
