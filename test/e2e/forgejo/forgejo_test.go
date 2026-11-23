//go:build e2e_forgejo

package forgejo

import (
	"context"
	"log/slog"
	"os"
	"strings"
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

		_, err := clonedRepo.Commit(t.Context(), "feat: cool new thing", TestAuthor, nil)
		require.NoError(t, err)

		clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch)

		require.NoError(t, f.Run(t, repo, []string{}))

		pr := f.HasReleasePR(t, repo, "chore(main): release v0.1.0")
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

		_, err := clonedRepo.Commit(t.Context(), "fix: readme was broken", TestAuthor, nil)
		require.NoError(t, err)

		clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch)

		require.NoError(t, f.Run(t, repo, []string{}))

		pr := f.HasReleasePR(t, repo, "chore(main): release v0.1.1")
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}))
		f.HasTag(t, repo, "v0.1.1")
	}

}

func calVerArgs(now string) []string {
	return []string{
		"--versioning=calver",
		"--calver-format=YYYY.0W.MICRO",
		"--now=" + now,
	}
}

// pushCommit creates a new commit with the message on the default branch.
func pushCommit(t *testing.T, repo *e2e.Repository, message string) {
	t.Helper()

	clonedRepo := f.CloneRepo(t, repo)

	require.NoError(t, clonedRepo.UpdateFile(t.Context(), "README.md", true, func(content string) (string, error) {
		return content + "\n" + message, nil
	}))

	_, err := clonedRepo.Commit(t.Context(), message, TestAuthor, nil)
	require.NoError(t, err)

	require.NoError(t, clonedRepo.ForcePush(t.Context(), e2e.TestDefaultBranch))
}

func TestCalVer(t *testing.T) {
	repo := f.NewRepository(t, t.Name())

	steps := []struct {
		message     string
		now         string
		version     string
		section     string
		description string
	}{
		{message: "feat(api): add endpoint", now: "2026-11-18T12:00:00Z", version: "2026.47.0", section: "### Features", description: "**api**: add endpoint"},
		{message: "fix: typo in readme", now: "2026-11-19T12:00:00Z", version: "2026.47.1", section: "### Bug Fixes", description: "- typo in readme"},
		{message: "feat(ui): new button", now: "2026-11-25T12:00:00Z", version: "2026.48.0", section: "### Features", description: "**ui**: new button"},
	}

	for _, step := range steps {
		pushCommit(t, repo, step.message)

		require.NoError(t, f.Run(t, repo, []string{}, calVerArgs(step.now)...))

		pr := f.HasReleasePR(t, repo, "chore(main): release "+step.version)
		require.Contains(t, pr.Description, step.section)
		require.Contains(t, pr.Description, step.description)
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}, calVerArgs(step.now)...))
		f.HasTag(t, repo, step.version)
	}
}

func TestCalVerPrerelease(t *testing.T) {
	repo := f.NewRepository(t, t.Name())
	const now = "2026-11-18T12:00:00Z"

	steps := []struct {
		message string
		labels  []string
		version string
	}{
		{message: "feat(api): first rc", labels: []string{"rp-next-version::rc"}, version: "2026.47.0-rc.0"},
		{message: "feat(api): second rc", labels: []string{"rp-next-version::rc"}, version: "2026.47.0-rc.1"},
		{message: "feat(api): stable", version: "2026.47.0"},
	}

	for _, step := range steps {
		pushCommit(t, repo, step.message)

		require.NoError(t, f.Run(t, repo, []string{}, calVerArgs(now)...))

		if len(step.labels) > 0 {
			// The release PR is always opened as a stable release first, labels can only be set afterwards.
			pr := f.HasReleasePR(t, repo, "chore(main): release 2026.47.0")
			f.AddReleasePRLabels(t, repo, pr, step.labels...)

			require.NoError(t, f.Run(t, repo, []string{}, calVerArgs(now)...))
		}

		pr := f.HasReleasePR(t, repo, "chore(main): release "+step.version)
		require.Contains(t, pr.Description, strings.TrimPrefix(step.message, "feat(api): "))
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}, calVerArgs(now)...))
		f.HasTag(t, repo, step.version)
	}
}

func TestCalVerScopedCommits(t *testing.T) {
	repo := f.NewRepository(t, t.Name())

	steps := []struct {
		message     string
		now         string
		version     string
		description string
	}{
		{message: "api: add endpoint", now: "2026-11-18T12:00:00Z", version: "2026.47.0", description: "**api**: add endpoint"},
		{message: "fix typo in readme", now: "2026-11-19T12:00:00Z", version: "2026.47.1", description: "- fix typo in readme"},
		{message: "ui: new button", now: "2026-11-25T12:00:00Z", version: "2026.48.0", description: "**ui**: new button"},
	}

	for _, step := range steps {
		args := append([]string{"--commit-parser=scoped"}, calVerArgs(step.now)...)

		pushCommit(t, repo, step.message)

		require.NoError(t, f.Run(t, repo, []string{}, args...))

		pr := f.HasReleasePR(t, repo, "release: "+step.version)
		require.Contains(t, pr.Description, "### Changes")
		require.Contains(t, pr.Description, step.description)
		// Release commits of the previous releases must not show up in the changelog.
		require.NotContains(t, pr.Description, "release: 2026.")
		f.MergeReleasePR(t, repo, pr)

		require.NoError(t, f.Run(t, repo, []string{}, args...))
		f.HasTag(t, repo, step.version)
	}
}
