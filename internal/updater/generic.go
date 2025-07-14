package updater

import (
	"regexp"
	"strings"
)

var GenericUpdaterSemVerRegex = regexp.MustCompile(`\d+\.\d+\.\d+(-[\w.]+)?(.*x-releaser-pleaser-version)`)

func Generic(info ReleaseInfo) Updater {
	return func(content string, filename string) (string, error) {
		// We strip the "v" prefix to avoid adding/removing it from the users input.
		version := strings.TrimPrefix(info.Version, "v")

		return GenericUpdaterSemVerRegex.ReplaceAllString(content, version+"${2}"), nil
	}
}
