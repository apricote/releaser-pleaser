package rp

import (
	"context"
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

type ReleaserPleaser struct {
	forge        forge.Forge
	logger       *slog.Logger
	targetBranch string
	commitParser commitparser.CommitParser
	nextVersion  versioning.Strategy
	extraFiles   []string
	updaters     []updater.NewUpdater
}

func New(forge forge.Forge, logger *slog.Logger, targetBranch string, commitParser commitparser.CommitParser, versioningStrategy versioning.Strategy, extraFiles []string, updaters []updater.NewUpdater) *ReleaserPleaser {
	return &ReleaserPleaser{
		forge:        forge,
		logger:       logger,
		targetBranch: targetBranch,
		commitParser: commitParser,
		nextVersion:  versioningStrategy,
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

	err = rp.runReconcileReleasePR(ctx)
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

	changelog, err := pr.ChangelogText()
	if err != nil {
		return err
	}

	// TODO: pre-release & latest

	logger.DebugContext(ctx, "Creating release on forge")
	err = rp.forge.CreateRelease(ctx, *pr.ReleaseCommit, version, changelog, false, true)
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

	// By default, we want to show everything that has happened since the last stable release
	lastReleaseCommit := releases.Stable
	if releaseOverrides.NextVersionType.IsPrerelease() {
		// if the new release will be a prerelease,
		// only show changes since the latest release (stable or prerelease)
		lastReleaseCommit = releases.Latest
	}

	releasableCommits, err := rp.forge.CommitsSince(ctx, lastReleaseCommit)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Found releasable commits", "length", len(releasableCommits))

	// TODO: Handle commit overrides
	analyzedCommits, err := rp.commitParser.Analyze(releasableCommits)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Analyzed commits", "length", len(analyzedCommits))

	if len(analyzedCommits) == 0 {
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

	versionBump := versioning.BumpFromCommits(analyzedCommits)
	// TODO: Set version in release pr
	nextVersion, err := rp.nextVersion(releases, versionBump, releaseOverrides.NextVersionType)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "next version", "version", nextVersion)

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

	changelogEntry, err := changelog.NewChangelogEntry(analyzedCommits, nextVersion, rp.forge.ReleaseURL(nextVersion), releaseOverrides.Prefix, releaseOverrides.Suffix)
	if err != nil {
		return fmt.Errorf("failed to build changelog entry: %w", err)
	}

	// Info for updaters
	info := updater.ReleaseInfo{Version: nextVersion, ChangelogEntry: changelogEntry}

	err = repo.UpdateFile(ctx, updater.ChangelogFile, updater.WithInfo(info, updater.Changelog))
	if err != nil {
		return fmt.Errorf("failed to update changelog file: %w", err)
	}

	for _, path := range rp.extraFiles {
		// TODO: Check for missing files
		err = repo.UpdateFile(ctx, path, updater.WithInfo(info, rp.updaters...))
		if err != nil {
			return fmt.Errorf("failed to run file updater: %w", err)
		}
	}

	releaseCommitMessage := fmt.Sprintf("chore(%s): release %s", rp.targetBranch, nextVersion)
	releaseCommit, err := repo.Commit(ctx, releaseCommitMessage)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.InfoContext(ctx, "created release commit", "commit.hash", releaseCommit.Hash, "commit.message", releaseCommit.Message)

	// Check if anything changed in comparison to the remote branch (if exists)
	newReleasePRChanges, err := repo.HasChangesWithRemote(ctx, rpBranch)
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

	// Open/Update PR
	if pr == nil {
		pr, err = releasepr.NewReleasePullRequest(rpBranch, rp.targetBranch, nextVersion, changelogEntry)
		if err != nil {
			return err
		}

		err = rp.forge.CreatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "opened pull request", "pr.title", pr.Title, "pr.id", pr.ID)
	} else {
		pr.SetTitle(rp.targetBranch, nextVersion)

		overrides, err := pr.GetOverrides()
		if err != nil {
			return err
		}
		err = pr.SetDescription(changelogEntry, overrides)
		if err != nil {
			return err
		}

		err = rp.forge.UpdatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "updated pull request", "pr.title", pr.Title, "pr.id", pr.ID)
	}

	return nil
}
