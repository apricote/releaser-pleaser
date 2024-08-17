package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:  "run",
	RunE: run,
}

var (
	flagForge      string
	flagBranch     string
	flagOwner      string
	flagRepo       string
	flagExtraFiles string
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	runCmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	runCmd.PersistentFlags().StringVar(&flagBranch, "branch", "main", "")
	runCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "")
	runCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
	runCmd.PersistentFlags().StringVar(&flagExtraFiles, "extra-files", "", "")
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	logger.DebugContext(ctx, "run called",
		"forge", flagForge,
		"branch", flagBranch,
		"owner", flagOwner,
		"repo", flagRepo,
	)

	var forge rp.Forge

	forgeOptions := rp.ForgeOptions{
		Repository: flagRepo,
		BaseBranch: flagBranch,
	}

	switch flagForge {
	//case "gitlab":
	//f = rp.NewGitLab(forgeOptions)
	case "github":
		logger.DebugContext(ctx, "using forge GitHub")
		forge = rp.NewGitHub(logger, &rp.GitHubOptions{
			ForgeOptions: forgeOptions,
			Owner:        flagOwner,
			Repo:         flagRepo,
		})
	}

	extraFiles := parseExtraFiles(flagExtraFiles)

	releaserPleaser := rp.New(
		forge,
		logger,
		flagBranch,
		rp.NewConventionalCommitsParser(),
		rp.SemVerNextVersion,
		extraFiles,
		[]rp.Updater{&rp.GenericUpdater{}},
	)

	return releaserPleaser.Run(ctx)
}

func parseExtraFiles(input string) []string {
	lines := strings.Split(input, "\n")

	extraFiles := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			extraFiles = append(extraFiles, line)
		}
	}

	return extraFiles
}
