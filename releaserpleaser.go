package rp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/apricote/releaser-pleaser/internal/changelog"
	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
	"github.com/apricote/releaser-pleaser/internal/updater"
	"github.com/apricote/releaser-pleaser/internal/versioning"
)

const (
	PullRequestBranchFormat = "releaser-pleaser--branches--%s"
)

const (
	PullRequestConflictAttempts = 3
)

var (
	ErrorPullRequestConflict = errors.New("conflict: pull request description was changed while releaser-pleaser was running")
)

type ReleaserPleaser struct {
	forge        forge.Forge
	logger       *slog.Logger
	targetBranch string
	commitParser commitparser.CommitParser
	versioning   versioning.Strategy
	extraFiles   []string
	updaters     []updater.Updater
}

func New(forge forge.Forge, logger *slog.Logger, targetBranch string, commitParser commitparser.CommitParser, versioningStrategy versioning.Strategy, extraFiles []string, updaters []updater.Updater) *ReleaserPleaser {
	return &ReleaserPleaser{
		forge:        forge,
		logger:       logger,
		targetBranch: targetBranch,
		commitParser: commitParser,
		versioning:   versioningStrategy,
		extraFiles:   extraFiles,
		updaters:     updaters,
	}
}

func (rp *ReleaserPleaser) EnsureLabels(ctx context.Context) error {
	// TODO: Wrap Error

	return rp.forge.EnsureLabelsExist(ctx, releasepr.KnownLabels)
}

func (rp *ReleaserPleaser) Run(ctx context.Context) error {
	err := rp.runOnboarding(ctx)
	if err != nil {
		return fmt.Errorf("failed to onboard repository: %w", err)
	}

	err = rp.runCreatePendingReleases(ctx)
	if err != nil {
		return fmt.Errorf("failed to create pending releases: %w", err)
	}

	err = rp.runReconcileReleasePRWithRetries(ctx)
	if err != nil {
		return fmt.Errorf("failed to reconcile release pull request: %w", err)
	}

	return nil
}

func (rp *ReleaserPleaser) runOnboarding(ctx context.Context) error {
	err := rp.EnsureLabels(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure all labels exist: %w", err)
	}

	return nil
}

