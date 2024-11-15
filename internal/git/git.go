package git

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/apricote/releaser-pleaser/internal/updater"
)

const (
	remoteName         = "origin"
	newFilePermissions = 0o644
)

type Commit struct {
	Hash    string
	Message string

	PullRequest *PullRequest
}

type PullRequest struct {
	ID          int
	Title       string
	Description string
}

type Tag struct {
	Hash string
	Name string
}

type Releases struct {
	Latest *Tag
	Stable *Tag
}

func CloneRepo(ctx context.Context, logger *slog.Logger, cloneURL, branch string, auth transport.AuthMethod) (*Repository, error) {
	dir, err := os.MkdirTemp("", "releaser-pleaser.*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for repo clone: %w", err)
	}

	repo, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:           cloneURL,
		RemoteName:    remoteName,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  false,
		Auth:          auth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return &Repository{r: repo, logger: logger, auth: auth}, nil
}

type Repository struct {
	r      *git.Repository
	logger *slog.Logger
	auth   transport.AuthMethod
}

func (r *Repository) DeleteBranch(ctx context.Context, branch string) error {
	if b, _ := r.r.Branch(branch); b != nil {
		r.logger.DebugContext(ctx, "deleting local branch", "branch.name", branch)
		if err := r.r.DeleteBranch(branch); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) Checkout(_ context.Context, branch string) error {
	worktree, err := r.r.Worktree()
	if err != nil {
		return err
	}

	if err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	}); err != nil {
		return fmt.Errorf("failed to check out branch: %w", err)
	}

	return nil
}

func (r *Repository) UpdateFile(_ context.Context, path string, create bool, updaters []updater.Updater) error {
	worktree, err := r.r.Worktree()
	if err != nil {
		return err
	}

	fileFlags := os.O_RDWR
	if create {
		fileFlags |= os.O_CREATE
	}

	file, err := worktree.Filesystem.OpenFile(path, fileFlags, newFilePermissions)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	updatedContent := string(content)

	for _, update := range updaters {
		updatedContent, err = update(updatedContent)
		if err != nil {
			return fmt.Errorf("failed to run updater on file %s", path)
		}
	}

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to replace file content: %w", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to replace file content: %w", err)
	}
	_, err = file.Write([]byte(updatedContent))
	if err != nil {
		return fmt.Errorf("failed to replace file content: %w", err)
	}

	_, err = worktree.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add updated file to git worktree: %w", err)
	}

	return nil
}

func (r *Repository) Commit(_ context.Context, message string) (Commit, error) {
	worktree, err := r.r.Worktree()
	if err != nil {
		return Commit{}, err
	}

	releaseCommitHash, err := worktree.Commit(message, &git.CommitOptions{
		Author:    signature(),
		Committer: signature(),
	})
	if err != nil {
		return Commit{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	return Commit{
		Hash:    releaseCommitHash.String(),
		Message: message,
	}, nil
}

func (r *Repository) HasChangesWithRemote(ctx context.Context, branch string) (bool, error) {
	remoteRef, err := r.r.Reference(plumbing.NewRemoteReferenceName(remoteName, branch), false)
	if err != nil {
		if err.Error() == "reference not found" {
			// No remote branch means that there are changes
			return true, nil
		}

		return false, err
	}

	remoteCommit, err := r.r.CommitObject(remoteRef.Hash())
	if err != nil {
		return false, err
	}

	localRef, err := r.r.Reference(plumbing.NewBranchReferenceName(branch), false)
	if err != nil {
		return false, err
	}

	localCommit, err := r.r.CommitObject(localRef.Hash())
	if err != nil {
		return false, err
	}

	diff, err := localCommit.PatchContext(ctx, remoteCommit)
	if err != nil {
		return false, err
	}

	hasChanges := len(diff.FilePatches()) > 0

	return hasChanges, nil
}

func (r *Repository) ForcePush(ctx context.Context, branch string) error {
	pushRefSpec := config.RefSpec(fmt.Sprintf(
		"+%s:%s",
		plumbing.NewBranchReferenceName(branch),
		// This needs to be the local branch name, not the remotes/origin ref
		// See https://stackoverflow.com/a/75727620
		plumbing.NewBranchReferenceName(branch),
	))

	r.logger.DebugContext(ctx, "pushing branch", "branch.name", branch, "refspec", pushRefSpec.String())
	return r.r.PushContext(ctx, &git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{pushRefSpec},
		Force:      true,
		Auth:       r.auth,
	})
}

func signature() *object.Signature {
	return &object.Signature{
		Name:  "releaser-pleaser",
		Email: "",
		When:  time.Now(),
	}
}
