package rp

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/leodido/go-conventionalcommits"
)

func NextVersion(latestTag *Tag, stableTag *Tag, versionBump conventionalcommits.VersionBump, nextVersionType NextVersionType) (string, error) {
	// TODO: Validate for versioning after pre-releases
	latestVersion := "v0.0.0"
	if latestTag != nil {
		latestVersion = latestTag.Name
	}
	stableVersion := "v0.0.0"
	if stableTag != nil {
		stableVersion = stableTag.Name
	}

	// The lib can not handle v prefixes
	latestVersion = strings.TrimPrefix(latestVersion, "v")
	stableVersion = strings.TrimPrefix(stableVersion, "v")

	latest, err := semver.Parse(latestVersion)
	if err != nil {
		return "", err
	}

	stable, err := semver.Parse(stableVersion)
	if err != nil {
		return "", err
	}

	next := stable // Copy all fields

	switch versionBump {
	case conventionalcommits.UnknownVersion:
		return "", fmt.Errorf("invalid latest bump (unknown)")
	case conventionalcommits.PatchVersion:
		err = next.IncrementPatch()
	case conventionalcommits.MinorVersion:
		err = next.IncrementMinor()
	case conventionalcommits.MajorVersion:
		err = next.IncrementMajor()
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

	if err != nil {
		return "", err
	}

	return "v" + next.String(), nil
}

func VersionBumpFromChangesets(changesets []Changeset) conventionalcommits.VersionBump {
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