func (rp *ReleaserPleaser) runCreatePendingReleases(ctx context.Context) error {
	logger := rp.logger.With("method", "runCreatePendingReleases")

	logger.InfoContext(ctx, "checking for pending releases")
	prs, err := rp.forge.PendingReleases(ctx, releasepr.LabelReleasePending)
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		logger.InfoContext(ctx, "No pending releases found")
		return nil
	}

	logger.InfoContext(ctx, "Found pending releases", "length", len(prs))

	for _, pr := range prs {
		err = rp.createPendingRelease(ctx, pr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rp *ReleaserPleaser) createPendingRelease(ctx context.Context, pr *releasepr.ReleasePullRequest) error {
	logger := rp.logger.With(
		"method", "createPendingRelease",
		"pr.id", pr.ID,
		"pr.title", pr.Title)

	if pr.ReleaseCommit == nil {
		return fmt.Errorf("pull request is missing the merge commit")
	}

	logger.Info("Creating release", "commit.hash", pr.ReleaseCommit.Hash)

	version, err := pr.Version()
	if err != nil {
		return err
	}

	changelogText, err := pr.ChangelogText()
	if err != nil {
		return err
	}

	// TODO: Check if version should be marked latest

	logger.DebugContext(ctx, "Creating release on forge")
	err = rp.forge.CreateRelease(ctx, *pr.ReleaseCommit, version, changelogText, rp.versioning.IsPrerelease(version), true)
	if err != nil {
		return fmt.Errorf("failed to create release on forge: %w", err)
	}
	logger.DebugContext(ctx, "created release", "release.title", version, "release.url", rp.forge.ReleaseURL(version))

	logger.DebugContext(ctx, "updating pr labels")
	err = rp.forge.SetPullRequestLabels(ctx, pr, []releasepr.Label{releasepr.LabelReleasePending}, []releasepr.Label{releasepr.LabelReleaseTagged})
	if err != nil {
		return err
	}
	logger.DebugContext(ctx, "updated pr labels")

	logger.InfoContext(ctx, "Created release", "release.title", version, "release.url", rp.forge.ReleaseURL(version))

	return nil
}

// runReconcileReleasePRWithRetries retries runReconcileReleasePR up to PullRequestConflictAttempts times, but only
// when a ErrorPullRequestConflict was encountered.
func (rp *ReleaserPleaser) runReconcileReleasePRWithRetries(ctx context.Context) error {
	logger := rp.logger.With("method", "runReconcileReleasePRWithRetries", "totalAttempts", PullRequestConflictAttempts)
	var err error

	for i := range PullRequestConflictAttempts {
		logger := logger.With("attempt", i+1)
		logger.DebugContext(ctx, "attempting runReconcileReleasePR")

		err = rp.runReconcileReleasePR(ctx)
		if err != nil {
			if errors.Is(err, ErrorPullRequestConflict) {
				logger.WarnContext(ctx, "detected conflict while updating pull request description, retrying")
				continue
			}

			break
		}

		break
	}

	if err != nil {
		return err
	}

	return nil
}

func (rp *ReleaserPleaser) runReconcileReleasePR(ctx context.Context) error {
	logger := rp.logger.With("method", "runReconcileReleasePR")

	rpBranch := fmt.Sprintf(PullRequestBranchFormat, rp.targetBranch)

	pr, err := rp.forge.PullRequestForBranch(ctx, rpBranch)
	if err != nil {
		return err
	}

	var releaseOverrides releasepr.ReleaseOverrides

	if pr != nil {
		logger = logger.With("pr.id", pr.ID, "pr.title", pr.Title)
		logger.InfoContext(ctx, "found existing release pull request")

		releaseOverrides, err = pr.GetOverrides()
		if err != nil {
			return err
		}
	}

	releases, err := rp.forge.LatestTags(ctx)
	if err != nil {
		return err
	}

	if releases.Latest != nil {
		logger.InfoContext(ctx, "found latest tag", "tag.hash", releases.Latest.Hash, "tag.name", releases.Latest.Name)
		if releases.Stable != nil && releases.Latest.Hash != releases.Stable.Hash {
			logger.InfoContext(ctx, "found stable tag", "tag.hash", releases.Stable.Hash, "tag.name", releases.Stable.Name)
		}
	} else {
		logger.InfoContext(ctx, "no latest tag found")
	}

	// For stable releases, we want to consider all changes since the last stable release for version and changelog.
	// For prereleases, we want to consider all changes...
	// - since the last stable release for the version
	// - since the latest release (stable or prerelease) for the changelog
	analyzedCommitsForVersioning, err := rp.analyzedCommitsSince(ctx, releases.Stable)
	if err != nil {
		return err
	}

	if len(analyzedCommitsForVersioning) == 0 {
		if pr != nil {
			logger.InfoContext(ctx, "closing existing pull requests, no commits available", "pr.id", pr.ID, "pr.title", pr.Title)
			err = rp.forge.ClosePullRequest(ctx, pr)
			if err != nil {
				return err
			}
		} else {
			logger.InfoContext(ctx, "No commits available for release")
		}

		return nil
	}

	versionBump := versioning.BumpFromCommits(analyzedCommitsForVersioning)
	// TODO: Set version in release pr
	nextVersion, err := rp.versioning.NextVersion(releases, versionBump, releaseOverrides.NextVersionType)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "next version", "version", nextVersion)

	analyzedCommitsForChangelog := analyzedCommitsForVersioning
	if releaseOverrides.NextVersionType.IsPrerelease() && releases.Latest != releases.Stable {
		analyzedCommitsForChangelog, err = rp.analyzedCommitsSince(ctx, releases.Latest)
		if err != nil {
			return err
		}
	}

	logger.DebugContext(ctx, "cloning repository", "clone.url", rp.forge.CloneURL())
	repo, err := git.CloneRepo(ctx, logger, rp.forge.CloneURL(), rp.targetBranch, rp.forge.GitAuth())
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	if err = repo.DeleteBranch(ctx, rpBranch); err != nil {
		return err
	}

	if err = repo.Checkout(ctx, rpBranch); err != nil {
		return err
	}

	changelogData := changelog.New(commitparser.ByType(analyzedCommitsForChangelog), nextVersion, rp.forge.ReleaseURL(nextVersion), releaseOverrides.Prefix, releaseOverrides.Suffix)

	changelogEntry, err := changelog.Entry(logger, changelog.DefaultTemplate(), changelogData, changelog.Formatting{})
	if err != nil {
		return fmt.Errorf("failed to build changelog entry: %w", err)
	}

	// Info for updaters
	info := updater.ReleaseInfo{Version: nextVersion, ChangelogEntry: changelogEntry}

	for _, u := range rp.updaters {
		for _, file := range u.Files() {
			err = repo.UpdateFile(ctx, file, u.CreateNewFiles(), u.Update(info))
			if err != nil {
				return fmt.Errorf("failed to run updater %T: %w", u, err)
			}
		}
	}

	releaseCommitAuthor, err := rp.forge.CommitAuthor(ctx)
	if err != nil {
		return fmt.Errorf("failed to get commit author: %w", err)
	}

	releaseCommitMessage := fmt.Sprintf("chore(%s): release %s", rp.targetBranch, nextVersion)
	releaseCommit, err := repo.Commit(ctx, releaseCommitMessage, releaseCommitAuthor)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.InfoContext(ctx, "created release commit", "commit.hash", releaseCommit.Hash, "commit.message", releaseCommit.Message, "commit.author", releaseCommitAuthor)

	// Check if anything changed in comparison to the remote branch (if exists)
	newReleasePRChanges, err := repo.HasChangesWithRemote(ctx, rp.targetBranch, rpBranch)
	if err != nil {
		return err
	}

	if newReleasePRChanges {
		err = repo.ForcePush(ctx, rpBranch)
		if err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}

		logger.InfoContext(ctx, "pushed branch", "commit.hash", releaseCommit.Hash, "branch.name", rpBranch)
	} else {
		logger.InfoContext(ctx, "file content is already up-to-date in remote branch, skipping push")
	}

	// We do not need the version title here. In the pull request the version is available from the title, and in the
	// release on the Forge its usually in a heading somewhere above the text.
	changelogEntryPullRequest, err := changelog.Entry(logger, changelog.DefaultTemplate(), changelogData, changelog.Formatting{HideVersionTitle: true})
	if err != nil {
		return fmt.Errorf("failed to build pull request changelog entry: %w", err)
	}

	// Open/Update PR
	if pr == nil {
		pr, err = releasepr.NewReleasePullRequest(rpBranch, rp.targetBranch, nextVersion, changelogEntryPullRequest)
		if err != nil {
			return err
		}

		err = rp.forge.CreatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "opened pull request", "pr.title", pr.Title, "pr.id", pr.ID, "pr.url", rp.forge.PullRequestURL(pr.ID))
	} else {
		// Check if the pull request was updated while releaser-pleaser was running.
		// This avoids a conflict where the user updated the PR while releaser-pleaser already pulled the info, and
		// releaser-pleaser subsequently reverts the users changes. There is still a minimal time window for this to
		// happen between us checking the PR again and submitting our changes.

		logger.DebugContext(ctx, "checking for conflict in pr description", "pr.id", pr.ID)
		recheckPR, err := rp.forge.PullRequestForBranch(ctx, rpBranch)
		if err != nil {
			return err
		}
		if recheckPR == nil {
			return fmt.Errorf("PR was deleted while releaser-pleaser was running")
		}
		if recheckPR.Description != pr.Description {
			return ErrorPullRequestConflict
		}

		pr.SetTitle(rp.targetBranch, nextVersion)

		overrides, err := pr.GetOverrides()
		if err != nil {
			return err
		}
		err = pr.SetDescription(changelogEntryPullRequest, overrides)
		if err != nil {
			return err
		}

		err = rp.forge.UpdatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "updated pull request", "pr.title", pr.Title, "pr.id", pr.ID, "pr.url", rp.forge.PullRequestURL(pr.ID))
	}

	return nil
}

func (rp *ReleaserPleaser) analyzedCommitsSince(ctx context.Context, since *git.Tag) ([]commitparser.AnalyzedCommit, error) {
	logger := rp.logger.With("method", "analyzedCommitsSince")

	if since != nil {
		logger = rp.logger.With("tag.hash", since.Hash, "tag.name", since.Name)
	}

	commits, err := rp.forge.CommitsSince(ctx, since)
	if err != nil {
		return nil, err
	}

	commits, err = parsePRBodyForCommitOverrides(commits)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Found releasable commits", "length", len(commits))

	analyzedCommits, err := rp.commitParser.Analyze(commits)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Analyzed commits", "length", len(analyzedCommits))

	return analyzedCommits, nil
}
