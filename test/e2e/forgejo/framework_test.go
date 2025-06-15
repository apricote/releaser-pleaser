//go:build e2e_forgejo

package forgejo

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	suffix := randomSuffix()

	TestToken = setupTestUser(ctx, suffix)

	TestClient = setupTestClient(ctx, TestToken)

	os.Exit(m.Run())
}

func TestAPIAccess(t *testing.T) {
	user, _, err := TestClient.GetMyUserInfo()
	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestCreateRepository(t *testing.T) {
	_ = NewRepository(t, t.Name())
}

func TestRun(t *testing.T) {
	repo := NewRepository(t, t.Name())
	MustRun(t, repo, []string{})

}
