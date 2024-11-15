package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v66/github"

	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/pointer"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
)

const (
	PerPageMax    = 100
	PRStateOpen   = "open"
	PRStateClosed = "closed"
	EnvAPIToken   = "GITHUB_TOKEN" // nolint:gosec // Not actually a hardcoded credential
	EnvUsername   = "GITHUB_USER"
	EnvRepository = "GITHUB_REPOSITORY"
)

var _ forge.Forge = &GitHub{}

type GitHub struct {
	options *Options

	client *github.Client
	log    *slog.Logger
}

func (g *GitHub) RepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", g.options.Owner, g.options.Repo)
}

func (g *GitHub) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s/%s.git", g.options.Owner, g.options.Repo)
}

func (g *GitHub) ReleaseURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", g.options.Owner, g.options.Repo, version)
}

func (g *GitHub) PullRequestURL(id int) string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", g.options.Owner, g.options.Repo, id)
}

func (g *GitHub) GitAuth() transport.AuthMethod {
	return &http.BasicAuth{
		Username: g.options.Username,
		Password: g.options.APIToken,
	}
}

func (g *GitHub) LatestTags(ctx context.Context) (git.Releases, error) {
	g.log.DebugContext(ctx, "listing all tags in github repository")

	tags, err := all(func(listOptions github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
		return g.client.Repositories.ListTags(
			ctx, g.options.Owner, g.options.Repo,
			&listOptions,
		)
	})
	if err != nil {
		return git.Releases{}, err
	}

	var releases git.Releases

	for _, ghTag := range tags {
		tag := &git.Tag{
			Hash: ghTag.GetCommit().GetSHA(),
			Name: ghTag.GetName(),
		}

		version, err := semver.Parse(strings.TrimPrefix(tag.Name, "v"))
		if err != nil {
			g.log.WarnContext(
				ctx, "unable to parse tag as semver, skipping",
				"tag.name", tag.Name,
				"tag.hash", tag.Hash,
				"error", err,
			)
			continue
		}

		if releases.Latest == nil {
			releases.Latest = tag
		}
		if len(version.Pre) == 0 {
			// Stable version tag
			// We return once we have found the latest stable tag, not needed to look at every single tag.
			releases.Stable = tag
			break
		}
	}

	return releases, nil
}

