package versioning

import (
	"context"
	"log/slog"
	"slices"

	"github.com/leodido/go-conventionalcommits"

	"github.com/apricote/releaser-pleaser/internal/git"
)

type Strategy interface {
	NextVersion(git.Releases, VersionBump, NextVersionType) (string, error)
	IsPrerelease(version string) bool

	// Validate returns an error if the tag name is not a version of this Strategy. Tags that are not valid are ignored
	// when looking for previous releases.
	Validate(tag string) error

	// Compare compares two valid versions of this Strategy. It returns a negative number if a < b, zero if a == b and
	// a positive number if a > b.
	Compare(a, b string) int
}

// LatestReleases selects the latest release and the latest stable release from tags. The order of tags does not
// matter, they are sorted by the Strategy. Tags that are not valid for the Strategy are skipped.
func LatestReleases(ctx context.Context, logger *slog.Logger, tags []git.Tag, s Strategy) git.Releases {
	versions := make([]git.Tag, 0, len(tags))
	for _, tag := range tags {
		if err := s.Validate(tag.Name); err != nil {
			logger.DebugContext(ctx, "tag is not a valid version, skipping", "tag.name", tag.Name, "tag.hash", tag.Hash, "error", err)
			continue
		}
		versions = append(versions, tag)
	}

	if len(versions) == 0 && len(tags) > 0 {
		logger.WarnContext(ctx, "none of the tags in the repository are valid versions, ignoring all of them", "tags", len(tags))
	}

	// Most recent version first. The sort is stable, so equal versions (e.g. "v1.0.0" and "1.0.0") keep their order.
	slices.SortStableFunc(versions, func(a, b git.Tag) int {
		return s.Compare(b.Name, a.Name)
	})

	var releases git.Releases
	for i := range versions {
		if releases.Latest == nil {
			releases.Latest = &versions[i]
		}

		if !s.IsPrerelease(versions[i].Name) {
			releases.Stable = &versions[i]
			break
		}
	}

	return releases
}

type VersionBump conventionalcommits.VersionBump

const (
	UnknownVersion VersionBump = iota
	PatchVersion
	MinorVersion
	MajorVersion
)

type NextVersionType int

const (
	NextVersionTypeUndefined NextVersionType = iota
	NextVersionTypeNormal
	NextVersionTypeRC
	NextVersionTypeBeta
	NextVersionTypeAlpha
)

func (n NextVersionType) String() string {
	switch n {
	case NextVersionTypeUndefined:
		return "undefined"
	case NextVersionTypeNormal:
		return "normal"
	case NextVersionTypeRC:
		return "rc"
	case NextVersionTypeBeta:
		return "beta"
	case NextVersionTypeAlpha:
		return "alpha"
	default:
		return ""
	}
}

func (n NextVersionType) IsPrerelease() bool {
	switch n {
	case NextVersionTypeRC, NextVersionTypeBeta, NextVersionTypeAlpha:
		return true
	case NextVersionTypeUndefined, NextVersionTypeNormal:
		return false
	default:
		return false
	}
}
