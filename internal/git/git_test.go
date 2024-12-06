package git

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

const testMainBranch = "main"
const testPRBranch = "releaser-pleaser"

func TestRepository_HasChangesWithRemote(t *testing.T) {
	tests := []struct {
		name    string
		repo    TestRepo
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no remote pr branch",
			repo: WithTestRepo(
				WithCommit(
					"chore: release v1.0.0",
					OnBranch(plumbing.NewBranchReferenceName(testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewBranchReferenceName(testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
			),
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "remote pr branch matches local",
			repo: WithTestRepo(
				WithCommit(
					"chore: release v1.0.0",
					OnBranch(plumbing.NewBranchReferenceName(testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewBranchReferenceName(testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
			),
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "remote pr only needs rebase",
			repo: WithTestRepo(
				WithCommit(
					"chore: release v1.0.0",
					OnBranch(plumbing.NewBranchReferenceName(testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"feat: new feature on remote",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					WithFile("feature", "yes"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewBranchReferenceName(testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
			),
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "needs update",
			repo: WithTestRepo(
				WithCommit(
					"chore: release v1.0.0",
					OnBranch(plumbing.NewBranchReferenceName(testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewRemoteReferenceName(remoteName, testPRBranch)),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"chore: release v1.2.0",
					OnBranch(plumbing.NewRemoteReferenceName(remoteName, testMainBranch)),
					AsNewBranch(plumbing.NewBranchReferenceName(testPRBranch)),
					WithFile("VERSION", "v1.2.0"),
				),
			),
			want:    false,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo(t)
			got, err := repo.HasChangesWithRemote(context.Background(), testMainBranch, testPRBranch)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
