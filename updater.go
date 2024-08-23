package rp

import (
	"regexp"
	"strings"
)

var (
	GenericUpdaterSemVerRegex   = regexp.MustCompile(`\d+\.\d+\.\d+(-[\w.]+)?(.*x-releaser-pleaser-version)`)
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
