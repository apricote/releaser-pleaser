/*
Copyright © 2024 Julian Tölle
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:  "run",
	RunE: run,
}

var (
	flagForge string
	flagRepo  string
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	runCmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	runCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
}

func run(cmd *cobra.Command, args []string) error {
	var f rp.Forge

	forgeOptions := rp.ForgeOptions{
		Repository: flagRepo,
	}

	switch flagForge {
	case "gitlab":
		f = rp.NewGitLab(forgeOptions)
	case "github":
		f = rp.NewGitHub(forgeOptions)
	}

	log.Println("Repo URL: " + f.RepoURL())

	//repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
	//	URL:          .RepoURL(),
	//	SingleBranch: true,
	//	Depth:        CommitSearchDepth,
	//})
	repo, err := git.PlainOpen("~/git/listory")
	if err != nil {
		return err
	}

	commits, previousTag, err := rp.ReleasableCommits(repo)
	if err != nil {
		return err
	}

	analyzedCommits, versionBump, err := rp.AnalyzeCommits(commits)
	if err != nil {
		return err
	}

	for _, commit := range analyzedCommits {
		title, _, _ := strings.Cut(commit.Message, "\n")
		fmt.Printf("%s %s\n", commit.Hash, title)
	}
	fmt.Printf("Previous Tag: %s\n", previousTag.Name)
	fmt.Printf("Recommended Bump: %v\n", versionBump)

	return nil
}
