package updater

import (
	"regexp"
	"strings"
)

// PackageJson creates an updater that modifies the version field in package.json files
func PackageJson(info ReleaseInfo) Updater {
	return func(content string, filename string) (string, error) {
		if filename != "package.json" {
			return content, nil // No update needed for non-package.json files
		}
		// We strip the "v" prefix to match npm versioning convention
		version := strings.TrimPrefix(info.Version, "v")

		// Regex to match "version": "..." with flexible whitespace and quote styles
		versionRegex := regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)

		// Check if the file contains a version field
		if !versionRegex.MatchString(content) {
			return content, nil
		}

		// Replace the version value while preserving the original formatting
		updatedContent := versionRegex.ReplaceAllString(content, `${1}"`+version+`"`)

		return updatedContent, nil
	}
}
