package releasepr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/apricote/releaser-pleaser/internal/git"
)

// TitleFormat defines the title of release pull requests and the message of release commits. It should match the
// commit message format used in the repository.
type TitleFormat struct {
	format func(branch, version string) string
	// regex matches a title of this format, the first group is the version.
	regex *regexp.Regexp
}

var (
	// TitleFormatConventionalCommits creates titles like "chore(main): release v1.2.3".
	TitleFormatConventionalCommits = TitleFormat{
		format: func(branch, version string) string { return fmt.Sprintf("chore(%s): release %s", branch, version) },
		regex:  regexp.MustCompile(`^chore\([^)]*\): release (\S+)$`),
	}

	// TitleFormatScoped creates titles like "release: 2026.47.0".
	TitleFormatScoped = TitleFormat{
		format: func(_, version string) string { return "release: " + version },
		regex:  regexp.MustCompile(`^release: (\S+)$`),
	}

	// squashSuffixRegex matches the suffix that forges add to the commit message of squash merges.
	squashSuffixRegex = regexp.MustCompile(` \(#\d+\)$`)
)

// Title returns the title for a release of version on branch.
func (f TitleFormat) Title(branch, version string) string {
	return f.format(branch, version)
}

// ParseTitle returns the version from a release title of this format.
func (f TitleFormat) ParseTitle(title string) (version string, ok bool) {
	matches := f.regex.FindStringSubmatch(strings.TrimSpace(title))
	if matches == nil {
		return "", false
	}

	return matches[1], true
}

// IsReleaseCommit reports whether the commit is a release commit created by releaser-pleaser, including squash merges
// of the release pull request. isValidVersion is used to check the version in the title, to avoid matching regular
// commits like "release: notes for QA". These commits should never show up in the changelog.
func (f TitleFormat) IsReleaseCommit(commit git.Commit, isValidVersion func(version string) bool) bool {
	version, ok := f.ParseTitle(squashSuffixRegex.ReplaceAllString(commit.Subject(), ""))
	return ok && isValidVersion(version)
}
