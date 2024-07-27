/*
Copyright © 2024 Julian Tölle
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
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

	tag, err := f.LatestTag(ctx)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Latest Tag", "tag.hash", tag.Hash, "tag.name", tag.Name)

	releaseableCommits, err := f.CommitsSince(ctx, tag)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Found releasable commits", "length", len(releaseableCommits))

	changesets, err := f.Changesets(ctx, releaseableCommits)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "Found changesets", "length", len(changesets))

	for _, changeset := range changesets {
		fmt.Printf("%s %s\n", changeset.Identifier, changeset.URL)
		for _, entry := range changeset.ChangelogEntries {
			fmt.Printf("  - %s %s\n", entry.Hash, entry.Description)
		}
	}
	fmt.Printf("Previous Tag: %s\n", tag.Name)

	return nil
}
