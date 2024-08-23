package rp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	PullRequestBranchFormat = "releaser-pleaser--branches--%s"
)

type ReleaserPleaser struct {
	forge        Forge
	logger       *slog.Logger
	targetBranch string
	commitParser CommitParser
	nextVersion  VersioningStrategy
	extraFiles   []string
	updaters     []Updater
}

func New(forge Forge, logger *slog.Logger, targetBranch string, commitParser CommitParser, versioningStrategy VersioningStrategy, extraFiles []string, updaters []Updater) *ReleaserPleaser {
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
	return rp.forge.EnsureLabelsExist(ctx, KnownLabels)
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
	prs, err := rp.forge.PendingReleases(ctx, LabelReleasePending)
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

func (rp *ReleaserPleaser) createPendingRelease(ctx context.Context, pr *ReleasePullRequest) error {
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
	err = rp.forge.SetPullRequestLabels(ctx, pr, []Label{LabelReleasePending}, []Label{LabelReleaseTagged})
	if err != nil {
		return err
	}
	logger.DebugContext(ctx, "updated pr labels")

	logger.InfoContext(ctx, "Created release", "release.title", version, "release.url", rp.forge.ReleaseURL(version))

	return nil
}

func (rp *ReleaserPleaser) runReconcileReleasePR(ctx context.Context) error {
	logger := rp.logger.With("method", "runReconcileReleasePR")

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

	releasableCommits, err := rp.forge.CommitsSince(ctx, releases.Stable)
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

	rpBranch := fmt.Sprintf(PullRequestBranchFormat, rp.targetBranch)
	rpBranchRef := plumbing.NewBranchReferenceName(rpBranch)
	// Check Forge for open PR
	// Get any modifications from open PR
	// Clone Repo
	// Run Updaters + Changelog
	// Upsert PR
	pr, err := rp.forge.PullRequestForBranch(ctx, rpBranch)
	if err != nil {
		return err
	}

	if pr != nil {
		logger.InfoContext(ctx, "found existing release pull request", "pr.id", pr.ID, "pr.title", pr.Title)
	}

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

	var releaseOverrides ReleaseOverrides
	if pr != nil {
		releaseOverrides, err = pr.GetOverrides()
		if err != nil {
			return err
		}
	}

	versionBump := VersionBumpFromCommits(analyzedCommits)
	// TODO: Set version in release pr
	nextVersion, err := rp.nextVersion(releases, versionBump, releaseOverrides.NextVersionType)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "next version", "version", nextVersion)

	logger.DebugContext(ctx, "cloning repository", "clone.url", rp.forge.CloneURL())
	repo, err := CloneRepo(ctx, rp.forge.CloneURL(), rp.targetBranch, rp.forge.GitAuth())
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if branch, _ := repo.Branch(rpBranch); branch != nil {
		logger.DebugContext(ctx, "deleting previous releaser-pleaser branch locally", "branch.name", rpBranch)
		if err = repo.DeleteBranch(rpBranch); err != nil {
			return err
		}
	}

	if err = worktree.Checkout(&git.CheckoutOptions{
		Branch: rpBranchRef,
		Create: true,
	}); err != nil {
		return fmt.Errorf("failed to check out branch: %w", err)
	}

	changelogEntry, err := NewChangelogEntry(analyzedCommits, nextVersion, rp.forge.ReleaseURL(nextVersion), releaseOverrides.Prefix, releaseOverrides.Suffix)
	if err != nil {
		return fmt.Errorf("failed to build changelog entry: %w", err)
	}

	// Info for updaters
	info := ReleaseInfo{Version: nextVersion, ChangelogEntry: changelogEntry}

	err = UpdateChangelogFile(worktree, changelogEntry)
	if err != nil {
		return fmt.Errorf("failed to update changelog file: %w", err)
	}

	updateFile := func(path string, updaters []Updater) error {
		file, err := worktree.Filesystem.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			return err
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		updatedContent := string(content)

		for _, updater := range updaters {
			updatedContent, err = updater.UpdateContent(updatedContent, info)
			if err != nil {
				return fmt.Errorf("failed to run updater %T on file %s", updater, path)
			}
		}

		err = file.Truncate(0)
		if err != nil {
			return fmt.Errorf("failed to replace file content: %w", err)
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to replace file content: %w", err)
		}
		_, err = file.Write([]byte(updatedContent))
		if err != nil {
			return fmt.Errorf("failed to replace file content: %w", err)
		}

		_, err = worktree.Add(path)
		if err != nil {
			return fmt.Errorf("failed to add updated file to git worktree: %w", err)
		}

		return nil
	}

	for _, path := range rp.extraFiles {
		_, err = worktree.Filesystem.Stat(path)
		if err != nil {
			// TODO: Check for non existing file or dirs
			return fmt.Errorf("failed to run file updater because the file %s does not exist: %w", path, err)
		}

		err = updateFile(path, rp.updaters)
		if err != nil {
			return fmt.Errorf("failed to run file updater: %w", err)
		}
	}

	releaseCommitMessage := fmt.Sprintf("chore(%s): release %s", rp.targetBranch, nextVersion)
	releaseCommitHash, err := worktree.Commit(releaseCommitMessage, &git.CommitOptions{
		Author:    GitSignature(),
		Committer: GitSignature(),
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.InfoContext(ctx, "created release commit", "commit.hash", releaseCommitHash.String(), "commit.message", releaseCommitMessage)

	newReleasePRChanges := true

	// Check if anything changed in comparison to the remote branch (if exists)
	if remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName(GitRemoteName, rpBranch), false); err != nil {
		if err.Error() != "reference not found" {
			// "reference not found" is expected and we should always push
			return err
		}
	} else {
		remoteCommit, err := repo.CommitObject(remoteRef.Hash())
		if err != nil {
			return err
		}

		localCommit, err := repo.CommitObject(releaseCommitHash)
		if err != nil {
			return err
		}

		diff, err := localCommit.PatchContext(ctx, remoteCommit)
		if err != nil {
			return err
		}

		newReleasePRChanges = len(diff.FilePatches()) > 0
	}

	if newReleasePRChanges {
		pushRefSpec := config.RefSpec(fmt.Sprintf(
			"+%s:%s",
			rpBranchRef,
			// This needs to be the local branch name, not the remotes/origin ref
			// See https://stackoverflow.com/a/75727620
			rpBranchRef,
		))
		logger.DebugContext(ctx, "pushing branch", "commit.hash", releaseCommitHash.String(), "branch.name", rpBranch, "refspec", pushRefSpec.String())
		if err = repo.PushContext(ctx, &git.PushOptions{
			RemoteName: GitRemoteName,
			RefSpecs:   []config.RefSpec{pushRefSpec},
			Force:      true,
			Auth:       rp.forge.GitAuth(),
		}); err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}

		logger.InfoContext(ctx, "pushed branch", "commit.hash", releaseCommitHash.String(), "branch.name", rpBranch, "refspec", pushRefSpec.String())
	} else {
		logger.InfoContext(ctx, "file content is already up-to-date in remote branch, skipping push")
	}

	// Open/Update PR
	if pr == nil {
		pr, err = NewReleasePullRequest(rpBranch, rp.targetBranch, nextVersion, changelogEntry)
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
