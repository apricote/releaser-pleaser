package rp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v63/github"
)

type Changeset struct {
	URL              string
	Identifier       string
	ChangelogEntries []AnalyzedCommit
}

type Forge interface {
	RepoURL() string

	// LatestTag returns the last tag created on the main branch. If no tag is found, it returns nil.
	LatestTag(context.Context) (*Tag, error)

	// CommitsSince returns all commits to main branch after the Tag. The tag can be `nil`, in which case this
	// function should return all commits.
	CommitsSince(context.Context, *Tag) ([]Commit, error)

	// Changesets looks up the Pull/Merge Requests for each commit, returning its parsed metadata.
	Changesets(context.Context, []Commit) ([]Changeset, error)
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
	return fmt.Sprintf("https://github.com/%s", g.options.Repository)
}

func (g *GitHub) LatestTag(ctx context.Context) (*Tag, error) {
	g.log.Debug("listing all tags in github repository")
	// We only get the first page because the latest tag is returned as the first item
	tags, _, err := g.client.Repositories.ListTags(ctx, g.options.Owner, g.options.Repo, nil)
	if err != nil {
		return nil, err
	}

	if len(tags) > 0 {
		// TODO: Is tags sorted?
		tag := tags[0]
		return &Tag{
			Hash: tag.GetCommit().GetSHA(),
			Name: tag.GetName(),
		}, nil
	}

	return nil, nil
}

func (g *GitHub) CommitsSince(ctx context.Context, tag *Tag) ([]Commit, error) {
	var repositoryCommits []*github.RepositoryCommit
	var err error
	if tag != nil {
		repositoryCommits, err = g.commitsSinceTag(ctx, tag)
	} else {
		return nil, fmt.Errorf("not implemented")
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
				PerPage: 100,
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
					PerPage: 100,
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
		changelogEntries, _, err := AnalyzeCommits([]Commit{commit})
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

func (g *GitHubOptions) autodiscover() {
	// TODO: Read settings from GitHub Actions env vars
}

type GitHubOptions struct {
	ForgeOptions

	Owner string
	Repo  string

	APIToken string
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