func (g *GitHub) CommitsSince(ctx context.Context, tag *git.Tag) ([]git.Commit, error) {
	var repositoryCommits []*github.RepositoryCommit
	var err error
	if tag != nil {
		repositoryCommits, err = g.commitsSinceTag(ctx, tag)
	} else {
		repositoryCommits, err = g.commitsSinceInit(ctx)
	}

	if err != nil {
		return nil, err
	}

	var commits = make([]git.Commit, 0, len(repositoryCommits))
	for _, ghCommit := range repositoryCommits {
		commit := git.Commit{
			Hash:    ghCommit.GetSHA(),
			Message: ghCommit.GetCommit().GetMessage(),
		}
		commit.PullRequest, err = g.prForCommit(ctx, commit)
		if err != nil {
			return nil, fmt.Errorf("failed to check for commit pull request: %w", err)
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (g *GitHub) commitsSinceTag(ctx context.Context, tag *git.Tag) ([]*github.RepositoryCommit, error) {
	head := g.options.BaseBranch
	log := g.log.With("base", tag.Hash, "head", head)
	log.Debug("comparing commits")

	repositoryCommits, err := all(
		func(listOptions github.ListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
			comparison, resp, err := g.client.Repositories.CompareCommits(
				ctx, g.options.Owner, g.options.Repo,
				tag.Hash, head, &listOptions)
			if err != nil {
				return nil, nil, err
			}

			return comparison.Commits, resp, err
		})
	if err != nil {
		return nil, err
	}

	return repositoryCommits, nil
}

func (g *GitHub) commitsSinceInit(ctx context.Context) ([]*github.RepositoryCommit, error) {
	head := g.options.BaseBranch
	log := g.log.With("head", head)
	log.Debug("listing all commits")

	repositoryCommits, err := all(
		func(listOptions github.ListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
			return g.client.Repositories.ListCommits(
				ctx, g.options.Owner, g.options.Repo,
				&github.CommitsListOptions{
					SHA:         head,
					ListOptions: listOptions,
				})
		})
	if err != nil {
		return nil, err
	}

	return repositoryCommits, nil
}

func (g *GitHub) prForCommit(ctx context.Context, commit git.Commit) (*git.PullRequest, error) {
	// We naively look up the associated PR for each commit through the "List pull requests associated with a commit"
	// endpoint. This requires len(commits) requests.
	// Using the "List pull requests" endpoint might be faster, as it allows us to fetch 100 arbitrary PRs per request,
	// but worst case we need to look up all PRs made in the repository ever.

	g.log.Debug("fetching pull requests associated with commit", "commit.hash", commit.Hash)

	associatedPRs, err := all(
		func(listOptions github.ListOptions) ([]*github.PullRequest, *github.Response, error) {
			return g.client.PullRequests.ListPullRequestsWithCommit(
				ctx, g.options.Owner, g.options.Repo,
				commit.Hash, &listOptions)
		})
	if err != nil {
		return nil, err
	}

	var pullRequest *github.PullRequest
	for _, pr := range associatedPRs {
		// We only look for the PR that has this commit set as the "merge commit" => The result of squashing this branch onto main
		if pr.GetMergeCommitSHA() == commit.Hash {
			pullRequest = pr
			break
		}
	}
	if pullRequest == nil {
		return nil, nil
	}

	return gitHubPRToPullRequest(pullRequest), nil
}

func (g *GitHub) EnsureLabelsExist(ctx context.Context, labels []releasepr.Label) error {
	g.log.Debug("fetching labels on repo")
	ghLabels, err := all(func(listOptions github.ListOptions) ([]*github.Label, *github.Response, error) {
		return g.client.Issues.ListLabels(
			ctx, g.options.Owner, g.options.Repo,
			&listOptions)
	})
	if err != nil {
		return err
	}

	for _, label := range labels {
		if !slices.ContainsFunc(ghLabels, func(ghLabel *github.Label) bool {
			return ghLabel.GetName() == label.Name
		}) {
			g.log.Info("creating label in repository", "label.name", label.Name)
			_, _, err = g.client.Issues.CreateLabel(
				ctx, g.options.Owner, g.options.Repo,
				&github.Label{
					Name:        pointer.Pointer(label.Name),
					Color:       pointer.Pointer(label.Color),
					Description: pointer.Pointer(label.Description),
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GitHub) PullRequestForBranch(ctx context.Context, branch string) (*releasepr.ReleasePullRequest, error) {

	prs, err := all(
		func(listOptions github.ListOptions) ([]*github.PullRequest, *github.Response, error) {
			return g.client.PullRequests.ListPullRequestsWithCommit(ctx, g.options.Owner, g.options.Repo, branch, &listOptions)
		})
	if err != nil {
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) {
			if ghErr.Message == fmt.Sprintf("No commit found for SHA: %s", branch) {
				return nil, nil
			}
		}
		return nil, err
	}

	for _, pr := range prs {
		if pr.GetBase().GetRef() == g.options.BaseBranch && pr.GetHead().GetRef() == branch && pr.GetState() == PRStateOpen {
			return gitHubPRToReleasePullRequest(pr), nil
		}
	}

	return nil, nil
}

func (g *GitHub) CreatePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	ghPR, _, err := g.client.PullRequests.Create(
		ctx, g.options.Owner, g.options.Repo,
		&github.NewPullRequest{
			Title: &pr.Title,
			Head:  &pr.Head,
			Base:  &g.options.BaseBranch,
			Body:  &pr.Description,
		},
	)
	if err != nil {
		return err
	}

	// TODO: String ID?
	pr.ID = ghPR.GetNumber()

	err = g.SetPullRequestLabels(ctx, pr, []releasepr.Label{}, pr.Labels)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) UpdatePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := g.client.PullRequests.Edit(
		ctx, g.options.Owner, g.options.Repo,
		pr.ID, &github.PullRequest{
			Title: &pr.Title,
			Body:  &pr.Description,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) SetPullRequestLabels(ctx context.Context, pr *releasepr.ReleasePullRequest, remove, add []releasepr.Label) error {
	for _, label := range remove {
		_, err := g.client.Issues.RemoveLabelForIssue(
			ctx, g.options.Owner, g.options.Repo,
			pr.ID, label.Name,
		)
		if err != nil {
			return err
		}
	}

	addString := make([]string, 0, len(add))
	for _, label := range add {
		addString = append(addString, label.Name)
	}

	_, _, err := g.client.Issues.AddLabelsToIssue(
		ctx, g.options.Owner, g.options.Repo,
		pr.ID, addString,
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) ClosePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := g.client.PullRequests.Edit(
		ctx, g.options.Owner, g.options.Repo,
		pr.ID, &github.PullRequest{
			State: pointer.Pointer(PRStateClosed),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) PendingReleases(ctx context.Context, pendingLabel releasepr.Label) ([]*releasepr.ReleasePullRequest, error) {
	ghPRs, err := all(func(listOptions github.ListOptions) ([]*github.PullRequest, *github.Response, error) {
		return g.client.PullRequests.List(
			ctx, g.options.Owner, g.options.Repo,
			&github.PullRequestListOptions{
				State:       PRStateClosed,
				Base:        g.options.BaseBranch,
				ListOptions: listOptions,
			})
	})
	if err != nil {
		return nil, err
	}

	prs := make([]*releasepr.ReleasePullRequest, 0, len(ghPRs))

	for _, pr := range ghPRs {
		pending := slices.ContainsFunc(pr.Labels, func(l *github.Label) bool {
			return l.GetName() == pendingLabel.Name
		})
		if !pending {
			continue
		}

		// pr.Merged is always nil :(
		if pr.MergedAt == nil {
			// Closed and not merged
			continue
		}

		prs = append(prs, gitHubPRToReleasePullRequest(pr))
	}

	return prs, nil
}

func (g *GitHub) CreateRelease(ctx context.Context, commit git.Commit, title, changelog string, preRelease, latest bool) error {
	makeLatest := ""
	if latest {
		makeLatest = "true"
	} else {
		makeLatest = "false"
	}
	_, _, err := g.client.Repositories.CreateRelease(
		ctx, g.options.Owner, g.options.Repo,
		&github.RepositoryRelease{
			TagName:         &title,
			TargetCommitish: &commit.Hash,
			Name:            &title,
			Body:            &changelog,
			Prerelease:      &preRelease,
			MakeLatest:      &makeLatest,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func all[T any](f func(listOptions github.ListOptions) ([]T, *github.Response, error)) ([]T, error) {
	results := make([]T, 0)
	page := 1

	for {
		pageResults, resp, err := f(github.ListOptions{Page: page, PerPage: PerPageMax})
		if err != nil {
			return nil, err
		}

		results = append(results, pageResults...)

		if page == resp.LastPage || resp.LastPage == 0 {
			return results, nil
		}
		page = resp.NextPage
	}
}

func gitHubPRToPullRequest(pr *github.PullRequest) *git.PullRequest {
	return &git.PullRequest{
		ID:          pr.GetNumber(),
		Title:       pr.GetTitle(),
		Description: pr.GetBody(),
	}
}

func gitHubPRToReleasePullRequest(pr *github.PullRequest) *releasepr.ReleasePullRequest {
	labels := make([]releasepr.Label, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labelName := label.GetName()
		if i := slices.IndexFunc(releasepr.KnownLabels, func(label releasepr.Label) bool {
			return label.Name == labelName
		}); i >= 0 {
			labels = append(labels, releasepr.KnownLabels[i])
		}
	}

	var releaseCommit *git.Commit
	if pr.MergeCommitSHA != nil {
		releaseCommit = &git.Commit{Hash: pr.GetMergeCommitSHA()}
	}

	return &releasepr.ReleasePullRequest{
		PullRequest: *gitHubPRToPullRequest(pr),
		Labels:      labels,

		Head:          pr.GetHead().GetRef(),
		ReleaseCommit: releaseCommit,
	}
}

func (g *Options) autodiscover() {
	if apiToken := os.Getenv(EnvAPIToken); apiToken != "" {
		g.APIToken = apiToken
	}
	// TODO: Check if there is a better solution for cloning/pushing locally
	if username := os.Getenv(EnvUsername); username != "" {
		g.Username = username
	}

	if envRepository := os.Getenv(EnvRepository); envRepository != "" {
		// GITHUB_REPOSITORY=apricote/releaser-pleaser
		parts := strings.Split(envRepository, "/")
		if len(parts) == 2 {
			g.Owner = parts[0]
			g.Repo = parts[1]
			g.Repository = envRepository
		}
	}
}

type Options struct {
	forge.Options

	Owner string
	Repo  string

	APIToken string
	Username string
}

func New(log *slog.Logger, options *Options) *GitHub {
	options.autodiscover()

	client := github.NewClient(nil)
	if options.APIToken != "" {
		client = client.WithAuthToken(options.APIToken)
	}

	gh := &GitHub{
		options: options,

		client: client,
		log:    log.With("forge", "github"),
	}

	return gh
}
