package releasepr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
)

func TestTitleFormat_Title(t *testing.T) {
	assert.Equal(t, "chore(main): release v1.2.3", TitleFormatConventionalCommits.Title("main", "v1.2.3"))
	assert.Equal(t, "release: 2026.47.0", TitleFormatScoped.Title("main", "2026.47.0"))
}

func TestTitleFormat_ParseTitle(t *testing.T) {
	tests := []struct {
		format  TitleFormat
		title   string
		version string
		ok      bool
	}{
		{TitleFormatConventionalCommits, "chore(main): release v1.2.3", "v1.2.3", true},
		{TitleFormatConventionalCommits, "chore(release-1.x): release v1.2.3-rc.0", "v1.2.3-rc.0", true},
		{TitleFormatConventionalCommits, "chore(main): release the kraken today", "", false},
		{TitleFormatConventionalCommits, "api: chore(main): release v1.2.3", "", false},
		{TitleFormatConventionalCommits, "chore: release v1.2.3", "", false},
		{TitleFormatConventionalCommits, "feat: add release command", "", false},
		{TitleFormatConventionalCommits, "release: 2026.47.0", "", false},
		{TitleFormatScoped, "release: 2026.47.0", "2026.47.0", true},
		{TitleFormatScoped, "release: 2026.47.0-rc.1\n", "2026.47.0-rc.1", true},
		{TitleFormatScoped, "release: notes for QA", "", false},
		{TitleFormatScoped, "api: release 2026.47.0", "", false},
		{TitleFormatScoped, "chore(main): release 2026.47.0", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			version, ok := tt.format.ParseTitle(tt.title)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.version, version)
		})
	}
}

func TestReleasePullRequest_Version(t *testing.T) {
	pr := &ReleasePullRequest{Title: "chore(main): release v1.2.3"}
	version, err := pr.Version(TitleFormatConventionalCommits)
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", version)

	pr = &ReleasePullRequest{Title: "release: 2026.47.0"}
	version, err = pr.Version(TitleFormatScoped)
	require.NoError(t, err)
	assert.Equal(t, "2026.47.0", version)

	_, err = pr.Version(TitleFormatConventionalCommits)
	assert.ErrorContains(t, err, `title has unexpected format: "release: 2026.47.0"`)
}

func TestTitleFormat_IsReleaseCommit(t *testing.T) {
	// Accepts versions that look like SemVer or CalVer, like the real strategies would
	isValidVersion := func(version string) bool { return version != "invalid" }

	tests := []struct {
		format  TitleFormat
		message string
		want    bool
	}{
		{TitleFormatConventionalCommits, "chore(main): release v1.2.3", true},
		{TitleFormatConventionalCommits, "chore(main): release 2026.47.0-rc.0", true},
		{TitleFormatConventionalCommits, "chore(main): release 2026.47.0 (#12)", true},
		{TitleFormatConventionalCommits, "chore(main): release v1.2.3\n\nbody text", true},
		{TitleFormatConventionalCommits, "chore(release-1.x): release v1.2.3\n", true},
		{TitleFormatConventionalCommits, "chore(main): release invalid", false},
		{TitleFormatConventionalCommits, "chore(main): release the kraken today", false},
		{TitleFormatConventionalCommits, "api: chore(main): release v1.2.3", false},
		{TitleFormatConventionalCommits, "chore: release v1.2.3", false},
		{TitleFormatConventionalCommits, "feat: add release command", false},
		{TitleFormatConventionalCommits, "release: v1.2.3", false},
		{TitleFormatScoped, "release: 2026.47.0", true},
		{TitleFormatScoped, "release: 2026.47.0-rc.1 (#3)", true},
		{TitleFormatScoped, "release: invalid", false},
		{TitleFormatScoped, "release: notes for QA", false},
		{TitleFormatScoped, "api: release 2026.47.0", false},
		{TitleFormatScoped, "api: add endpoint", false},
		{TitleFormatScoped, "chore(main): release 2026.47.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.format.IsReleaseCommit(git.Commit{Message: tt.message}, isValidVersion))
		})
	}
}
