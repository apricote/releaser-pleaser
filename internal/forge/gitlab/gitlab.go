package gitlab

import (
	"fmt"

	"github.com/apricote/releaser-pleaser/internal/forge"
)

// var _ forge.Forge = &GitLab{}

type GitLab struct {
	options forge.Options
}

func (g *GitLab) autodiscover() {
	// Read settings from GitLab-CI env vars
}

func New(options forge.Options) *GitLab {
	gl := &GitLab{
		options: options,
	}

	gl.autodiscover()

	return gl
}

func (g *GitLab) RepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%s", g.options.Repository)
}
