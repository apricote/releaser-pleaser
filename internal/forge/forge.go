package forge

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
)

type Forge interface {
	RepoURL() string
	CloneURL() string
	ReleaseURL(version string) string
	PullRequestURL(id int) string

	GitAuth() transport.AuthMethod

	// LatestTags returns the last stable tag created on the main branch. If there is a more recent pre-release tag,
	// that is also returned. If no tag is found, it returns nil.
	LatestTags(context.Context) (git.Releases, error)

	// CommitsSince returns all commits to main branch after the Tag. The tag can be `nil`, in which case this
	// function should return all commits.
	CommitsSince(context.Context, *git.Tag) ([]git.Commit, error)

	// EnsureLabelsExist verifies that all desired labels are available on the repository. If labels are missing, they
	// are created them.
	EnsureLabelsExist(context.Context, []releasepr.Label) error

	// PullRequestForBranch returns the open pull request between the branch and Options.BaseBranch. If no open PR
	// exists, it returns nil.
	PullRequestForBranch(context.Context, string) (*releasepr.ReleasePullRequest, error)

	// CreatePullRequest opens a new pull/merge request for the ReleasePullRequest.
	CreatePullRequest(context.Context, *releasepr.ReleasePullRequest) error

	// UpdatePullRequest updates the pull/merge request identified through the ID of
	// the ReleasePullRequest to the current description and title.
	UpdatePullRequest(context.Context, *releasepr.ReleasePullRequest) error

	// SetPullRequestLabels updates the pull/merge request identified through the ID of
	// the ReleasePullRequest to the current labels.
	SetPullRequestLabels(ctx context.Context, pr *releasepr.ReleasePullRequest, remove, add []releasepr.Label) error

	// ClosePullRequest closes the pull/merge request identified through the ID of
	// the ReleasePullRequest, as it is no longer required.
	ClosePullRequest(context.Context, *releasepr.ReleasePullRequest) error

	// PendingReleases returns a list of ReleasePullRequest. The list should contain all pull/merge requests that are
	// merged and have the matching label.
	PendingReleases(context.Context, releasepr.Label) ([]*releasepr.ReleasePullRequest, error)

	// CreateRelease creates a release on the Forge, pointing at the commit with the passed in details.
	CreateRelease(ctx context.Context, commit git.Commit, title, changelog string, prerelease, latest bool) error
}

type Options struct {
	Repository string
	BaseBranch string
}
