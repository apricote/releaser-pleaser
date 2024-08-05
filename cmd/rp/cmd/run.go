package cmd

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
)

const (
	RELEASER_PLEASER_BRANCH = "releaser-pleaser--branches--%s"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:  "run",
	RunE: run,
}

var (
	flagForge  string
	flagBranch string
	flagOwner  string
	flagRepo   string
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	runCmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	runCmd.PersistentFlags().StringVar(&flagBranch, "branch", "main", "")
	runCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "")
	runCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	logger.DebugContext(ctx, "run called",
		"forge", flagForge,
		"branch", flagBranch,
		"owner", flagOwner,
		"repo", flagRepo,
	)

	var f rp.Forge

	forgeOptions := rp.ForgeOptions{
		Repository: flagRepo,
		BaseBranch: flagBranch,
	}

	switch flagForge {
	//case "gitlab":
	//f = rp.NewGitLab(forgeOptions)
	case "github":
		logger.DebugContext(ctx, "using forge GitHub")
		f = rp.NewGitHub(logger, &rp.GitHubOptions{
			ForgeOptions: forgeOptions,
			Owner:        flagOwner,
			Repo:         flagRepo,
		})
	}

	err := ensureLabels(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to ensure all labels exist: %w", err)
	}

	err = createPendingReleases(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to create pending releases: %w", err)
	}

	changesets, releases, err := getChangesetsFromForge(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to get changesets: %w", err)
	}

	err = reconcileReleasePR(ctx, f, changesets, releases)
	if err != nil {
		return fmt.Errorf("failed to reconcile release pr: %w", err)
	}

	return nil
}

func ensureLabels(ctx context.Context, forge rp.Forge) error {
	return forge.EnsureLabelsExist(ctx, rp.Labels)
}

