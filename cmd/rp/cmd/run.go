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
	flagForge             string
	flagBranch            string
	flagOwner             string
	flagRepo              string
	flagExtraFiles        string
	flagUpdatePackageJson bool
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	runCmd.PersistentFlags().StringVar(&flagBranch, "branch", "main", "")
	runCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "")
	runCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
	runCmd.PersistentFlags().StringVar(&flagExtraFiles, "extra-files", "", "")
	runCmd.PersistentFlags().BoolVar(&flagUpdatePackageJson, "update-package-json", false, "")
}

func run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	var err error

	logger.DebugContext(ctx, "run called",
		"forge", flagForge,
		"branch", flagBranch,
		"owner", flagOwner,
		"repo", flagRepo,
		"update-package-json", flagUpdatePackageJson,
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
			Path:    fmt.Sprintf("%s/%s", flagOwner, flagRepo),
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

	updaters := []updater.NewUpdater{updater.Generic}

	if flagUpdatePackageJson {
		logger.DebugContext(ctx, "package.json updater enabled")
		updaters = append(updaters, updater.PackageJson)
	}

	releaserPleaser := rp.New(
		f,
		logger,
		flagBranch,
		conventionalcommits.NewParser(logger),
		versioning.SemVer,
		extraFiles,
		updaters,
	)

	return releaserPleaser.Run(ctx)
}

func parseExtraFiles(input string) []string {
	// We quote the arg to avoid issues with the expected newlines in the value.
	// Need to remove those quotes before parsing the data
	input = strings.Trim(input, `"`)
	// In some situations we get a "\n" sequence, where we actually expect new lines,
	// replace the two characters with an actual new line
	input = strings.ReplaceAll(input, `\n`, "\n")
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
