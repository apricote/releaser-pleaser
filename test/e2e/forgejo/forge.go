package forgejo

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"testing"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/apricote/releaser-pleaser/test/e2e"
)

const (
	TestAPIURL = "http://localhost:3000"

	TestUserNameTemplate  = "rp-%s"
	TestUserPassword      = "releaser-pleaser"
	TestUserEmailTemplate = "releaser-pleaser-%s@example.com"
	TestTokenName         = "rp"
	TestTokenScopes       = "write:user,write:issue,write:repository"
)

type TestForge struct {
	username string
	token    string
	client   *forgejo.Client
}

func (f *TestForge) Init(ctx context.Context, runID string) error {
	if err := f.initUser(ctx, runID); err != nil {
		return err
	}
	if err := f.initClient(ctx); err != nil {
		return err
	}

	return nil
}

func (f *TestForge) initUser(ctx context.Context, runID string) error {
	f.username = fmt.Sprintf(TestUserNameTemplate, runID)

	//gosec:disable G204
	if output, err := exec.CommandContext(ctx,
		"docker", "compose", "exec", "--user=1000", "forgejo",
		"forgejo", "admin", "user", "create",
		"--username", f.username,
		"--password", TestUserPassword,
		"--email", fmt.Sprintf(TestUserEmailTemplate, runID),
		"--must-change-password=false",
	).CombinedOutput(); err != nil {
		slog.Debug("create forgejo user output", "output", output)
		return fmt.Errorf("failed to create forgejo user: %w", err)
	}

	//gosec:disable G204
	token, err := exec.CommandContext(ctx,
		"docker", "compose", "exec", "--user=1000", "forgejo",
		"forgejo", "admin", "user", "generate-access-token",
		"--username", f.username,
		"--token-name", TestTokenName,
		"--scopes", TestTokenScopes,
		"--raw",
	).Output()
	if err != nil {
		return fmt.Errorf("failed to create forgejo token: %w", err)
	}

	f.token = strings.TrimSpace(string(token))

	return nil
}

func (f *TestForge) initClient(ctx context.Context) (err error) {
	f.client, err = forgejo.NewClient(TestAPIURL,
		forgejo.SetToken(f.token),
		forgejo.SetUserAgent("releaser-pleaser-e2e-tests"),
		forgejo.SetContext(ctx),
		// forgejo.SetDebugMode(),
	)
	return err
}

func (f *TestForge) CreateRepo(t *testing.T, opts e2e.CreateRepoOpts) (*e2e.Repository, error) {
	t.Helper()

	repo, _, err := f.client.CreateRepo(forgejo.CreateRepoOption{
		Name:          opts.Name,
		Description:   opts.Description,
		DefaultBranch: opts.DefaultBranch,
		Readme:        "Default",
		AutoInit:      true,
	})
	if err != nil {
		return nil, err
	}

	return &e2e.Repository{
		Name: repo.Name,
	}, nil
}

func (f *TestForge) RunArguments() []string {
	return []string{"--forge=forgejo",
		fmt.Sprintf("--owner=%s", f.username),
		fmt.Sprintf("--api-url=%s", TestAPIURL),
		fmt.Sprintf("--api-token=%s", f.token),
		fmt.Sprintf("--username=%s", f.username),
	}
}
