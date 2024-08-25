package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:     "releaser-pleaser",
	Short:   "",
	Long:    ``,
	Version: version(),
}

func version() string {
	vcsrevision := "unknown"
	vcsdirty := ""

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		fmt.Println(buildInfo.String())
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
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

}
