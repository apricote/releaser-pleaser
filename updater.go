package rp

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	GenericUpdaterSemVerRegex   = regexp.MustCompile(`\d+\.\d+\.\d+(-[\w.]+)?(.*x-releaser-pleaser-version)`)
	ChangelogUpdaterHeaderRegex = regexp.MustCompile(`^# Changelog\n`)
)

type ReleaseInfo struct {
	Version        string
	ChangelogEntry string
}

type Updater interface {
	UpdateContent(content string, info ReleaseInfo) (string, error)
}

type GenericUpdater struct{}

func (u *GenericUpdater) UpdateContent(content string, info ReleaseInfo) (string, error) {
	// We strip the "v" prefix to avoid adding/removing it from the users input.
	version := strings.TrimPrefix(info.Version, "v")

	return GenericUpdaterSemVerRegex.ReplaceAllString(content, version+"${2}"), nil
}

type ChangelogUpdater struct{}

func (u *ChangelogUpdater) UpdateContent(content string, info ReleaseInfo) (string, error) {
	headerIndex := ChangelogUpdaterHeaderRegex.FindStringIndex(content)
	if headerIndex == nil && len(content) != 0 {
		return "", fmt.Errorf("unexpected format of CHANGELOG.md, header does not match")
	}
	if headerIndex != nil {
		// Remove the header from the content
		content = content[headerIndex[1]:]
	}

	content = ChangelogHeader + "\n\n" + info.ChangelogEntry + content

	return content, nil
}
