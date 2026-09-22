//go:build e2e_forgejo

package forgejo

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestRunExtraPatchTypes(t *testing.T) {
	forge := &TestForge{}
	framework, err := e2e.NewFramework(t.Context(), forge)
	require.NoError(t, err)
	repo := framework.NewRepository(t, t.Name())
	clonedRepo := framework.CloneRepo(t, repo)

	for _, message := range []string{"docs: document usage", "chore: tidy files", "test: add coverage"} {
		err = clonedRepo.UpdateFile(t.Context(), "README.md", false, func(content string) (string, error) {
			return content + "\n" + message, nil
		})
		require.NoError(t, err)
		_, err = clonedRepo.Commit(t.Context(), message, TestAuthor)
		require.NoError(t, err)
	}
	require.NoError(t, clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch))

	require.NoError(t, framework.Run(t, repo, nil))
	prs, err := forge.ListOpenPRs(t, repo)
	require.NoError(t, err)
	assert.Empty(t, prs)

	const extraPatchTypes = "--extra-patch-types=docs|chore"
	require.NoError(t, framework.Run(t, repo, nil, extraPatchTypes))
	pr := framework.HasReleasePR(t, repo, "v0.0.1")
	assert.Contains(t, pr.Description, "### Other")
	assert.Contains(t, pr.Description, "document usage")
	assert.Contains(t, pr.Description, "tidy files")
	assert.NotContains(t, pr.Description, "add coverage")

	require.NoError(t, framework.Run(t, repo, nil, extraPatchTypes))
	assert.Equal(t, pr.ID, framework.HasReleasePR(t, repo, "v0.0.1").ID)
	framework.MergeReleasePR(t, repo, pr)
	require.NoError(t, framework.Run(t, repo, nil, extraPatchTypes))
	framework.HasTag(t, repo, "v0.0.1")
	prs, err = forge.ListOpenPRs(t, repo)
	require.NoError(t, err)
	assert.Empty(t, prs)
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
