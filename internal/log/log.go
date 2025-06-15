package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	))
}