func createPendingReleases(ctx context.Context, forge rp.Forge) error {
	logger.InfoContext(ctx, "checking for pending releases")
	prs, err := forge.PendingReleases(ctx)
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		logger.InfoContext(ctx, "No pending releases found")
		return nil
	}

	logger.InfoContext(ctx, "Found pending releases", "length", len(prs))

	for _, pr := range prs {
		err = createPendingRelease(ctx, forge, pr)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPendingRelease(ctx context.Context, forge rp.Forge, pr *rp.ReleasePullRequest) error {
	logger := logger.With("pr.id", pr.ID, "pr.title", pr.Title)

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
	err = forge.CreateRelease(ctx, *pr.ReleaseCommit, version, changelog, false, true)
	if err != nil {
		return fmt.Errorf("failed to create release on forge: %w", err)
	}
	logger.DebugContext(ctx, "created release", "release.title", version, "release.url", forge.ReleaseURL(version))

	logger.DebugContext(ctx, "updating pr labels")
	err = forge.SetPullRequestLabels(ctx, pr, []string{rp.LabelReleasePending}, []string{rp.LabelReleaseTagged})
	if err != nil {
		return err
	}
	logger.DebugContext(ctx, "updated pr labels")

	logger.InfoContext(ctx, "Created release", "release.title", version, "release.url", forge.ReleaseURL(version))

	return nil
}

func getChangesetsFromForge(ctx context.Context, forge rp.Forge) ([]rp.Changeset, rp.Releases, error) {
	releases, err := forge.LatestTags(ctx)
	if err != nil {
		return nil, rp.Releases{}, err
	}

	if releases.Latest != nil {
		logger.InfoContext(ctx, "found latest tag", "tag.hash", releases.Latest.Hash, "tag.name", releases.Latest.Name)
		if releases.Stable != nil && releases.Latest.Hash != releases.Stable.Hash {
			logger.InfoContext(ctx, "found stable tag", "tag.hash", releases.Stable.Hash, "tag.name", releases.Stable.Name)
		}
	} else {
		logger.InfoContext(ctx, "no latest tag found")
	}

	releasableCommits, err := forge.CommitsSince(ctx, releases.Stable)
	if err != nil {
		return nil, rp.Releases{}, err
	}

	logger.InfoContext(ctx, "Found releasable commits", "length", len(releasableCommits))

	changesets, err := forge.Changesets(ctx, releasableCommits)
	if err != nil {
		return nil, rp.Releases{}, err
	}

	logger.InfoContext(ctx, "Found changesets", "length", len(changesets))

	return changesets, releases, nil
}

func reconcileReleasePR(ctx context.Context, forge rp.Forge, changesets []rp.Changeset, releases rp.Releases) error {
	rpBranch := fmt.Sprintf(RELEASER_PLEASER_BRANCH, flagBranch)
	rpBranchRef := plumbing.NewBranchReferenceName(rpBranch)
	// Check Forge for open PR
	// Get any modifications from open PR
	// Clone Repo
	// Run Updaters + Changelog
	// Upsert PR
	pr, err := forge.PullRequestForBranch(ctx, fmt.Sprintf(RELEASER_PLEASER_BRANCH, flagBranch))
	if err != nil {
		return err
	}

	if pr != nil {
		logger.InfoContext(ctx, "found existing release pull request", "pr.id", pr.ID, "pr.title", pr.Title)
	}

	if len(changesets) == 0 {
		if pr != nil {
			logger.InfoContext(ctx, "closing existing pull requests, no changesets available", "pr.id", pr.ID, "pr.title", pr.Title)
			err = forge.ClosePullRequest(ctx, pr)
			if err != nil {
				return err
			}
		} else {
			logger.InfoContext(ctx, "No changesets available for release")
		}

		return nil
	}

	var releaseOverrides rp.ReleaseOverrides
	if pr != nil {
		releaseOverrides, err = pr.GetOverrides()
		if err != nil {
			return err
		}
	}

	versionBump := rp.VersionBumpFromChangesets(changesets)
	nextVersion, err := releases.NextVersion(versionBump, releaseOverrides.NextVersionType)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "next version", "version", nextVersion)

	logger.DebugContext(ctx, "cloning repository", "clone.url", forge.CloneURL())
	repo, err := rp.CloneRepo(ctx, forge.CloneURL(), flagBranch, forge.GitAuth())
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

	err = rp.RunUpdater(ctx, nextVersion, worktree)
	if err != nil {
		return fmt.Errorf("failed to update files with new version: %w", err)
	}

	changelogEntry, err := rp.NewChangelogEntry(changesets, nextVersion, forge.ReleaseURL(nextVersion), releaseOverrides.Prefix, releaseOverrides.Suffix)
	if err != nil {
		return fmt.Errorf("failed to build changelog entry: %w", err)
	}

	err = rp.UpdateChangelogFile(worktree, changelogEntry)
	if err != nil {
		return fmt.Errorf("failed to update changelog file: %w", err)
	}

	releaseCommitMessage := fmt.Sprintf("chore(%s): release %s", flagBranch, nextVersion)
	releaseCommitHash, err := worktree.Commit(releaseCommitMessage, &git.CommitOptions{
		Author:    rp.GitSignature(),
		Committer: rp.GitSignature(),
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.InfoContext(ctx, "created release commit", "commit.hash", releaseCommitHash.String(), "commit.message", releaseCommitMessage)

	newReleasePRChanges := true

	// Check if anything changed in comparison to the remote branch (if exists)
	if remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName(rp.GitRemoteName, rpBranch), false); err != nil {
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
			RemoteName: rp.GitRemoteName,
			RefSpecs:   []config.RefSpec{pushRefSpec},
			Force:      true,
			Auth:       forge.GitAuth(),
		}); err != nil {
			return err
		}

		logger.InfoContext(ctx, "pushed branch", "commit.hash", releaseCommitHash.String(), "branch.name", rpBranch, "refspec", pushRefSpec.String())
	} else {
		logger.InfoContext(ctx, "file content is already up-to-date in remote branch, skipping push")
	}

	// Open/Update PR
	if pr == nil {
		pr, err = rp.NewReleasePullRequest(rpBranch, flagBranch, nextVersion, changelogEntry)
		if err != nil {
			return err
		}

		err = forge.CreatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "opened pull request", "pr.title", pr.Title, "pr.id", pr.ID)
	} else {
		pr.SetTitle(flagBranch, nextVersion)
		err = pr.SetDescription(changelogEntry)
		if err != nil {
			return err
		}

		err = forge.UpdatePullRequest(ctx, pr)
		if err != nil {
			return err
		}
		logger.InfoContext(ctx, "updated pull request", "pr.title", pr.Title, "pr.id", pr.ID)
	}

	return nil
}
