package rp

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	CommitSearchDepth = 50 // TODO: Increase
	GitRemoteName     = "origin"
)

type Commit struct {
	Hash    string
	Message string
}

type Tag struct {
	Hash string
	Name string
}

func CloneRepo(ctx context.Context, cloneURL, branch string, auth transport.AuthMethod) (*git.Repository, error) {
	dir, err := os.MkdirTemp("", "releaser-pleaser.*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for repo clone: %w", err)
	}

	// TODO: Log tmpdir
	fmt.Printf("Clone tmpdir: %s\n", dir)
	repo, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:           cloneURL,
		RemoteName:    GitRemoteName,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  false,
		Auth:          auth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return repo, nil
}
