package e2e

import (
	"context"
	"testing"
)

type TestForge interface {
	Init(ctx context.Context, runID string) error
	CreateRepo(t *testing.T, opts CreateRepoOpts) (*Repository, error)

	RunArguments() []string
}

type CreateRepoOpts struct {
	Name          string
	Description   string
	DefaultBranch string
}
