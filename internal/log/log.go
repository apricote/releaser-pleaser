package log

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func GetLogger(w io.Writer) *slog.Logger {
	return slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)
}

func init() {
	slog.SetDefault(GetLogger(os.Stderr))
}
