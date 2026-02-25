//go:build e2e_forgejo

package forgejo

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/test/e2e"
)

var (
	f *e2e.Framework
)

var (
	TestAuthor = git.Author{
		Name:  "Peter Parker",
		Email: "parker@example.com",
	}
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	f, err = e2e.NewFramework(ctx, &TestForge{})
	if err != nil {
		slog.Error("failed to set up test framework", "err", err)
	}

	os.Exit(m.Run())
}

func TestCreateRepository(t *testing.T) {
	_ = f.NewRepository(t, t.Name())
}

func TestEmptyRun(t *testing.T) {
	repo := f.NewRepository(t, t.Name())
	require.NoError(t, f.Run(t, repo, []string{}))
}

func TestRunMultipleSimpleReleases(t *testing.T) {
	repo := f.NewRepository(t, t.Name())

	// First release
	{
		clonedRepo := f.CloneRepo(t, repo)

		clonedRepo.UpdateFile(t.Context(), "README.md", true, func(_ string) (string, error) {
			return "# Hello World", nil
		})

		_, err := clonedRepo.Commit(t.Context(), "feat: cool new thing", TestAuthor)
		require.NoError(t, err)

		clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch)

		require.NoError(t, f.Run(t, repo, []string{}))

		pr := f.HasReleasePR(t, repo, "v0.1.0")
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}))
		f.HasTag(t, repo, "v0.1.0")
	}

	// Second release
	{
		clonedRepo := f.CloneRepo(t, repo)

		clonedRepo.UpdateFile(t.Context(), "README.md", true, func(_ string) (string, error) {
			return "# Goodbye", nil
		})

		_, err := clonedRepo.Commit(t.Context(), "fix: readme was broken", TestAuthor)
		require.NoError(t, err)

		clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch)

		require.NoError(t, f.Run(t, repo, []string{}))

		pr := f.HasReleasePR(t, repo, "v0.1.1")
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}))
		f.HasTag(t, repo, "v0.1.1")
	}

}
