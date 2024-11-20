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

type Author struct {
	Name  string
	Email string
}

func (a Author) signature(when time.Time) *object.Signature {
	return &object.Signature{
		Name:  a.Name,
		Email: a.Email,
		When:  when,
	}
}

func (a Author) String() string {
	return fmt.Sprintf("%s <%s>", a.Name, a.Email)
}

var (
	committer = Author{Name: "releaser-pleaser", Email: ""}
)

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
	defer file.Close() //nolint:errcheck

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

func (r *Repository) Commit(_ context.Context, message string, author Author) (Commit, error) {
	worktree, err := r.r.Worktree()
	if err != nil {
		return Commit{}, err
	}

	now := time.Now()

	releaseCommitHash, err := worktree.Commit(message, &git.CommitOptions{
		Author:    author.signature(now),
		Committer: committer.signature(now),
	})
	if err != nil {
		return Commit{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	return Commit{
		Hash:    releaseCommitHash.String(),
		Message: message,
	}, nil
}

// HasChangesWithRemote checks if the following two diffs are equal:
//
// - **Local**:                                 remote/main..branch
// - **Remote**: (git merge-base remote/main remote/branch)..remote/branch
//
// This is done to avoid pushing when the only change would be a rebase of remote/branch onto the current remote/main.
func (r *Repository) HasChangesWithRemote(ctx context.Context, mainBranch, prBranch string) (bool, error) {
	return r.hasChangesWithRemote(ctx,
		plumbing.NewRemoteReferenceName(remoteName, mainBranch),
		plumbing.NewBranchReferenceName(prBranch),
		plumbing.NewRemoteReferenceName(remoteName, prBranch),
	)
}

func (r *Repository) hasChangesWithRemote(ctx context.Context, mainBranchRef, localPRBranchRef, remotePRBranchRef plumbing.ReferenceName) (bool, error) {
	commitOnRemoteMain, err := r.commitFromRef(mainBranchRef)
	if err != nil {
		return false, err
	}

	commitOnRemotePRBranch, err := r.commitFromRef(remotePRBranchRef)
	if err != nil {
		if err.Error() == "reference not found" {
			// No remote branch means that there are changes
			return true, nil
		}

		return false, err
	}

	currentRemotePRMergeBase, err := r.mergeBase(commitOnRemoteMain, commitOnRemotePRBranch)
	if err != nil {
		return false, err
	}
	if currentRemotePRMergeBase == nil {
		// If there is no merge base something weird has happened with the
		// remote main branch, and we should definitely push updates.
		return false, nil
	}

	remoteDiff, err := commitOnRemotePRBranch.PatchContext(ctx, currentRemotePRMergeBase)
	if err != nil {
		return false, err
	}

	commitOnLocalPRBranch, err := r.commitFromRef(localPRBranchRef)
	if err != nil {
		return false, err
	}

	localDiff, err := commitOnRemoteMain.PatchContext(ctx, commitOnLocalPRBranch)
	if err != nil {
		return false, err
	}

	return remoteDiff.String() == localDiff.String(), nil
}

func (r *Repository) commitFromRef(refName plumbing.ReferenceName) (*object.Commit, error) {
	ref, err := r.r.Reference(refName, false)
	if err != nil {
		return nil, err
	}

	commit, err := r.r.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (r *Repository) mergeBase(a, b *object.Commit) (*object.Commit, error) {
	mergeBases, err := a.MergeBase(b)
	if err != nil {
		return nil, err
	}

	if len(mergeBases) == 0 {
		return nil, nil
	}

	// :shrug: We dont really care which commit we pick, at worst we do an unnecessary push.
	return mergeBases[0], nil
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
