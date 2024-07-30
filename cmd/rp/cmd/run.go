/*
Copyright © 2024 Julian Tölle
*/
package cmd

import (
	"context"
	"fmt"

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

	changesets, err := getChangesetsFromForge(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to get changesets: %w", err)
	}

	err = reconcileReleasePR(ctx, f, changesets)
	if err != nil {
		return fmt.Errorf("failed to reconcile release pr: %w", err)
	}

	return nil
}

func getChangesetsFromForge(ctx context.Context, forge rp.Forge) ([]rp.Changeset, error) {
	tag, err := forge.LatestTag(ctx)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Latest Tag", "tag.hash", tag.Hash, "tag.name", tag.Name)

	releasableCommits, err := forge.CommitsSince(ctx, tag)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Found releasable commits", "length", len(releasableCommits))

	changesets, err := forge.Changesets(ctx, releasableCommits)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, "Found changesets", "length", len(changesets))

	return changesets, nil
}

func reconcileReleasePR(ctx context.Context, forge rp.Forge, changesets []rp.Changeset) error {
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
		logger.InfoContext(ctx, "found existing release pull request: %d: %s", pr.ID, pr.Title)
	}

	releaseOverrides, err := pr.GetOverrides()
	if err != nil {
		return err
	}

	// ...

	return nil
}
