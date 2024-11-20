package git

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestAuthor_signature(t *testing.T) {
	now := time.Now()

	tests := []struct {
		author Author
		want   *object.Signature
	}{
		{author: Author{Name: "foo", Email: "bar@example.com"}, want: &object.Signature{Name: "foo", Email: "bar@example.com", When: now}},
		{author: Author{Name: "bar", Email: "foo@example.com"}, want: &object.Signature{Name: "bar", Email: "foo@example.com", When: now}},
	}
	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			if got := tt.author.signature(now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("signature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthor_String(t *testing.T) {
	tests := []struct {
		author Author
		want   string
	}{
		{author: Author{Name: "foo", Email: "bar@example.com"}, want: "foo <bar@example.com>"},
		{author: Author{Name: "bar", Email: "foo@example.com"}, want: "bar <foo@example.com>"},
	}
	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			if got := tt.author.String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

const testMainBranch = "main"
const testPRBranch = "releaser-pleaser"

func TestRepository_HasChangesWithRemote(t *testing.T) {
	// go-git/v5 has a bug where it tries to delete the repo root dir (".") multiple times if there is no file left in it.
	// this happens while switching branches in worktree.go rmFileAndDirsIfEmpty.
	// TODO: Fix bug upstream
	// For now I just make sure that there is always at least one file left in the dir by adding an empty "README.md" in the test util.

	mainBranchRef := plumbing.NewBranchReferenceName(testMainBranch)
	localPRBranchRef := plumbing.NewBranchReferenceName(testPRBranch)
	remotePRBranchRef := plumbing.NewBranchReferenceName("remote/" + testPRBranch)

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
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(localPRBranchRef),
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
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(remotePRBranchRef),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(localPRBranchRef),
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
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(remotePRBranchRef),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"feat: new feature on remote",
					OnBranch(mainBranchRef),
					WithFile("feature", "yes"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(localPRBranchRef),
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
					WithFile("VERSION", "v1.0.0"),
				),
				WithCommit(
					"chore: release v1.1.0",
					OnBranch(mainBranchRef),
					AsNewBranch(remotePRBranchRef),
					WithFile("VERSION", "v1.1.0"),
				),
				WithCommit(
					"chore: release v1.2.0",
					OnBranch(mainBranchRef),
					AsNewBranch(localPRBranchRef),
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
			got, err := repo.hasChangesWithRemote(context.Background(), mainBranchRef, localPRBranchRef, remotePRBranchRef)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
