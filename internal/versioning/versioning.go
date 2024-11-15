package versioning

import (
	"github.com/leodido/go-conventionalcommits"

	"github.com/apricote/releaser-pleaser/internal/git"
)

type Strategy interface {
	NextVersion(git.Releases, VersionBump, NextVersionType) (string, error)
	IsPrerelease(version string) bool
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
