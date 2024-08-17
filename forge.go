package rp

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
	"github.com/google/go-github/v63/github"
)

const (
	GitHubPerPageMax    = 100
	GitHubPRStateOpen   = "open"
	GitHubPRStateClosed = "closed"
	GitHubEnvAPIToken   = "GITHUB_TOKEN"
	GitHubEnvUsername   = "GITHUB_USER"
	GitHubEnvRepository = "GITHUB_REPOSITORY"
	GitHubLabelColor    = "dedede"
)

type Changeset struct {
	URL              string
	Identifier       string
	ChangelogEntries []AnalyzedCommit
}

type Forge interface {
	RepoURL() string
	CloneURL() string
	ReleaseURL(version string) string

	GitAuth() transport.AuthMethod

	// LatestTags returns the last stable tag created on the main branch. If there is a more recent pre-release tag,
	// that is also returned. If no tag is found, it returns nil.
	LatestTags(context.Context) (Releases, error)

	// CommitsSince returns all commits to main branch after the Tag. The tag can be `nil`, in which case this
	// function should return all commits.
	CommitsSince(context.Context, *Tag) ([]Commit, error)

	// Changesets looks up the Pull/Merge Requests for each commit, returning its parsed metadata.
	Changesets(context.Context, []Commit) ([]Changeset, error)

	// EnsureLabelsExist verifies that all desired labels are available on the repository. If labels are missing, they
	// are created them.
	EnsureLabelsExist(context.Context, []Label) error

	// PullRequestForBranch returns the open pull request between the branch and ForgeOptions.BaseBranch. If no open PR
	// exists, it returns nil.
	PullRequestForBranch(context.Context, string) (*ReleasePullRequest, error)

	// CreatePullRequest opens a new pull/merge request for the ReleasePullRequest.
	CreatePullRequest(context.Context, *ReleasePullRequest) error

	// UpdatePullRequest updates the pull/merge request identified through the ID of
	// the ReleasePullRequest to the current description and title.
	UpdatePullRequest(context.Context, *ReleasePullRequest) error

	// SetPullRequestLabels updates the pull/merge request identified through the ID of
	// the ReleasePullRequest to the current labels.
	SetPullRequestLabels(ctx context.Context, pr *ReleasePullRequest, remove, add []Label) error

	// ClosePullRequest closes the pull/merge request identified through the ID of
	// the ReleasePullRequest, as it is no longer required.
	ClosePullRequest(context.Context, *ReleasePullRequest) error

	// PendingReleases returns a list of ReleasePullRequest. The list should contain all pull/merge requests that are
	// merged and have the matching label.
	PendingReleases(context.Context, Label) ([]*ReleasePullRequest, error)

	// CreateRelease creates a release on the Forge, pointing at the commit with the passed in details.
	CreateRelease(ctx context.Context, commit Commit, title, changelog string, prerelease, latest bool) error
}

type ForgeOptions struct {
	Repository string
	BaseBranch string
}

var _ Forge = &GitHub{}

// var _ Forge = &GitLab{}

