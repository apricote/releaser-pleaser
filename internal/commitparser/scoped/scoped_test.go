package scoped

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name     string
		commits  []git.Commit
		expected []commitparser.AnalyzedCommit
	}{
		{
			name:     "empty commits",
			commits:  []git.Commit{},
			expected: []commitparser.AnalyzedCommit{},
		},
		{
			name:    "simple scope",
			commits: []git.Commit{{Message: "api: add endpoint"}},
			expected: []commitparser.AnalyzedCommit{
				{Message: "api: add endpoint", Type: TypeChange, Scope: new("api"), Description: "add endpoint"},
			},
		},
		{
			name: "scope with slash and dash",
			commits: []git.Commit{
				{Message: "api/v2: add endpoint"},
				{Message: "my-svc: fix crash"},
			},
			expected: []commitparser.AnalyzedCommit{
				{Message: "api/v2: add endpoint", Type: TypeChange, Scope: new("api/v2"), Description: "add endpoint"},
				{Message: "my-svc: fix crash", Type: TypeChange, Scope: new("my-svc"), Description: "fix crash"},
			},
		},
		{
			name:    "no scope",
			commits: []git.Commit{{Message: "fix typo in readme"}},
			expected: []commitparser.AnalyzedCommit{
				{Message: "fix typo in readme", Type: TypeChange, Description: "fix typo in readme"},
			},
		},
		{
			name:    "merge commit has no scope",
			commits: []git.Commit{{Message: "Merge branch 'feature' into 'main'"}},
			expected: []commitparser.AnalyzedCommit{
				{Message: "Merge branch 'feature' into 'main'", Type: TypeChange, Description: "Merge branch 'feature' into 'main'"},
			},
		},
		{
			name:    "colon without space is not a scope",
			commits: []git.Commit{{Message: "see http://example.com"}},
			expected: []commitparser.AnalyzedCommit{
				{Message: "see http://example.com", Type: TypeChange, Description: "see http://example.com"},
			},
		},
		{
			name:    "only subject is used",
			commits: []git.Commit{{Message: "\nui: new button\n\nsome body\nwith: colon\n"}},
			expected: []commitparser.AnalyzedCommit{
				{Message: "\nui: new button\n\nsome body\nwith: colon\n", Type: TypeChange, Scope: new("ui"), Description: "new button"},
			},
		},
		{
			name:     "empty message is skipped",
			commits:  []git.Commit{{Message: "  \n "}},
			expected: []commitparser.AnalyzedCommit{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParser(slog.Default()).Analyze(tt.commits)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
