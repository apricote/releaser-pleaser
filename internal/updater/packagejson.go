package updater

import (
	"regexp"
	"strings"
)

// PackageJson creates an updater that modifies the version field in package.json files
func PackageJson() Updater {
	return packagejson{}
}

type packagejson struct{}

func (p packagejson) Files() []string {
	return []string{"package.json"}
}

func (p packagejson) CreateNewFiles() bool {
	return false
}

func (p packagejson) Update(info ReleaseInfo) func(content string) (string, error) {
	return func(content string) (string, error) {
		// We strip the "v" prefix to match npm versioning convention
		version := strings.TrimPrefix(info.Version, "v")

		// Regex to match "version": "..." with flexible whitespace and quote styles
		versionRegex := regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)

		// Check if the file contains a version field
		if !versionRegex.MatchString(content) {
			return content, nil
		}

		// Replace the version value while preserving the original formatting
		return versionRegex.ReplaceAllString(content, `${1}"`+version+`"`), nil
	}
}
