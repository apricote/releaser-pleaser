package cmd

import (
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:     "rp",
	Short:   "",
	Long:    ``,
	Version: version(),
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
	err := rootCmd.Execute()
	if err != nil {
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
