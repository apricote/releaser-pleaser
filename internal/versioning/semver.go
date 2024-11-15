package versioning

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

var SemVer Strategy = semVer{}

type semVer struct{}

func (s semVer) NextVersion(r git.Releases, versionBump VersionBump, nextVersionType NextVersionType) (string, error) {
	latest, err := parseSemverWithDefault(r.Latest)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest version: %w", err)
	}

	stable, err := parseSemverWithDefault(r.Stable)
	if err != nil {
		return "", fmt.Errorf("failed to parse stable version: %w", err)
	}

	// If there is a previous stable release, we use that as the version anchor. Falling back to any pre-releases
	// if they are the only tags in the repo.
	next := latest
	if r.Stable != nil {
		next = stable
	}

	switch versionBump {
	case UnknownVersion:
		return "", fmt.Errorf("invalid latest bump (unknown)")
	case PatchVersion:
		err = next.IncrementPatch()
	case MinorVersion:
		err = next.IncrementMinor()
	case MajorVersion:
		err = next.IncrementMajor()
	}
	if err != nil {
		return "", err
	}

	switch nextVersionType {
	case NextVersionTypeUndefined, NextVersionTypeNormal:
		next.Pre = make([]semver.PRVersion, 0)
	case NextVersionTypeAlpha, NextVersionTypeBeta, NextVersionTypeRC:
		id := uint64(0)

		if len(latest.Pre) >= 2 && latest.Pre[0].String() == nextVersionType.String() {
			if latest.Pre[1].String() == "" || !latest.Pre[1].IsNumeric() {
				return "", fmt.Errorf("invalid format of previous tag")
			}
			id = latest.Pre[1].VersionNum + 1
		}

		setPRVersion(&next, nextVersionType.String(), id)
	}

	return "v" + next.String(), nil
}

func BumpFromCommits(commits []commitparser.AnalyzedCommit) VersionBump {
	bump := UnknownVersion

	for _, commit := range commits {
		entryBump := UnknownVersion
		switch {
		case commit.BreakingChange:
			entryBump = MajorVersion
		case commit.Type == "feat":
			entryBump = MinorVersion
		case commit.Type == "fix":
			entryBump = PatchVersion
		}

		if entryBump > bump {
			bump = entryBump
		}
	}

	return bump
}

func setPRVersion(version *semver.Version, prType string, count uint64) {
	version.Pre = []semver.PRVersion{
		{VersionStr: prType},
		{VersionNum: count, IsNum: true},
	}
}

func parseSemverWithDefault(tag *git.Tag) (semver.Version, error) {
	version := "v0.0.0"
	if tag != nil {
		version = tag.Name
	}

	// The lib can not handle v prefixes
	version = strings.TrimPrefix(version, "v")

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse version %q: %w", version, err)
	}

	return parsedVersion, nil
}

func (s semVer) IsPrerelease(version string) bool {
	semVersion, err := parseSemverWithDefault(&git.Tag{Hash: "", Name: version})
	if err != nil {
		return false
	}

	if len(semVersion.Pre) > 0 {
		return true
	}

	return false
}
