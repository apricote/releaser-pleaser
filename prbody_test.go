package rp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/git"
)

func Test_parsePRBodyForCommitOverrides(t *testing.T) {
	tests := []struct {
		name    string
		commits []git.Commit
		want    []git.Commit
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "no commits",
			commits: []git.Commit{},
			want:    []git.Commit{},
			wantErr: assert.NoError,
		},
		{
			name: "single commit",
			commits: []git.Commit{
				{
					Hash:    "123",
					Message: "321",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n",
					},
				},
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "321",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "multiple commits",
			commits: []git.Commit{
				{
					Hash:    "123",
					Message: "321",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
					},
				},
				{
					Hash:    "456",
					Message: "654",
					PullRequest: &git.PullRequest{
						ID:          2,
						Title:       "Bar",
						Description: "# Foobazzle\n\n",
					},
				},
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "feat: shiny",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
					},
				},
				{
					Hash:    "123",
					Message: "fix: boom",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
					},
				},
				{
					Hash:    "456",
					Message: "654",
					PullRequest: &git.PullRequest{
						ID:          2,
						Title:       "Bar",
						Description: "# Foobazzle\n\n",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRBodyForCommitOverrides(tt.commits)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseSinglePRBodyForCommitOverrides(t *testing.T) {
	tests := []struct {
		name    string
		commit  git.Commit
		want    []git.Commit
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "same commit if no PR is available",
			commit: git.Commit{
				Hash:        "123",
				Message:     "321",
				PullRequest: nil,
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "321",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "same commit if no overrides are defined",
			commit: git.Commit{
				Hash:    "123",
				Message: "321",
				PullRequest: &git.PullRequest{
					ID:          1,
					Title:       "Foo",
					Description: "# Cool new thingy\n\n",
				},
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "321",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "no commit if override is defined but empty",
			commit: git.Commit{
				Hash:    "123",
				Message: "321",
				PullRequest: &git.PullRequest{
					ID:          1,
					Title:       "Foo",
					Description: "```rp-commits\n```\n",
				},
			},
			want:    []git.Commit{},
			wantErr: assert.NoError,
		},
		{
			name: "commit messages from override",
			commit: git.Commit{
				Hash:    "123",
				Message: "321",
				PullRequest: &git.PullRequest{
					ID:          1,
					Title:       "Foo",
					Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
				},
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "feat: shiny",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
					},
				},
				{
					Hash:    "123",
					Message: "fix: boom",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\nfeat: shiny\nfix: boom\n```\n",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "ignore empty lines",
			commit: git.Commit{
				Hash:    "123",
				Message: "321",
				PullRequest: &git.PullRequest{
					ID:          1,
					Title:       "Foo",
					Description: "# Cool new thingy\n\n```rp-commits\n\n       \nfeat: shiny\n\n```\n",
				},
			},
			want: []git.Commit{
				{
					Hash:    "123",
					Message: "feat: shiny",
					PullRequest: &git.PullRequest{
						ID:          1,
						Title:       "Foo",
						Description: "# Cool new thingy\n\n```rp-commits\n\n       \nfeat: shiny\n\n```\n",
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSinglePRBodyForCommitOverrides(tt.commit)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
