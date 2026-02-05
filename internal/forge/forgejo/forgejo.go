package forgejo

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/pointer"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
)

const ()

var _ forge.Forge = &Forgejo{}

type Forgejo struct {
	options *Options

	client *forgejo.Client
	log    *slog.Logger
}

func (f *Forgejo) RepoURL() string {
	return fmt.Sprintf("%s/%s/%s", f.options.APIURL, f.options.Owner, f.options.Repo)
}

func (f *Forgejo) CloneURL() string {
	return fmt.Sprintf("%s.git", f.RepoURL())
}

func (f *Forgejo) ReleaseURL(version string) string {
	return fmt.Sprintf("%s/releases/tag/%s", f.RepoURL(), version)
}

func (f *Forgejo) PullRequestURL(id int) string {
	return fmt.Sprintf("%s/pulls/%d", f.RepoURL(), id)
}

func (f *Forgejo) GitAuth() transport.AuthMethod {
	return &http.BasicAuth{
		Username: f.options.Username,
		Password: f.options.APIToken,
	}
}

func (f *Forgejo) CommitAuthor(ctx context.Context) (git.Author, error) {
	f.log.DebugContext(ctx, "getting commit author from current token user")

	user, _, err := f.client.GetMyUserInfo()
	if err != nil {
		return git.Author{}, err
	}

	// TODO: Same for other forges?
	name := user.FullName
	if name == "" {
		name = user.UserName
	}

	return git.Author{
		Name:  name,
		Email: user.Email,
	}, nil
}

