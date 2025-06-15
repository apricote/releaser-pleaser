//go:build e2e_forgejo

package forgejo

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"testing"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/apricote/releaser-pleaser/cmd/rp/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestAPIURL = "http://localhost:3000"

	TestUserNameTemplate  = "rp-%s"
	TestUserPassword      = "releaser-pleaser"
	TestUserEmailTemplate = "releaser-pleaser-%s@example.com"
	TestTokenName         = "rp"
	TestTokenScopes       = "write:user,write:issue,write:repository"

	TestDefaultBranch = "main"
)

var (
	TestToken    string
	TestUserName string
	TestClient   *forgejo.Client
)

func randomSuffix() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func setupTestUser(ctx context.Context, suffix string) string {
	TestUserName = fmt.Sprintf(TestUserNameTemplate, suffix)

	if output, err := exec.CommandContext(ctx,
		"docker", "compose", "exec", "--user=1000", "forgejo",
		"forgejo", "admin", "user", "create",
		"--username", TestUserName,
		"--password", TestUserPassword,
		"--email", fmt.Sprintf(TestUserEmailTemplate, suffix),
		"--must-change-password=false",
	).CombinedOutput(); err != nil {
		slog.ErrorContext(ctx, "failed to create forgejo user", "err", err, "output", output)
		panic(err)
	}

	token, err := exec.CommandContext(ctx,
		"docker", "compose", "exec", "--user=1000", "forgejo",
		"forgejo", "admin", "user", "generate-access-token",
		"--username", TestUserName,
		"--token-name", TestTokenName,
		"--scopes", TestTokenScopes,
		"--raw",
	).Output()
	if err != nil {
		slog.ErrorContext(ctx, "failed to create forgejo token", "err", err)
		panic(err)
	}

	return strings.TrimSpace(string(token))
}

func setupTestClient(ctx context.Context, token string) *forgejo.Client {
	client, err := forgejo.NewClient(TestAPIURL,
		forgejo.SetToken(token),
		forgejo.SetUserAgent("releaser-pleaser-e2e-tests"),
		forgejo.SetContext(ctx),
		// forgejo.SetDebugMode(),
	)
	if err != nil {
		panic(err)
	}

	return client
}

type Repository struct {
	Name string
}

func NewRepository(t *testing.T, name string) *Repository {
	t.Helper()

	r := &Repository{
		Name: fmt.Sprintf("%s-%s", name, randomSuffix())}

	repo, _, err := TestClient.CreateRepo(forgejo.CreateRepoOption{
		Name:          r.Name,
		Description:   name,
		DefaultBranch: TestDefaultBranch,
	})
	require.NoError(t, err)
	require.NotNil(t, repo)

	return r
}

func Run(t *testing.T, r *Repository, extraFiles []string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	t.Helper()

	ctx := t.Context()

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"run", "--forge=forgejo",
		fmt.Sprintf("--branch=%s", TestDefaultBranch),
		fmt.Sprintf("--owner=%s", TestUserName),
		fmt.Sprintf("--repo=%s", r.Name),
		fmt.Sprintf("--extra-files=%q", strings.Join(extraFiles, "\n")),
		fmt.Sprintf("--api-url=%s", TestAPIURL),
		fmt.Sprintf("--api-token=%s", TestToken),
		fmt.Sprintf("--username=%s", TestUserName),
	})

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err = rootCmd.ExecuteContext(ctx)

	return stdout, stderr, err
}

func MustRun(t *testing.T, r *Repository, extraFiles []string) {
	t.Helper()

	stdout, stderr, err := Run(t, r, extraFiles)
	if !assert.NoError(t, err) {
		t.Log(stdout)
		t.Log(stderr)
	}
}
