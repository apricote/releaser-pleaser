//go:build e2e_forgejo

package forgejo

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/test/e2e"
)

var (
	f *e2e.Framework
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	f, err = e2e.NewFramework(ctx, &TestForge{})
	if err != nil {
		slog.Error("failed to set up test framework", "err", err)
	}

	os.Exit(m.Run())
}

func TestCreateRepository(t *testing.T) {
	_ = f.NewRepository(t, t.Name())
}

func TestRun(t *testing.T) {
	repo := f.NewRepository(t, t.Name())
	require.NoError(t, f.Run(t, repo, []string{}))
}
