package testutils

import (
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

var author = &object.Signature{
	Name: "releaser-pleaser",
	When: time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC),
}

type CommitOption func(*commitOptions)

type commitOptions struct {
	cleanFiles bool
	files      []commitFile
	tags       []string
}

type commitFile struct {
	path    string
	content string
}

type Commit func(*testing.T, *git.Repository) error

type Repo func(*testing.T) *git.Repository

func WithCommit(message string, options ...CommitOption) Commit {
	return func(t *testing.T, repo *git.Repository) error {
		t.Helper()

		require.NotEmpty(t, message, "commit message is required")

		opts := &commitOptions{}
		for _, opt := range options {
			opt(opts)
		}

		wt, err := repo.Worktree()
		require.NoError(t, err)

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
			file.Close()
			require.NoError(t, err, "failed to write content to file %q", fileInfo.path)
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
			_, err = repo.CreateTag(tagName, commitHash, nil)
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

func WithCleanFiles() CommitOption {
	return func(opts *commitOptions) {
		opts.cleanFiles = true
	}
}

func WithTag(name string) CommitOption {
	return func(opts *commitOptions) {
		opts.tags = append(opts.tags, name)
	}
}

func WithTestRepo(commits ...Commit) Repo {
	return func(t *testing.T) *git.Repository {
		t.Helper()

		repo, err := git.Init(memory.NewStorage(), memfs.New())
		require.NoError(t, err, "failed to create in-memory repository")

		// Make initial commit
		err = WithCommit("chore: init")(t, repo)
		require.NoError(t, err, "failed to create init commit")

		for i, commit := range commits {
			err = commit(t, repo)
			require.NoError(t, err, "failed to create commit %d", i)
		}

		return repo
	}
}
