package rp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReleasePullRequest_SetTitle(t *testing.T) {
	type args struct {
		branch  string
		version string
	}
	tests := []struct {
		name string
		pr   *ReleasePullRequest
		args args
		want string
	}{
		{
			name: "simple update",
			pr:   &ReleasePullRequest{Title: "foo: bar"},
			args: args{
				branch:  "main",
				version: "v1.0.0",
			},
			want: "chore(main): release v1.0.0",
		},
		{
			name: "no previous title",
			pr:   &ReleasePullRequest{},
			args: args{
				branch:  "release-1.x",
				version: "v1.1.1-rc.0",
			},
			want: "chore(release-1.x): release v1.1.1-rc.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pr.SetTitle(tt.args.branch, tt.args.version)

			assert.Equal(t, tt.want, tt.pr.Title)
		})
	}
}

func TestReleasePullRequest_SetDescription(t *testing.T) {

	tests := []struct {
		name           string
		changelogEntry string
		overrides      ReleaseOverrides
		want           string
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name:           "no overrides",
			changelogEntry: `## v1.0.0`,
			overrides:      ReleaseOverrides{},
			want: `<!-- section-start changelog -->
## v1.0.0
<!-- section-end changelog -->

---

<details>
  <summary><h4>PR by <a href="https://github.com/apricote/releaser-pleaser">releaser-pleaser</a> 🤖</h4></summary>

If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

## Release Notes

### Prefix / Start

This will be added to the start of the release notes.

` + "```" + `rp-prefix
` + "```" + `

### Suffix / End

This will be added to the end of the release notes.

` + "```" + `rp-suffix
` + "```" + `

</details>
`,
			wantErr: assert.NoError,
		},
		{
			name:           "existing overrides",
			changelogEntry: `## v1.0.0`,
			overrides: ReleaseOverrides{
				Prefix: "This release is awesome!",
				Suffix: "Fooo",
			},
			want: `<!-- section-start changelog -->
## v1.0.0
<!-- section-end changelog -->

---

<details>
  <summary><h4>PR by <a href="https://github.com/apricote/releaser-pleaser">releaser-pleaser</a> 🤖</h4></summary>

If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

## Release Notes

### Prefix / Start

This will be added to the start of the release notes.

` + "```" + `rp-prefix
This release is awesome!
` + "```" + `

### Suffix / End

This will be added to the end of the release notes.

` + "```" + `rp-suffix
Fooo
` + "```" + `

</details>
`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &ReleasePullRequest{}
			err := pr.SetDescription(tt.changelogEntry, tt.overrides)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, pr.Description)
		})
	}
}
