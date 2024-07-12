package rp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/testutils"
)

var InitCommit = Commit{
	Hash:    "fca23c2d67580780d5bbb4f73987624b81926aad",
	Message: "chore: init",
}

func TestReleasableCommits(t *testing.T) {
	tests := []struct {
		name        string
		repoFn      testutils.Repo
		wantCommits []Commit
		wantTag     *Tag
		wantErr     bool
	}{
		{
			name:   "Empty Repo",
			repoFn: testutils.WithTestRepo(),
			wantCommits: []Commit{
				InitCommit,
			},
			wantTag: nil,
			wantErr: false,
		},
		{
			name: "Single Commit",
			repoFn: testutils.WithTestRepo(
				testutils.WithCommit("feat: foobar"),
			),
			wantCommits: []Commit{
				InitCommit,
				{
					Hash:    "ff0815947e8211485d4f97ff8cf5deb49866e228",
					Message: "feat: foobar",
				},
			},
			wantTag: nil,
			wantErr: false,
		},
		{
			name: "Single Commit After Tag",
			repoFn: testutils.WithTestRepo(
				testutils.WithCommit("feat: foobar", testutils.WithTag("v1.0.0")),
				testutils.WithCommit("feat: baz"),
			),
			wantCommits: []Commit{
				{
					Hash:    "ccdc724ef1755095d5a58c2421eec75d4010b3b7",
					Message: "feat: baz",
				},
			},
			wantTag: &Tag{
				Hash: "ff0815947e8211485d4f97ff8cf5deb49866e228",
				Name: "v1.0.0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repoFn(t)

			commits, tag, err := ReleasableCommits(repo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCommits, commits)
				assert.Equal(t, tt.wantTag, tag)
			}
		})
	}
}