func (f *Forgejo) LatestTags(ctx context.Context) (git.Releases, error) {
	f.log.DebugContext(ctx, "listing all tags in forgejo repository")

	tags, err := all(func(listOptions forgejo.ListOptions) ([]*forgejo.Tag, *forgejo.Response, error) {
		return f.client.ListRepoTags(f.options.Owner, f.options.Repo,
			forgejo.ListRepoTagsOptions{ListOptions: listOptions},
		)
	})
	if err != nil {
		return git.Releases{}, err
	}

	var releases git.Releases

	for _, fTag := range tags {
		tag := &git.Tag{
			Hash: fTag.Commit.SHA,
			Name: fTag.Name,
		}

		version, err := semver.Parse(strings.TrimPrefix(tag.Name, "v"))
		if err != nil {
			f.log.WarnContext(
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

func (f *Forgejo) CommitsSince(ctx context.Context, tag *git.Tag) ([]git.Commit, error) {
	var repositoryCommits []*forgejo.Commit
	var err error
	if tag != nil {
		repositoryCommits, err = f.commitsSinceTag(ctx, tag)
	} else {
		repositoryCommits, err = f.commitsSinceInit(ctx)
	}

	if err != nil {
		return nil, err
	}

	var commits = make([]git.Commit, 0, len(repositoryCommits))
	for _, fCommit := range repositoryCommits {
		commit := git.Commit{
			Hash:    fCommit.SHA,
			Message: fCommit.RepoCommit.Message,
		}
		commit.PullRequest, err = f.prForCommit(ctx, commit)
		if err != nil {
			return nil, fmt.Errorf("failed to check for commit pull request: %w", err)
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (f *Forgejo) commitsSinceTag(_ context.Context, tag *git.Tag) ([]*forgejo.Commit, error) {
	head := f.options.BaseBranch
	log := f.log.With("base", tag.Hash, "head", head)
	log.Debug("comparing commits")

	compare, _, err := f.client.CompareCommits(
		f.options.Owner, f.options.Repo,
		tag.Hash, head)
	if err != nil {
		return nil, err
	}

	return compare.Commits, nil
}

func (f *Forgejo) commitsSinceInit(_ context.Context) ([]*forgejo.Commit, error) {
	head := f.options.BaseBranch
	log := f.log.With("head", head)
	log.Debug("listing all commits")

	repositoryCommits, err := all(
		func(listOptions forgejo.ListOptions) ([]*forgejo.Commit, *forgejo.Response, error) {
			return f.client.ListRepoCommits(
				f.options.Owner, f.options.Repo,
				forgejo.ListCommitOptions{
					ListOptions: listOptions,
					SHA:         f.options.BaseBranch,
				})
		})
	if err != nil {
		return nil, err
	}

	return repositoryCommits, nil
}

func (f *Forgejo) prForCommit(_ context.Context, commit git.Commit) (*git.PullRequest, error) {
	// We naively look up the associated PR for each commit through the "List pull requests associated with a commit"
	// endpoint. This requires len(commits) requests.
	// Using the "List pull requests" endpoint might be faster, as it allows us to fetch 100 arbitrary PRs per request,
	// but worst case we need to look up all PRs made in the repository ever.

	f.log.Debug("fetching pull requests associated with commit", "commit.hash", commit.Hash)

	pullRequest, _, err := f.client.GetCommitPullRequest(
		f.options.Owner, f.options.Repo,
		commit.Hash,
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "pull request does not exist") {
			return nil, nil
		}

		return nil, err
	}

	return forgejoPRToPullRequest(pullRequest), nil
}

func (f *Forgejo) EnsureLabelsExist(_ context.Context, labels []releasepr.Label) error {
	f.log.Debug("fetching labels on repo")
	fLabels, err := all(func(listOptions forgejo.ListOptions) ([]*forgejo.Label, *forgejo.Response, error) {
		return f.client.ListRepoLabels(
			f.options.Owner, f.options.Repo,
			forgejo.ListLabelsOptions{ListOptions: listOptions})
	})
	if err != nil {
		return err
	}

	for _, label := range labels {
		if !slices.ContainsFunc(fLabels, func(fLabel *forgejo.Label) bool {
			return fLabel.Name == label.Name
		}) {
			f.log.Info("creating label in repository", "label.name", label.Name)
			_, _, err = f.client.CreateLabel(
				f.options.Owner, f.options.Repo,
				forgejo.CreateLabelOption{
					Name:        label.Name,
					Color:       label.Color,
					Description: label.Description,
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *Forgejo) PullRequestForBranch(_ context.Context, branch string) (*releasepr.ReleasePullRequest, error) {
	prs, err := all(
		func(listOptions forgejo.ListOptions) ([]*forgejo.PullRequest, *forgejo.Response, error) {
			return f.client.ListRepoPullRequests(
				f.options.Owner, f.options.Repo,
				forgejo.ListPullRequestsOptions{
					ListOptions: listOptions,
					State:       forgejo.StateOpen,
				},
			)
		},
	)
	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		if pr.Base.Ref == f.options.BaseBranch && pr.Head.Ref == branch {
			return forgejoPRToReleasePullRequest(pr), nil
		}
	}

	return nil, nil
}

func (f *Forgejo) CreatePullRequest(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	fPR, _, err := f.client.CreatePullRequest(
		f.options.Owner, f.options.Repo,
		forgejo.CreatePullRequestOption{
			Title: pr.Title,
			Head:  pr.Head,
			Base:  f.options.BaseBranch,
			Body:  pr.Description,
		},
	)
	if err != nil {
		return err
	}

	// TODO: String ID?
	pr.ID = int(fPR.Index)

	err = f.SetPullRequestLabels(ctx, pr, []releasepr.Label{}, pr.Labels)
	if err != nil {
		return err
	}

	return nil
}

func (f *Forgejo) UpdatePullRequest(_ context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := f.client.EditPullRequest(
		f.options.Owner, f.options.Repo,
		int64(pr.ID), forgejo.EditPullRequestOption{
			Title: pr.Title,
			Body:  pr.Description,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *Forgejo) SetPullRequestLabels(_ context.Context, pr *releasepr.ReleasePullRequest, remove, add []releasepr.Label) error {
	allLabels, err := all(
		func(listOptions forgejo.ListOptions) ([]*forgejo.Label, *forgejo.Response, error) {
			return f.client.ListRepoLabels(f.options.Owner, f.options.Repo, forgejo.ListLabelsOptions{ListOptions: listOptions})
		},
	)
	if err != nil {
		return err
	}

	findLabel := func(labelName string) *forgejo.Label {
		for _, fLabel := range allLabels {
			if fLabel.Name == labelName {
				return fLabel
			}
		}

		return nil
	}

	for _, label := range remove {
		fLabel := findLabel(label.Name)
		if fLabel == nil {
			return fmt.Errorf("unable to remove label %q, not found in API", label.Name)
		}

		_, err = f.client.DeleteIssueLabel(
			f.options.Owner, f.options.Repo,
			int64(pr.ID), fLabel.ID,
		)
		if err != nil {
			return err
		}
	}

	addIDs := make([]int64, 0, len(add))
	for _, label := range add {
		fLabel := findLabel(label.Name)
		if fLabel == nil {
			return fmt.Errorf("unable to add label %q, not found in API", label.Name)
		}

		addIDs = append(addIDs, fLabel.ID)
	}

	_, _, err = f.client.AddIssueLabels(
		f.options.Owner, f.options.Repo,
		int64(pr.ID), forgejo.IssueLabelsOption{Labels: addIDs},
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *Forgejo) ClosePullRequest(_ context.Context, pr *releasepr.ReleasePullRequest) error {
	_, _, err := f.client.EditPullRequest(
		f.options.Owner, f.options.Repo,
		int64(pr.ID), forgejo.EditPullRequestOption{
			State: pointer.Pointer(forgejo.StateClosed),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *Forgejo) PendingReleases(_ context.Context, pendingLabel releasepr.Label) ([]*releasepr.ReleasePullRequest, error) {
	fPRs, err := all(func(listOptions forgejo.ListOptions) ([]*forgejo.PullRequest, *forgejo.Response, error) {
		return f.client.ListRepoPullRequests(
			f.options.Owner, f.options.Repo,
			forgejo.ListPullRequestsOptions{
				// Filtering by Label ID is possible in the API, but not implemented in the Go SDK.
				State:       forgejo.StateClosed,
				ListOptions: listOptions,
			})
	})
	if err != nil {
		// "The target couldn't be found." means that the repo does not have pull requests activated.
		return nil, err
	}

	prs := make([]*releasepr.ReleasePullRequest, 0, len(fPRs))

	for _, pr := range fPRs {
		pending := slices.ContainsFunc(pr.Labels, func(l *forgejo.Label) bool {
			return l.Name == pendingLabel.Name
		})
		if !pending {
			continue
		}

		// pr.Merged is always nil :(
		if !pr.HasMerged {
			// Closed and not merged
			continue
		}

		prs = append(prs, forgejoPRToReleasePullRequest(pr))
	}

	return prs, nil
}

func (f *Forgejo) CreateRelease(_ context.Context, commit git.Commit, title, changelog string, preRelease, latest bool) error {
	// latest can not be set through the API

	_, _, err := f.client.CreateRelease(
		f.options.Owner, f.options.Repo,
		forgejo.CreateReleaseOption{
			TagName:      title,
			Target:       commit.Hash,
			Title:        title,
			Note:         changelog,
			IsPrerelease: preRelease,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func all[T any](f func(listOptions forgejo.ListOptions) ([]T, *forgejo.Response, error)) ([]T, error) {
	results := make([]T, 0)
	page := 1

	for {
		pageResults, resp, err := f(forgejo.ListOptions{Page: page})
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

func forgejoPRToPullRequest(pr *forgejo.PullRequest) *git.PullRequest {
	return &git.PullRequest{
		ID:          int(pr.Index),
		Title:       pr.Title,
		Description: pr.Body,
	}
}

func forgejoPRToReleasePullRequest(pr *forgejo.PullRequest) *releasepr.ReleasePullRequest {
	labels := make([]releasepr.Label, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		labelName := label.Name
		if i := slices.IndexFunc(releasepr.KnownLabels, func(label releasepr.Label) bool {
			return label.Name == labelName
		}); i >= 0 {
			labels = append(labels, releasepr.KnownLabels[i])
		}
	}

	var releaseCommit *git.Commit
	if pr.MergedCommitID != nil {
		releaseCommit = &git.Commit{Hash: *pr.MergedCommitID}
	}

	return &releasepr.ReleasePullRequest{
		PullRequest: *forgejoPRToPullRequest(pr),
		Labels:      labels,

		Head:          pr.Head.Ref,
		ReleaseCommit: releaseCommit,
	}
}

func (g *Options) autodiscover() {
	// TODO
}

func (g *Options) ClientOptions() []forgejo.ClientOption {
	options := []forgejo.ClientOption{}

	if g.APIToken != "" {
		options = append(options, forgejo.SetToken(g.APIToken))
	}

	return options
}

type Options struct {
	forge.Options

	Owner string
	Repo  string

	APIURL   string
	Username string
	APIToken string
}

func New(log *slog.Logger, options *Options) (*Forgejo, error) {
	options.autodiscover()

	client, err := forgejo.NewClient(options.APIURL, options.ClientOptions()...)
	if err != nil {
		return nil, err
	}

	client.SetUserAgent("releaser-pleaser")

	f := &Forgejo{
		options: options,

		client: client,
		log:    log.With("forge", "forgejo"),
	}

	return f, nil
}
