package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/xanzy/go-gitlab"

	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/pointer"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
)

const (
	PerPageMax        = 100
	PRStateOpen       = "opened"
	PRStateMerged     = "merged"
	PRStateEventClose = "close"
	EnvAPIToken       = "GITLAB_TOKEN" // nolint:gosec // Not actually a hardcoded credential
)

type GitLab struct {
	options *Options

	client *gitlab.Client
	log    *slog.Logger
}

func (g *GitLab) RepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%s", g.options.Repository)
}

func (g *GitLab) CloneURL() string {
	return fmt.Sprintf("https://gitlab.com/%s/%s.git", g.options.Path, g.options.Repo)
}

func (g *GitLab) ReleaseURL(version string) string {
	return fmt.Sprintf("https://gitlab.com/%s/%s/-/releases/%s", g.options.Path, g.options.Repo, version)
}

func (g *GitLab) PullRequestURL(id int) string {
	return fmt.Sprintf("https://gitlab.com/%s/%s/-/merge_requests/%d", g.options.Path, g.options.Repo, id)
}

func (g *GitLab) GitAuth() transport.AuthMethod {
	return &http.BasicAuth{
		// Username just needs to be any non-blank value
		Username: "api-token",
		Password: g.options.APIToken,
	}
}

func (g *GitLab) LatestTags(ctx context.Context) (git.Releases, error) {
	g.log.DebugContext(ctx, "listing all tags in gitlab repository")

	tags, err := all(func(listOptions gitlab.ListOptions) ([]*gitlab.Tag, *gitlab.Response, error) {
		return g.client.Tags.ListTags(g.options.ProjectID, &gitlab.ListTagsOptions{
			OrderBy:     pointer.Pointer("updated"),
			ListOptions: listOptions,
		}, gitlab.WithContext(ctx))
	})
	if err != nil {
		return git.Releases{}, err
	}

	var releases git.Releases
	for _, glTag := range tags {
		tag := &git.Tag{
			Hash: glTag.Commit.ID,
			Name: glTag.Name,
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

func (g *GitLab) CommitsSince(ctx context.Context, tag *git.Tag) ([]git.Commit, error) {
	var err error

	head := g.options.BaseBranch
	log := g.log.With("head", head)

	refName := ""
	if tag != nil {
		log = log.With("base", tag.Hash)
		refName = fmt.Sprintf("%s..%s", tag.Hash, head)
	} else {
		refName = head
	}
	log.Debug("listing commits", "ref.name", refName)

	gitLabCommits, err := all(func(listOptions gitlab.ListOptions) ([]*gitlab.Commit, *gitlab.Response, error) {
		return g.client.Commits.ListCommits(g.options.ProjectID, &gitlab.ListCommitsOptions{
			RefName:     &refName,
			ListOptions: listOptions,
		}, gitlab.WithContext(ctx))
	})
	if err != nil {
		return nil, err
	}

	var commits = make([]git.Commit, 0, len(gitLabCommits))
	for _, ghCommit := range gitLabCommits {
		commit := git.Commit{
			Hash:    ghCommit.ID,
			Message: ghCommit.Message,
		}
		commit.PullRequest, err = g.prForCommit(ctx, commit)
		if err != nil {
			return nil, fmt.Errorf("failed to check for commit pull request: %w", err)
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (g *GitLab) prForCommit(ctx context.Context, commit git.Commit) (*git.PullRequest, error) {
	// We naively look up the associated MR for each commit through the "List merge requests associated with a commit"
	// endpoint. This requires len(commits) requests.
	// Using the "List merge requests" endpoint might be faster, as it allows us to fetch 100 arbitrary MRs per request,
	// but worst case we need to look up all MRs made in the repository ever.

	log := g.log.With("commit.hash", commit.Hash)

	log.Debug("fetching pull requests associated with commit")
	associatedMRs, _, err := g.client.Commits.ListMergeRequestsByCommit(
		g.options.ProjectID, commit.Hash,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	var mergeRequest *gitlab.MergeRequest
	for _, mr := range associatedMRs {
		// We only look for the MR that has this commit set as the "merge commit" => The result of squashing this branch onto main
		if mr.MergeCommitSHA == commit.Hash {
			mergeRequest = mr
			break
		}
	}

	if mergeRequest == nil {
		return nil, nil
	}

	return gitlabMRToPullRequest(mergeRequest), nil
}

func (g *GitLab) EnsureLabelsExist(ctx context.Context, labels []releasepr.Label) error {
	g.log.Debug("fetching labels on repo")
	glLabels, err := all(func(listOptions gitlab.ListOptions) ([]*gitlab.Label, *gitlab.Response, error) {
		return g.client.Labels.ListLabels(g.options.ProjectID, &gitlab.ListLabelsOptions{
			ListOptions: listOptions,
		}, gitlab.WithContext(ctx))
	})
	if err != nil {
		return err
	}

	for _, label := range labels {
		if !slices.ContainsFunc(glLabels, func(glLabel *gitlab.Label) bool {
			return glLabel.Name == label.Name
		}) {
			g.log.Info("creating label in repository", "label.name", label)
			_, _, err := g.client.Labels.CreateLabel(g.options.ProjectID, &gitlab.CreateLabelOptions{
				Name:        pointer.Pointer(label.Name),
				Color:       pointer.Pointer("#" + label.Color),
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

func (g *GitLab) PullRequestForBranch(ctx context.Context, branch string) (*releasepr.ReleasePullRequest, error) {
	// There should only be a single open merge request from branch into g.options.BaseBranch at any given moment.
	// We can skip pagination and just return the first result.
	mrs, _, err := g.client.MergeRequests.ListProjectMergeRequests(g.options.ProjectID, &gitlab.ListProjectMergeRequestsOptions{
		State:        pointer.Pointer(PRStateOpen),
		SourceBranch: pointer.Pointer(branch),
		TargetBranch: pointer.Pointer(g.options.BaseBranch),
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: PerPageMax,
		},
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if len(mrs) >= 1 {
		return gitlabMRToReleasePullRequest(mrs[0]), nil
	}

	return nil, nil
}

func (g *GitLab) CreatePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	labels := make(gitlab.LabelOptions, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}

	glMR, _, err := g.client.MergeRequests.CreateMergeRequest(g.options.ProjectID, &gitlab.CreateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		SourceBranch: &pr.Head,
		TargetBranch: &g.options.BaseBranch,
		Labels:       &labels,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return err
	}

	pr.ID = glMR.IID

	return nil
}

func (g *GitLab) UpdatePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := g.client.MergeRequests.UpdateMergeRequest(g.options.ProjectID, pr.ID, &gitlab.UpdateMergeRequestOptions{
		Title:       &pr.Title,
		Description: &pr.Description,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return err
	}

	return nil
}

func (g *GitLab) SetPullRequestLabels(ctx context.Context, pr *releasepr.ReleasePullRequest, remove, add []releasepr.Label) error {
	removeLabels := make(gitlab.LabelOptions, 0, len(remove))
	for _, label := range remove {
		removeLabels = append(removeLabels, label.Name)
	}

	addLabels := make(gitlab.LabelOptions, 0, len(add))
	for _, label := range add {
		addLabels = append(addLabels, label.Name)
	}

	_, _, err := g.client.MergeRequests.UpdateMergeRequest(g.options.ProjectID, pr.ID, &gitlab.UpdateMergeRequestOptions{
		RemoveLabels: &removeLabels,
		AddLabels:    &addLabels,
	}, gitlab.WithContext(ctx))

	if err != nil {
		return err
	}

	return nil
}

func (g *GitLab) ClosePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := g.client.MergeRequests.UpdateMergeRequest(g.options.ProjectID, pr.ID, &gitlab.UpdateMergeRequestOptions{
		StateEvent: pointer.Pointer(PRStateEventClose),
	}, gitlab.WithContext(ctx))

	if err != nil {
		return err
	}

	return nil
}

func (g *GitLab) PendingReleases(ctx context.Context, pendingLabel releasepr.Label) ([]*releasepr.ReleasePullRequest, error) {
	glMRs, err := all(func(listOptions gitlab.ListOptions) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
		return g.client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
			State:        pointer.Pointer(PRStateMerged),
			Labels:       &gitlab.LabelOptions{pendingLabel.Name},
			TargetBranch: pointer.Pointer(g.options.BaseBranch),
			ListOptions:  listOptions,
		}, gitlab.WithContext(ctx))
	})
	if err != nil {
		return nil, err
	}

	prs := make([]*releasepr.ReleasePullRequest, 0, len(glMRs))

	for _, mr := range glMRs {
		prs = append(prs, gitlabMRToReleasePullRequest(mr))
	}

	return prs, nil
}

func (g *GitLab) CreateRelease(ctx context.Context, commit git.Commit, title, changelog string, _, _ bool) error {
	_, _, err := g.client.Releases.CreateRelease(g.options.ProjectID, &gitlab.CreateReleaseOptions{
		Name:        &title,
		TagName:     &title,
		Description: &changelog,
		Ref:         &commit.Hash,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func all[T any](f func(listOptions gitlab.ListOptions) ([]T, *gitlab.Response, error)) ([]T, error) {
	results := make([]T, 0)
	page := 1

	for {
		pageResults, resp, err := f(gitlab.ListOptions{Page: page, PerPage: PerPageMax})
		if err != nil {
			return nil, err
		}

		results = append(results, pageResults...)

		if page == resp.TotalPages || resp.TotalPages == 0 {
			return results, nil
		}
		page = resp.NextPage
	}
}

func gitlabMRToPullRequest(pr *gitlab.MergeRequest) *git.PullRequest {
	return &git.PullRequest{
		ID:          pr.IID,
		Title:       pr.Title,
		Description: pr.Description,
	}
}

func gitlabMRToReleasePullRequest(pr *gitlab.MergeRequest) *releasepr.ReleasePullRequest {
	labels := make([]releasepr.Label, 0, len(pr.Labels))
	for _, labelName := range pr.Labels {
		if i := slices.IndexFunc(releasepr.KnownLabels, func(label releasepr.Label) bool {
			return label.Name == labelName
		}); i >= 0 {
			labels = append(labels, releasepr.KnownLabels[i])
		}
	}

	var releaseCommit *git.Commit
	if pr.MergeCommitSHA != "" {
		releaseCommit = &git.Commit{Hash: pr.MergeCommitSHA}
	}

	return &releasepr.ReleasePullRequest{
		PullRequest: *gitlabMRToPullRequest(pr),
		Labels:      labels,

		Head:          pr.SHA,
		ReleaseCommit: releaseCommit,
	}
}

func (g *Options) autodiscover() {
	// Read settings from GitLab-CI env vars
	if apiToken := os.Getenv(EnvAPIToken); apiToken != "" {
		g.APIToken = apiToken
	}

	// TODO: Replace hardcode project-id with a better alternative
	g.ProjectID = 60698565
}

type Options struct {
	forge.Options

	Path      string
	Repo      string
	ProjectID int

	APIToken string
}

func New(log *slog.Logger, options *Options) (*GitLab, error) {
	log = log.With("forge", "gitlab")
	options.autodiscover()

	client, err := gitlab.NewClient(options.APIToken)
	if err != nil {
		return nil, err
	}

	gl := &GitLab{
		options: options,

		client: client,
		log:    log,
	}

	return gl, nil
}
