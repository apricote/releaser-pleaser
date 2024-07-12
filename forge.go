package rp

import (
	"fmt"
)

type Forge interface {
	RepoURL() string
}

type ForgeOptions struct {
	Repository string
}

var _ Forge = &GitHub{}
var _ Forge = &GitLab{}

type GitHub struct {
	options ForgeOptions
}

func (g *GitHub) RepoURL() string {
	return fmt.Sprintf("https://github.com/%s", g.options.Repository)
}

func (g *GitHub) autodiscover() {
	// Read settings from GitHub Actions env vars
}

func NewGitHub(options ForgeOptions) *GitHub {
	gh := &GitHub{
		options: options,
	}

	gh.autodiscover()

	return gh
}

type GitLab struct {
	options ForgeOptions
}

func (g *GitLab) autodiscover() {
	// Read settings from GitLab-CI env vars
}

func NewGitLab(options ForgeOptions) *GitLab {
	gl := &GitLab{
		options: options,
	}

	gl.autodiscover()

	return gl
}

func (g *GitLab) RepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%s", g.options.Repository)
}
