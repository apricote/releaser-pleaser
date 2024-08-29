package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
	"github.com/apricote/releaser-pleaser/internal/commitparser/conventionalcommits"
	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/forge/github"
	"github.com/apricote/releaser-pleaser/internal/forge/gitlab"
	"github.com/apricote/releaser-pleaser/internal/updater"
	"github.com/apricote/releaser-pleaser/internal/versioning"
)

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

	runCmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	runCmd.PersistentFlags().StringVar(&flagBranch, "branch", "main", "")
	runCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "")
	runCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
	runCmd.PersistentFlags().StringVar(&flagExtraFiles, "extra-files", "", "")
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	var err error

	logger.DebugContext(ctx, "run called",
		"forge", flagForge,
		"branch", flagBranch,
		"owner", flagOwner,
		"repo", flagRepo,
	)

	var f forge.Forge

	forgeOptions := forge.Options{
		Repository: flagRepo,
		BaseBranch: flagBranch,
	}

	switch flagForge {
	case "gitlab":
		logger.DebugContext(ctx, "using forge GitLab")
		f, err = gitlab.New(logger, &gitlab.Options{
			Options: forgeOptions,
			Path:    flagOwner,
			Repo:    flagRepo,
		})
		if err != nil {
			logger.ErrorContext(ctx, "failed to create client", "err", err)
			return fmt.Errorf("failed to create gitlab client: %w", err)
		}
	case "github":
		logger.DebugContext(ctx, "using forge GitHub")
		f = github.New(logger, &github.Options{
			Options: forgeOptions,
			Owner:   flagOwner,
			Repo:    flagRepo,
		})
	default:
		return fmt.Errorf("unknown --forge: %s", flagForge)
	}

	extraFiles := parseExtraFiles(flagExtraFiles)

	releaserPleaser := rp.New(
		f,
		logger,
		flagBranch,
		conventionalcommits.NewParser(logger),
		versioning.SemVerNextVersion,
		extraFiles,
		[]updater.NewUpdater{updater.Generic},
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
