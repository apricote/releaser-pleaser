package updater

import (
	"fmt"
	"regexp"
)

const (
	ChangelogHeader = "# Changelog"
	ChangelogFile   = "CHANGELOG.md"
)

var (
	ChangelogUpdaterHeaderRegex = regexp.MustCompile(`^# Changelog\n`)
)

func Changelog() Updater {
	return changelog{}
}

type changelog struct {
}

func (c changelog) Files() []string {
	return []string{ChangelogFile}
}

func (c changelog) CreateNewFiles() bool {
	return true
}

func (c changelog) Update(info ReleaseInfo) func(content string) (string, error) {
	return func(content string) (string, error) {
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
}
