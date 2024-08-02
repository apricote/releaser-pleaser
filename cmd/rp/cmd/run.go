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

func run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var f rp.Forge

	forgeOptions := rp.ForgeOptions{
		Repository: flagRepo,
		BaseBranch: flagBranch,
	}

	switch flagForge {
	//case "gitlab":
	//f = rp.NewGitLab(forgeOptions)
	case "github":
		f = rp.NewGitHub(logger, &rp.GitHubOptions{
			ForgeOptions: forgeOptions,
			Owner:        flagOwner,
			Repo:         flagRepo,
		})
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
		return err
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
		return err
	}

	err = rp.RunUpdater(ctx, nextVersion, worktree)
	if err != nil {
		return err
	}

	changelogEntry, err := rp.NewChangelogEntry(changesets, nextVersion, forge.ReleaseURL(nextVersion))
	if err != nil {
		return err
	}

	err = rp.UpdateChangelogFile(worktree, changelogEntry)
	if err != nil {
		return err
	}

	releaseCommitMessage := fmt.Sprintf("chore(%s): release %s", flagBranch, nextVersion)
	releaseCommit, err := worktree.Commit(releaseCommitMessage, &git.CommitOptions{})
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "created release commit", "commit.hash", releaseCommit.String(), "commit.message", releaseCommitMessage)

	// TODO: Check if there is a diff between forge/rpBranch..rpBranch..forge/rpBranch and only push if there are changes
	// To reduce wasted CI cycles

	pushRefSpec := config.RefSpec(fmt.Sprintf(
		"+%s:%s",
		rpBranchRef,
		// This needs to be the local branch name, not the remotes/origin ref
		// See https://stackoverflow.com/a/75727620
		rpBranchRef,
	))
	logger.DebugContext(ctx, "pushing branch", "commit.hash", releaseCommit.String(), "branch.name", rpBranch, "refspec", pushRefSpec.String())
	if err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: rp.GitRemoteName,
		RefSpecs:   []config.RefSpec{pushRefSpec},
		Force:      true,
		Auth:       forge.GitAuth(),
	}); err != nil {
		return err
	}

	logger.InfoContext(ctx, "pushed branch", "commit.hash", releaseCommit.String(), "branch.name", rpBranch, "refspec", pushRefSpec.String())

	// TODO Open PR

	return nil
}
