package rp

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5"
)

const (
	ChangelogFile       = "CHANGELOG.md"
	ChangelogFileBuffer = "CHANGELOG.md.tmp"
	ChangelogHeader     = "# Changelog"
)

var (
	changelogTemplate *template.Template

	headerRegex = regexp.MustCompile(`^# Changelog\n`)
)

//go:embed changelog.md.tpl
var rawChangelogTemplate string

func init() {
	var err error
	changelogTemplate, err = template.New("changelog").Parse(rawChangelogTemplate)
	if err != nil {
		log.Fatalf("failed to parse changelog template: %v", err)
	}
}

func UpdateChangelogFile(wt *git.Worktree, newEntry string) error {
	file, err := wt.Filesystem.OpenFile(ChangelogFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	headerIndex := headerRegex.FindIndex(content)
	if headerIndex == nil && len(content) != 0 {
		return fmt.Errorf("unexpected format of CHANGELOG.md, header does not match")
	}
	if headerIndex != nil {
		// Remove the header from the content
		content = content[headerIndex[1]:]
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(ChangelogHeader + "\n\n" + newEntry))
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	// Close file to make sure it is written to disk.
	err = file.Close()
	if err != nil {
		return err
	}

	_, err = wt.Add(ChangelogFile)
	if err != nil {
		return err
	}

	return nil
}

func NewChangelogEntry(changesets []Changeset, version, link, prefix, suffix string) (string, error) {
	features := make([]AnalyzedCommit, 0)
	fixes := make([]AnalyzedCommit, 0)

	for _, changeset := range changesets {
		for _, commit := range changeset.ChangelogEntries {
			switch commit.Type {
			case "feat":
				features = append(features, commit)
			case "fix":
				fixes = append(fixes, commit)
			}
		}
	}

	var changelog bytes.Buffer
	err := changelogTemplate.Execute(&changelog, map[string]any{
		"Features":    features,
		"Fixes":       fixes,
		"Version":     version,
		"VersionLink": link,
		"Prefix":      prefix,
		"Suffix":      suffix,
	})
	if err != nil {
		return "", err
	}

	return changelog.String(), nil

}
