package rp

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/leodido/go-conventionalcommits"
)

func NextVersion(currentTag *Tag, changesets []Changeset, nextVersionType NextVersionType) (string, error) {
	// TODO: Validate for versioning after pre-releases
	currentVersion := "v0.0.0"
	if currentTag != nil {
		currentVersion = currentTag.Name
	}

	// The lib can not handle v prefixes
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	version, err := semver.Parse(currentVersion)
	if err != nil {
		return "", err
	}

	versionBump := maxVersionBump(changesets)
	switch versionBump {
	case conventionalcommits.UnknownVersion:
		// No new version, TODO: Throw error?
	case conventionalcommits.PatchVersion:
		err = version.IncrementPatch()
	case conventionalcommits.MinorVersion:
		err = version.IncrementMinor()
	case conventionalcommits.MajorVersion:
		err = version.IncrementMajor()
	}
	if err != nil {
		return "", err
	}

	switch nextVersionType {
	case NextVersionTypeAlpha, NextVersionTypeBeta, NextVersionTypeRC:
		id := uint64(0)

		if version.Pre[0].String() == nextVersionType.String() {
			if version.Pre[1].String() == "" || !version.Pre[1].IsNumeric() {
				return "", fmt.Errorf("invalid format of previous tag")
			}
			id = version.Pre[1].VersionNum + 1
		}

		setPRVersion(&version, nextVersionType.String(), id)
	case NextVersionTypeUndefined, NextVersionTypeNormal:
		version.Pre = make([]semver.PRVersion, 0)
	}

	return "v" + version.String(), nil
}

func maxVersionBump(changesets []Changeset) conventionalcommits.VersionBump {
	bump := conventionalcommits.UnknownVersion

	for _, changeset := range changesets {
		for _, entry := range changeset.ChangelogEntries {
			entryBump := conventionalcommits.UnknownVersion
			switch {
			case entry.BreakingChange:
				entryBump = conventionalcommits.MajorVersion
			case entry.Type == "feat":
				entryBump = conventionalcommits.MinorVersion
			case entry.Type == "fix":
				entryBump = conventionalcommits.PatchVersion
			}

			if entryBump > bump {
				bump = entryBump
			}
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
