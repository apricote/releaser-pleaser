package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/cmd/rp/cmd"
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
	runID string
	forge TestForge
}

func NewFramework(ctx context.Context, forge TestForge) (*Framework, error) {
	f := &Framework{
		runID: randomString(),
		forge: forge,
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

	t.Log(stdoutString)
	t.Log(stderrString)

	return err
}
