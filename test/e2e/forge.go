package e2e

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/apricote/releaser-pleaser/internal/git"
)

type TestForge interface {
	Init(ctx context.Context, runID string) error
	CreateRepo(t *testing.T, opts CreateRepoOpts) (*Repository, error)
	CloneURL(t *testing.T, repo *Repository) string
	GitAuth(t *testing.T) transport.AuthMethod

	ListOpenPRs(t *testing.T, repo *Repository) ([]*git.PullRequest, error)
	MergePR(t *testing.T, repo *Repository, pr *git.PullRequest) error
	ListTags(t *testing.T, repo *Repository) ([]*git.Tag, error)

	RunArguments() []string
}

type CreateRepoOpts struct {
	Name          string
	Description   string
	DefaultBranch string
}
