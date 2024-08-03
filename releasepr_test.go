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
		pr             *ReleasePullRequest
		changelogEntry string
		want           string
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name:           "empty description",
			pr:             &ReleasePullRequest{},
			changelogEntry: `## v1.0.0`,
			want: `## v1.0.0

---

## releaser-pleaser Instructions

<!-- section-start overrides -->
> If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

### Prefix

` + "```" + `rp-prefix
` + "```" + `

### Suffix

` + "```" + `rp-suffix
` + "```" + `

<!-- section-end overrides -->


#### PR by [releaser-pleaser](https://github.com/apricote/releaser-pleaser)
`,
			wantErr: assert.NoError,
		},
		{
			name: "existing overrides",
			pr: &ReleasePullRequest{
				Description: `## v0.1.0

### Features

- bedazzle

---

## releaser-pleaser Instructions

<!-- section-start overrides -->
> If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

### Prefix

` + "```" + `rp-prefix
This release is awesome!
` + "```" + `

### Suffix

` + "```" + `rp-suffix
` + "```" + `

<!-- section-end overrides -->

#### PR by [releaser-pleaser](https://github.com/apricote/releaser-pleaser)
`,
			},
			changelogEntry: `## v1.0.0`,
			want: `## v1.0.0

---

## releaser-pleaser Instructions

<!-- section-start overrides -->
> If you want to modify the proposed release, add you overrides here. You can learn more about the options in the docs.

### Prefix

` + "```" + `rp-prefix
This release is awesome!
` + "```" + `

### Suffix

` + "```" + `rp-suffix
` + "```" + `

<!-- section-end overrides -->

#### PR by [releaser-pleaser](https://github.com/apricote/releaser-pleaser)
`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pr.SetDescription(tt.changelogEntry)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, tt.pr.Description)
		})
	}
}
