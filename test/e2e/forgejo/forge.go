package forgejo

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"testing"
	"time"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
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

func (f *TestForge) CloneURL(t *testing.T, repo *e2e.Repository) string {
	t.Helper()

	return fmt.Sprintf("%s/%s/%s.git", TestAPIURL, f.username, repo.Name)
}

func (f *TestForge) GitAuth(t *testing.T) transport.AuthMethod {
	t.Helper()

	return &http.BasicAuth{
		Username: f.username,
		Password: f.token,
	}
}

func (f *TestForge) ListOpenPRs(t *testing.T, repo *e2e.Repository) ([]*git.PullRequest, error) {
	t.Helper()

	fPRs, _, err := f.client.ListRepoPullRequests(f.username, repo.Name, forgejo.ListPullRequestsOptions{
		State: forgejo.StateOpen,
	})
	if err != nil {
		return nil, err
	}

	prs := make([]*git.PullRequest, 0, len(fPRs))
	for _, pr := range fPRs {
		prs = append(prs, forgejoPRToPullRequest(pr))
	}

	return prs, nil
}

func (f *TestForge) MergePR(t *testing.T, repo *e2e.Repository, pr *git.PullRequest) error {
	t.Helper()

	// Wait for the PR to become mergable
	retries := 10
	sleep := 1 * time.Second

	var fPR *forgejo.PullRequest
	var err error
	for range retries {
		fPR, _, err = f.client.GetPullRequest(f.username, repo.Name, pr.ID)
		if err != nil {
			t.Logf("sleeping, error while checking pr mergeable status: %v", err)
			time.Sleep(sleep)
			continue
		}

		if !fPR.Mergeable {
			t.Log("sleeping, pr not marked as mergeable yet")
			time.Sleep(sleep)
			continue
		}

		break
	}

	if !fPR.Mergeable {
		return fmt.Errorf("pull request not marked as mergable by forgejo after retries")
	}

	ok, resp, err := f.client.MergePullRequest(f.username, repo.Name, pr.ID, forgejo.MergePullRequestOption{Style: forgejo.MergeStyleSquash})
	if err != nil {
		return err
	}
	if !ok {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("merging pull request #%d failed: %v", pr.ID, string(respBody))
	}

	return nil
}

func (f *TestForge) ListTags(t *testing.T, repo *e2e.Repository) ([]*git.Tag, error) {
	t.Helper()

	fTags, _, err := f.client.ListRepoTags(f.username, repo.Name, forgejo.ListRepoTagsOptions{})
	require.NoError(t, err)

	tags := make([]*git.Tag, 0, len(fTags))
	for _, tag := range fTags {
		tags = append(tags, forgejoTagToTag(tag))
	}

	return tags, nil
}

func (f *TestForge) RunArguments() []string {
	return []string{"--forge=forgejo",
		fmt.Sprintf("--owner=%s", f.username),
		fmt.Sprintf("--api-url=%s", TestAPIURL),
		fmt.Sprintf("--api-token=%s", f.token),
		fmt.Sprintf("--username=%s", f.username),
	}
}

func forgejoPRToPullRequest(pr *forgejo.PullRequest) *git.PullRequest {
	return &git.PullRequest{
		ID:          pr.Index,
		Title:       pr.Title,
		Description: pr.Body,
	}
}

func forgejoTagToTag(tag *forgejo.Tag) *git.Tag {
	return &git.Tag{
		Hash: tag.Commit.SHA,
		Name: tag.Name,
	}
}
