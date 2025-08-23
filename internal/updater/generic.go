package updater

import (
	"regexp"
	"strings"
)

var GenericUpdaterSemVerRegex = regexp.MustCompile(`\d+\.\d+\.\d+(-[\w.]+)?(.*x-releaser-pleaser-version)`)

func Generic(files []string) Updater {
	return generic{
		files: files,
	}
}

type generic struct {
	files []string
}

func (g generic) Files() []string {
	return g.files
}

func (g generic) CreateNewFiles() bool {
	return false
}

func (g generic) Update(info ReleaseInfo) func(content string) (string, error) {
	return func(content string) (string, error) {
		// We strip the "v" prefix to avoid adding/removing it from the users input.
		version := strings.TrimPrefix(info.Version, "v")

		return GenericUpdaterSemVerRegex.ReplaceAllString(content, version+"${2}"), nil
	}
}
