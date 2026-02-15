package updater

import (
	"regexp"
)

// HelmChart creates an updater that modifies the version field in Chart.yaml files
func HelmChart() Updater {
	return helmchart{}
}

type helmchart struct{}

func (h helmchart) Files() []string {
	return []string{"Chart.yaml"}
}

func (h helmchart) CreateNewFiles() bool {
	return false
}

func (h helmchart) Update(info ReleaseInfo) func(content string) (string, error) {
	return func(content string) (string, error) {
		// Regex to match ^version: ...$ with flexible whitespace in multiline mode
		versionRegex := regexp.MustCompile(`(?m:^(version:\s*)\S*$)`)

		// Check if the file contains a version field
		if !versionRegex.MatchString(content) {
			return content, nil
		}

		// Replace the version value while preserving the original formatting
		return versionRegex.ReplaceAllString(content, `${1}`+info.Version), nil
	}
}