type GitHub struct {
	options *GitHubOptions

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

func (g *GitHub) GitAuth() transport.AuthMethod {
	return &http.BasicAuth{
		Username: g.options.Username,
		Password: g.options.APIToken,
	}
}

func (g *GitHub) LatestTags(ctx context.Context) (Releases, error) {
	g.log.DebugContext(ctx, "listing all tags in github repository")

	page := 1

	var releases Releases

	for {
		tags, resp, err := g.client.Repositories.ListTags(
			ctx, g.options.Owner, g.options.Repo,
			&github.ListOptions{Page: page, PerPage: GitHubPerPageMax},
		)
		if err != nil {
			return Releases{}, err
		}

		for _, ghTag := range tags {
			tag := &Tag{
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

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return releases, nil
}

func (g *GitHub) CommitsSince(ctx context.Context, tag *Tag) ([]Commit, error) {
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

	var commits = make([]Commit, 0, len(repositoryCommits))
	for _, ghCommit := range repositoryCommits {
		commits = append(commits, Commit{
			Hash:    ghCommit.GetSHA(),
			Message: ghCommit.GetCommit().GetMessage(),
		})
	}

	return commits, nil
}

func (g *GitHub) commitsSinceTag(ctx context.Context, tag *Tag) ([]*github.RepositoryCommit, error) {
	head := g.options.BaseBranch
	log := g.log.With("base", tag.Hash, "head", head)
	log.Debug("comparing commits", "base", tag.Hash, "head", head)

	page := 1

	var repositoryCommits []*github.RepositoryCommit
	for {
		log.Debug("fetching page", "page", page)
		comparison, resp, err := g.client.Repositories.CompareCommits(
			ctx, g.options.Owner, g.options.Repo,
			tag.Hash, head, &github.ListOptions{
				Page:    page,
				PerPage: GitHubPerPageMax,
			})
		if err != nil {
			return nil, err
		}

		if repositoryCommits == nil {
			// Pre-initialize slice on first request
			log.Debug("found commits", "length", comparison.GetTotalCommits())
			repositoryCommits = make([]*github.RepositoryCommit, 0, comparison.GetTotalCommits())
		}

		repositoryCommits = append(repositoryCommits, comparison.Commits...)

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return repositoryCommits, nil
}

func (g *GitHub) commitsSinceInit(ctx context.Context) ([]*github.RepositoryCommit, error) {
	head := g.options.BaseBranch
	log := g.log.With("head", head)
	log.Debug("listing all commits")

	page := 1

	var repositoryCommits []*github.RepositoryCommit
	for {
		log.Debug("fetching page", "page", page)
		commits, resp, err := g.client.Repositories.ListCommits(
			ctx, g.options.Owner, g.options.Repo,
			&github.CommitsListOptions{
				SHA: head,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: GitHubPerPageMax,
				},
			})
		if err != nil {
			return nil, err
		}

		if repositoryCommits == nil && resp.LastPage > 0 {
			// Pre-initialize slice on first request
			log.Debug("found commits", "pages", resp.LastPage)
			repositoryCommits = make([]*github.RepositoryCommit, 0, resp.LastPage*GitHubPerPageMax)
		}

		repositoryCommits = append(repositoryCommits, commits...)

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return repositoryCommits, nil
}

func (g *GitHub) Changesets(ctx context.Context, commits []Commit) ([]Changeset, error) {
	// We naively look up the associated PR for each commit through the "List pull requests associated with a commit"
	// endpoint. This requires len(commits) requests.
	// Using the "List pull requests" endpoint might be faster, as it allows us to fetch 100 arbitrary PRs per request,
	// but worst case we need to look up all PRs made in the repository ever.

	changesets := make([]Changeset, 0, len(commits))

	for _, commit := range commits {
		log := g.log.With("commit.hash", commit.Hash)
		page := 1
		var associatedPRs []*github.PullRequest

		for {
			log.Debug("fetching pull requests associated with commit", "page", page)
			prs, resp, err := g.client.PullRequests.ListPullRequestsWithCommit(
				ctx, g.options.Owner, g.options.Repo,
				commit.Hash, &github.ListOptions{
					Page:    page,
					PerPage: GitHubPerPageMax,
				})
			if err != nil {
				return nil, err
			}

			associatedPRs = append(associatedPRs, prs...)

			if page == resp.LastPage || resp.LastPage == 0 {
				break
			}
			page = resp.NextPage
		}

		var pullrequest *github.PullRequest
		for _, pr := range associatedPRs {
			// We only look for the PR that has this commit set as the "merge commit" => The result of squashing this branch onto main
			if pr.GetMergeCommitSHA() == commit.Hash {
				pullrequest = pr
				break
			}
		}
		if pullrequest == nil {
			log.Warn("did not find associated pull request, not considering it for changesets")
			// No pull request was found for this commit, nothing to do here
			// TODO: We could also return the minimal changeset for this commit, so at least it would come up in the changelog.
			continue
		}

		log = log.With("pullrequest.id", pullrequest.GetID())

		// TODO: Parse PR description for overrides
		changelogEntries, err := NewConventionalCommitsParser().AnalyzeCommits([]Commit{commit})
		if err != nil {
			log.Warn("unable to parse changelog entries", "error", err)
			continue
		}

		if len(changelogEntries) > 0 {
			changesets = append(changesets, Changeset{
				URL:              pullrequest.GetHTMLURL(),
				Identifier:       fmt.Sprintf("#%d", pullrequest.GetNumber()),
				ChangelogEntries: changelogEntries,
			})
		}
	}

	return changesets, nil
}

func (g *GitHub) EnsureLabelsExist(ctx context.Context, labels []Label) error {
	existingLabels := make([]string, 0, len(labels))

	page := 1

	for {
		g.log.Debug("fetching labels on repo", "page", page)
		ghLabels, resp, err := g.client.Issues.ListLabels(
			ctx, g.options.Owner, g.options.Repo,
			&github.ListOptions{
				Page:    page,
				PerPage: GitHubPerPageMax,
			})
		if err != nil {
			return err
		}

		for _, label := range ghLabels {
			existingLabels = append(existingLabels, label.GetName())
		}

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}
		page = resp.NextPage
	}

	for _, label := range labels {
		if !slices.Contains(existingLabels, string(label)) {
			g.log.Info("creating label in repository", "label.name", label)
			_, _, err := g.client.Issues.CreateLabel(
				ctx, g.options.Owner, g.options.Repo,
				&github.Label{
					Name:  Pointer(string(label)),
					Color: Pointer(GitHubLabelColor),
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GitHub) PullRequestForBranch(ctx context.Context, branch string) (*ReleasePullRequest, error) {
	page := 1

	for {
		prs, resp, err := g.client.PullRequests.ListPullRequestsWithCommit(ctx, g.options.Owner, g.options.Repo, branch, &github.ListOptions{
			Page:    page,
			PerPage: GitHubPerPageMax,
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
			if pr.GetBase().GetRef() == g.options.BaseBranch && pr.GetHead().GetRef() == branch && pr.GetState() == GitHubPRStateOpen {
				return gitHubPRToReleasePullRequest(pr), nil
			}
		}

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return nil, nil
}

func (g *GitHub) CreatePullRequest(ctx context.Context, pr *ReleasePullRequest) error {
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

	err = g.SetPullRequestLabels(ctx, pr, []Label{}, pr.Labels)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) UpdatePullRequest(ctx context.Context, pr *ReleasePullRequest) error {
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

func (g *GitHub) SetPullRequestLabels(ctx context.Context, pr *ReleasePullRequest, remove, add []Label) error {
	for _, label := range remove {
		_, err := g.client.Issues.RemoveLabelForIssue(
			ctx, g.options.Owner, g.options.Repo,
			pr.ID, string(label),
		)
		if err != nil {
			return err
		}
	}

	addString := make([]string, 0, len(add))
	for _, label := range add {
		addString = append(addString, string(label))
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

func (g *GitHub) ClosePullRequest(ctx context.Context, pr *ReleasePullRequest) error {
	_, _, err := g.client.PullRequests.Edit(
		ctx, g.options.Owner, g.options.Repo,
		pr.ID, &github.PullRequest{
			State: Pointer(GitHubPRStateClosed),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) PendingReleases(ctx context.Context, pendingLabel Label) ([]*ReleasePullRequest, error) {
	page := 1

	var prs []*ReleasePullRequest

	for {
		ghPRs, resp, err := g.client.PullRequests.List(
			ctx, g.options.Owner, g.options.Repo,
			&github.PullRequestListOptions{
				State: GitHubPRStateClosed,
				Base:  g.options.BaseBranch,
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: GitHubPerPageMax,
				},
			})
		if err != nil {
			return nil, err
		}

		if prs == nil && resp.LastPage > 0 {
			// Pre-initialize slice on first request
			g.log.Debug("found pending releases", "pages", resp.LastPage)
			prs = make([]*ReleasePullRequest, 0, (resp.LastPage-1)*GitHubPerPageMax)
		}

		for _, pr := range ghPRs {
			pending := slices.ContainsFunc(pr.Labels, func(l *github.Label) bool {
				return l.GetName() == string(pendingLabel)
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

		if page == resp.LastPage || resp.LastPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return prs, nil
}

func (g *GitHub) CreateRelease(ctx context.Context, commit Commit, title, changelog string, preRelease, latest bool) error {
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

func gitHubPRToReleasePullRequest(pr *github.PullRequest) *ReleasePullRequest {
	labels := make([]Label, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labelName := Label(label.GetName())
		if slices.Contains(KnownLabels, Label(label.GetName())) {
			labels = append(labels, labelName)
		}
	}

	var releaseCommit *Commit
	if pr.MergeCommitSHA != nil {
		releaseCommit = &Commit{Hash: pr.GetMergeCommitSHA()}
	}

	return &ReleasePullRequest{
		ID:          pr.GetNumber(),
		Title:       pr.GetTitle(),
		Description: pr.GetBody(),
		Labels:      labels,

		Head:          pr.GetHead().GetRef(),
		ReleaseCommit: releaseCommit,
	}
}

func (g *GitHubOptions) autodiscover() {
	if apiToken := os.Getenv(GitHubEnvAPIToken); apiToken != "" {
		g.APIToken = apiToken
	}
	// TODO: Check if there is a better solution for cloning/pushing locally
	if username := os.Getenv(GitHubEnvUsername); username != "" {
		g.Username = username
	}

	if envRepository := os.Getenv(GitHubEnvRepository); envRepository != "" {
		// GITHUB_REPOSITORY=apricote/releaser-pleaser
		parts := strings.Split(envRepository, "/")
		if len(parts) == 2 {
			g.Owner = parts[0]
			g.Repo = parts[1]
			g.Repository = envRepository
		}
	}
}

type GitHubOptions struct {
	ForgeOptions

	Owner string
	Repo  string

	APIToken string
	Username string
}

func NewGitHub(log *slog.Logger, options *GitHubOptions) *GitHub {
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

type GitLab struct {
	options ForgeOptions
}

func (g *GitLab) autodiscover() {
	// Read settings from GitLab-CI env vars
}

func NewGitLab(options ForgeOptions) *GitLab {
	gl := &GitLab{
		options: options,
	}

	gl.autodiscover()

	return gl
}

func (g *GitLab) RepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%s", g.options.Repository)
}

func Pointer[T any](value T) *T {
	return &value
}
