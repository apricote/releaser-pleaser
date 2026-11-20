package versioning

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
)

func TestLatestReleases(t *testing.T) {
	calver, err := NewCalVer("YYYY.0W.MICRO", nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		strategy   Strategy
		tags       []string
		wantLatest string
		wantStable string
	}{
		{
			name:     "no tags",
			strategy: SemVer,
			tags:     nil,
		},
		{
			name:       "semver stable",
			strategy:   SemVer,
			tags:       []string{"v1.1.0", "v1.0.0"},
			wantLatest: "v1.1.0",
			wantStable: "v1.1.0",
		},
		{
			name:       "semver unsorted tags",
			strategy:   SemVer,
			tags:       []string{"v1.9.0", "v1.11.0-rc.0", "v1.10.0", "v1.10.0-rc.1", "v1.1.0"},
			wantLatest: "v1.11.0-rc.0",
			wantStable: "v1.10.0",
		},
		{
			name:       "semver prerelease and invalid tags",
			strategy:   SemVer,
			tags:       []string{"foo", "v1.1.0-rc.0", "2026.07.0", "v1.0.0", "v0.9.0"},
			wantLatest: "v1.1.0-rc.0",
			wantStable: "v1.0.0",
		},
		{
			name:       "semver only prerelease",
			strategy:   SemVer,
			tags:       []string{"v1.1.0-rc.0"},
			wantLatest: "v1.1.0-rc.0",
		},
		{
			name:       "calver ignores semver tags",
			strategy:   calver,
			tags:       []string{"v1.2.3", "2026.47.1-rc.0", "2026.47.0", "2026.46.0"},
			wantLatest: "2026.47.1-rc.0",
			wantStable: "2026.47.0",
		},
		{
			name:       "calver unsorted tags",
			strategy:   calver,
			tags:       []string{"2026.09.0", "2026.10.0", "2026.10.10", "2026.10.9", "2026.10.11-rc.0"},
			wantLatest: "2026.10.11-rc.0",
			wantStable: "2026.10.10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := make([]git.Tag, 0, len(tt.tags))
			for _, name := range tt.tags {
				tags = append(tags, git.Tag{Name: name, Hash: "hash-" + name})
			}

			got := LatestReleases(t.Context(), slog.Default(), tags, tt.strategy)

			if tt.wantLatest == "" {
				assert.Nil(t, got.Latest)
			} else if assert.NotNil(t, got.Latest) {
				assert.Equal(t, tt.wantLatest, got.Latest.Name)
			}
			if tt.wantStable == "" {
				assert.Nil(t, got.Stable)
			} else if assert.NotNil(t, got.Stable) {
				assert.Equal(t, tt.wantStable, got.Stable.Name)
			}
		})
	}
}
