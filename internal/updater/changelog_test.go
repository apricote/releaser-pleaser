package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangelogUpdater_UpdateContent(t *testing.T) {
	tests := []updaterTestCase{
		{
			name:    "empty file",
			content: "",
			info:    ReleaseInfo{ChangelogEntry: "## v1.0.0\n"},
			want:    "# Changelog\n\n## v1.0.0\n",
			wantErr: assert.NoError,
		},
		{
			name: "well-formatted changelog",
			content: `# Changelog

## v0.0.1

- Bazzle

## v0.1.0

### Bazuuum
`,
			info: ReleaseInfo{ChangelogEntry: "## v1.0.0\n\n- Version 1, juhu.\n"},
			want: `# Changelog

## v1.0.0

- Version 1, juhu.

## v0.0.1

- Bazzle

## v0.1.0

### Bazuuum
`,
			wantErr: assert.NoError,
		},
		{
			name:    "error on invalid header",
			content: "What even is this file?",
			info:    ReleaseInfo{ChangelogEntry: "## v1.0.0\n\n- Version 1, juhu.\n"},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUpdaterTest(t, Changelog, tt)
		})
	}
}
