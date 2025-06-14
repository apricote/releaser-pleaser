package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:           "rp",
	Short:         "",
	Long:          ``,
	Version:       version(),
	SilenceUsage:  true, // Makes it harder to find the actual error
	SilenceErrors: true, // We log manually with slog
}

func version() string {
	vcsrevision := "unknown"
	vcsdirty := ""

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				vcsrevision = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					vcsdirty = " (dirty)"
				}
			}
		}
	}

	return vcsrevision + vcsdirty
}

func Execute() {
	// Behaviour when cancelling jobs:
	//
	//   GitHub Actions: https://docs.github.com/en/actions/managing-workflow-runs-and-deployments/managing-workflow-runs/canceling-a-workflow#steps-github-takes-to-cancel-a-workflow-run
	// 	   1. SIGINT
	//     2. Wait 7500ms
	//     3. SIGTERM
	//     4. Wait 2500ms
	//     5. SIGKILL
	//
	//   GitLab CI/CD: https://gitlab.com/gitlab-org/gitlab-runner/-/merge_requests/4446
	//     1. SIGTERM
	//     2. Wait ???
	//     3. SIGKILL
	//
	// We therefore need to listen on SIGINT and SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		// Make sure to stop listening on signals after receiving the first signal to hand control of the signal back
		// to the runtime. The Go runtime implements a "force shutdown" if the signal is received again.
		<-ctx.Done()
		logger.InfoContext(ctx, "Received shutdown signal, stopping...")
		stop()
	}()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

func init() {
	logger = slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	slog.SetDefault(logger)
}
