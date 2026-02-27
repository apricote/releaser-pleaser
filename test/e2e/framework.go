package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/cmd/rp/cmd"
	"github.com/apricote/releaser-pleaser/internal/git"
)

const (
	TestDefaultBranch = "main"
)

func randomString() string {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(randomBytes)
}

type Framework struct {
	runID  string
	forge  TestForge
	logger *slog.Logger
}

func NewFramework(ctx context.Context, forge TestForge) (*Framework, error) {
	f := &Framework{
		runID:  randomString(),
		forge:  forge,
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	err := forge.Init(ctx, f.runID)
	if err != nil {
		return nil, err
	}

	return f, nil
}

type Repository struct {
	Name string
}

func (f *Framework) NewRepository(t *testing.T, name string) *Repository {
	t.Helper()

	r := &Repository{
		Name: fmt.Sprintf("%s-%s-%s", name, f.runID, randomString()),
	}

	repo, err := f.forge.CreateRepo(t, CreateRepoOpts{
		Name:          r.Name,
		Description:   name,
		DefaultBranch: TestDefaultBranch,
	})
	require.NoError(t, err)
	require.NotNil(t, repo)

	return r
}

func (f *Framework) CloneRepo(t *testing.T, r *Repository) *git.Repository {
	// TODO: Is this too low-level?
	t.Helper()

	repo, err := git.CloneRepo(t.Context(), f.logger, f.forge.CloneURL(t, r), TestDefaultBranch, f.forge.GitAuth(t))
	require.NoError(t, err)
	require.NotNil(t, repo)

	return repo
}

func (f *Framework) Run(t *testing.T, r *Repository, extraFiles []string) error {
	t.Helper()

	ctx := t.Context()

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs(append([]string{
		"run",
		fmt.Sprintf("--repo=%s", r.Name),
		fmt.Sprintf("--extra-files=%q", strings.Join(extraFiles, "\n")),
	}, f.forge.RunArguments()...))

	var stdout, stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.ExecuteContext(ctx)

	stdoutString := stdout.String()
	stderrString := stderr.String()

	if stdoutString != "" {
		t.Log("STDOUT:\n", stdoutString)
	}
	if stderrString != "" {
		t.Log("STDERR:\n", stderrString)
	}

	return err
}

func (f *Framework) HasReleasePR(t *testing.T, r *Repository, version string) *git.PullRequest {
	t.Helper()

	prs, err := f.forge.ListOpenPRs(t, r)
	require.NoError(t, err)

	expectedTitle := fmt.Sprintf("chore(main): release %s", version)
	index := slices.IndexFunc(prs, func(pr *git.PullRequest) bool {
		return pr.Title == expectedTitle
	})
	require.GreaterOrEqualf(t, index, 0, "release pull request for version %q does not exist", version)

	return prs[index]
}

func (f *Framework) MergeReleasePR(t *testing.T, r *Repository, pr *git.PullRequest) {
	t.Helper()

	err := f.forge.MergePR(t, r, pr)
	require.NoError(t, err)
}

func (f *Framework) HasTag(t *testing.T, r *Repository, version string) (tag *git.Tag) {
	t.Helper()

	tags, err := f.forge.ListTags(t, r)
	require.NoError(t, err)

	index := slices.IndexFunc(tags, func(tag *git.Tag) bool {
		return tag.Name == version
	})
	require.GreaterOrEqualf(t, index, 0, "tag %q does not exist", version)

	return tags[index]
}
