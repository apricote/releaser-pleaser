package git

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

var (
	author = &object.Signature{
		Name: "releaser-pleaser",
		When: time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC),
	}
)

type CommitOption func(*commitOptions)
type commitOptions struct {
	cleanFiles bool
	files      []commitFile
	tags       []string
	newRef     plumbing.ReferenceName
	parentRef  plumbing.ReferenceName
}
type commitFile struct {
	path    string
	content string
}

type TestCommit func(*testing.T, *Repository) error
type TestRepo func(*testing.T) *Repository

func WithCommit(message string, options ...CommitOption) TestCommit {
	return func(t *testing.T, repo *Repository) error {
		t.Helper()

		require.NotEmpty(t, message, "commit message is required")

		opts := &commitOptions{}
		for _, opt := range options {
			opt(opts)
		}

		wt, err := repo.r.Worktree()
		require.NoError(t, err)

		if opts.parentRef != "" {
			checkoutOptions := &git.CheckoutOptions{}

			if opts.newRef != "" {
				parentRef, err := repo.r.Reference(opts.parentRef, false)
				require.NoError(t, err)

				checkoutOptions.Create = true
				checkoutOptions.Hash = parentRef.Hash()
				checkoutOptions.Branch = opts.newRef
			} else {
				checkoutOptions.Branch = opts.parentRef
			}

			err = wt.Checkout(checkoutOptions)
			require.NoError(t, err)
		}

		// Yeet all files
		if opts.cleanFiles {
			files, err := wt.Filesystem.ReadDir(".")
			require.NoError(t, err, "failed to get current files")

			for _, fileInfo := range files {
				err = wt.Filesystem.Remove(fileInfo.Name())
				require.NoError(t, err, "failed to remove file %q", fileInfo.Name())
			}
		}

		// Create new files
		for _, fileInfo := range opts.files {
			file, err := wt.Filesystem.Create(fileInfo.path)
			require.NoError(t, err, "failed to create file %q", fileInfo.path)

			_, err = file.Write([]byte(fileInfo.content))
			_ = file.Close()
			require.NoError(t, err, "failed to write content to file %q", fileInfo.path)

			_, err = wt.Add(fileInfo.path)
			require.NoError(t, err, "failed to stage changes to file %q", fileInfo.path)

		}

		// Commit
		commitHash, err := wt.Commit(message, &git.CommitOptions{
			All:               true,
			AllowEmptyCommits: true,
			Author:            author,
			Committer:         author,
		})
		require.NoError(t, err, "failed to commit")

		// Create tags
		for _, tagName := range opts.tags {
			_, err = repo.r.CreateTag(tagName, commitHash, nil)
			require.NoError(t, err, "failed to create tag %q", tagName)
		}

		return nil
	}
}

func WithFile(path, content string) CommitOption {
	return func(opts *commitOptions) {
		opts.files = append(opts.files, commitFile{path: path, content: content})
	}
}

// WithCleanFiles removes all previous files from the repo. Make sure to leave at least one file in the root
// directory when switching branches!
func WithCleanFiles() CommitOption {
	return func(opts *commitOptions) {
		opts.cleanFiles = true
	}
}

func AsNewBranch(ref plumbing.ReferenceName) CommitOption {
	return func(opts *commitOptions) {
		opts.newRef = ref
	}
}

func OnBranch(ref plumbing.ReferenceName) CommitOption {
	return func(opts *commitOptions) {
		opts.parentRef = ref
	}
}

func WithTag(name string) CommitOption {
	return func(opts *commitOptions) {
		opts.tags = append(opts.tags, name)
	}
}

// Can be useful to debug git issues by using it in a terminal
const useOnDiskTestRepository = false

func WithTestRepo(commits ...TestCommit) TestRepo {
	return func(t *testing.T) *Repository {
		t.Helper()

		repo := &Repository{
			logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}

		var err error

		initOptions := git.InitOptions{DefaultBranch: plumbing.Main}

		if useOnDiskTestRepository {
			dir, err := os.MkdirTemp(os.TempDir(), "rp-test-repo-")
			require.NoError(t, err, "failed to create temp directory")

			repo.r, err = git.PlainInitWithOptions(dir, &git.PlainInitOptions{InitOptions: initOptions})
			require.NoError(t, err, "failed to create fs repository")

			fmt.Printf("using temp directory: %s", dir)
		} else {
			repo.r, err = git.InitWithOptions(memory.NewStorage(), memfs.New(), initOptions)
			require.NoError(t, err, "failed to create in-memory repository")
		}

		// Make initial commit
		err = WithCommit("chore: init", WithFile("README.md", "# git test util"))(t, repo)
		require.NoError(t, err, "failed to create init commit")

		for i, commit := range commits {
			err = commit(t, repo)
			require.NoError(t, err, "failed to create commit %d", i)
		}

		return repo
	}
}
