package cmd

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	rp "github.com/apricote/releaser-pleaser"
	"github.com/apricote/releaser-pleaser/internal/commitparser/conventionalcommits"
	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/forge/forgejo"
	"github.com/apricote/releaser-pleaser/internal/forge/github"
	"github.com/apricote/releaser-pleaser/internal/forge/gitlab"
	"github.com/apricote/releaser-pleaser/internal/log"
	"github.com/apricote/releaser-pleaser/internal/updater"
	"github.com/apricote/releaser-pleaser/internal/versioning"
)

func newRunCommand() *cobra.Command {
	var (
		flagForge      string
		flagBranch     string
		flagOwner      string
		flagRepo       string
		flagExtraFiles string
		flagUpdaters   []string

		flagVersioning   string
		flagCalVerFormat string
		flagNow          string

		flagAPIURL   string
		flagAPIToken string
		flagUsername string
	)

	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			logger := log.GetLogger(cmd.ErrOrStderr())

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
					Path:    fmt.Sprintf("%s/%s", flagOwner, flagRepo),
				})
				if err != nil {
					slog.ErrorContext(ctx, "failed to create client", "err", err)
					return fmt.Errorf("failed to create gitlab client: %w", err)
				}
			case "github":
				logger.DebugContext(ctx, "using forge GitHub")
				f, err = github.New(logger, &github.Options{
					Options: forgeOptions,
					Owner:   flagOwner,
					Repo:    flagRepo,
				})
				if err != nil {
					logger.ErrorContext(ctx, "failed to create client", "err", err)
					return fmt.Errorf("failed to create github client: %w", err)
				}
			case "forgejo":
				logger.DebugContext(ctx, "using forge Forgejo")
				f, err = forgejo.New(logger, &forgejo.Options{
					Options: forgeOptions,
					Owner:   flagOwner,
					Repo:    flagRepo,

					APIURL:   flagAPIURL,
					APIToken: flagAPIToken,
					Username: flagUsername,
				})
				if err != nil {
					logger.ErrorContext(ctx, "failed to create client", "err", err)
					return fmt.Errorf("failed to create forgejo client: %w", err)
				}
			default:
				return fmt.Errorf("unknown --forge: %s", flagForge)
			}

			extraFiles := parseExtraFiles(flagExtraFiles)

			updaterNames := parseUpdaters(flagUpdaters)
			updaters := []updater.Updater{}
			for _, name := range updaterNames {
				switch name {
				case "generic":
					updaters = append(updaters, updater.Generic(extraFiles))
				case "changelog":
					updaters = append(updaters, updater.Changelog())
				case "packagejson":
					updaters = append(updaters, updater.PackageJson())
				default:
					return fmt.Errorf("unknown updater: %s", name)
				}
			}

			signKey, err := signingKeyFromEnv()
			if err != nil {
				return err
			}
			if signKey != nil {
				logger.InfoContext(ctx, "signing release commits", "key.id", signKey.PrimaryKey.KeyIdString())
			}

			versioningStrategy, err := parseVersioning(flagVersioning, flagCalVerFormat, flagNow)
			if err != nil {
				return err
			}

			releaserPleaser := rp.New(
				f,
				logger,
				flagBranch,
				conventionalcommits.NewParser(logger),
				versioningStrategy,
				extraFiles,
				updaters,
				signKey,
			)

			return releaserPleaser.Run(ctx)
		},
	}

	cmd.PersistentFlags().StringVar(&flagForge, "forge", "", "")
	cmd.PersistentFlags().StringVar(&flagBranch, "branch", "main", "")
	cmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "")
	cmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "")
	cmd.PersistentFlags().StringVar(&flagExtraFiles, "extra-files", "", "")
	cmd.PersistentFlags().StringSliceVar(&flagUpdaters, "updaters", []string{}, "")

	cmd.PersistentFlags().StringVar(&flagVersioning, "versioning", VersioningSemVer, "Versioning scheme: semver, calver")
	cmd.PersistentFlags().StringVar(&flagCalVerFormat, "calver-format", versioning.DefaultCalVerFormat, "Format of CalVer versions, only used with --versioning=calver")
	// Override the current date for CalVer. Only meant for tests.
	cmd.PersistentFlags().StringVar(&flagNow, "now", "", "")
	_ = cmd.PersistentFlags().MarkHidden("now")

	cmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "")
	cmd.PersistentFlags().StringVar(&flagAPIToken, "api-token", "", "")
	cmd.PersistentFlags().StringVar(&flagUsername, "username", "", "")

	return cmd
}

const (
	VersioningSemVer = "semver"
	VersioningCalVer = "calver"
)

func parseVersioning(versioningName, calVerFormat, now string) (versioning.Strategy, error) {
	// Empty values happen when the GitHub Action/GitLab Component inputs are not set.
	if versioningName == "" {
		versioningName = VersioningSemVer
	}
	if calVerFormat == "" {
		calVerFormat = versioning.DefaultCalVerFormat
	}

	switch versioningName {
	case VersioningSemVer:
		return versioning.SemVer, nil
	case VersioningCalVer:
		clock := time.Now
		if now != "" {
			parsedNow, err := time.Parse(time.RFC3339, now)
			if err != nil {
				return nil, fmt.Errorf("invalid --now: %w", err)
			}
			clock = func() time.Time { return parsedNow }
		}

		strategy, err := versioning.NewCalVer(calVerFormat, clock)
		if err != nil {
			return nil, fmt.Errorf("invalid --calver-format: %w", err)
		}
		return strategy, nil
	default:
		return nil, fmt.Errorf("unknown --versioning: %s", versioningName)
	}
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

func parseUpdaters(input []string) []string {
	names := []string{"changelog", "generic"}

	for _, u := range input {
		if u == "" {
			continue
		}

		if strings.HasPrefix(u, "-") {
			name := u[1:]
			names = slices.DeleteFunc(names, func(existingName string) bool { return existingName == name })
		} else {
			names = append(names, u)
		}
	}

	// Make sure we only have unique updaters
	slices.Sort(names)
	names = slices.Compact(names)

	return names
}
