package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericUpdater_UpdateContent(t *testing.T) {
	tests := []updaterTestCase{
		{
			name:     "single line",
			content:  "v1.0.0 // x-releaser-pleaser-version",
			filename: "version.txt",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "v1.2.0 // x-releaser-pleaser-version",
			wantErr: assert.NoError,
		},
		{
			name:     "multiline line",
			content:  "Foooo\n\v1.2.0\nv1.0.0 // x-releaser-pleaser-version\n",
			filename: "version.txt",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "Foooo\n\v1.2.0\nv1.2.0 // x-releaser-pleaser-version\n",
			wantErr: assert.NoError,
		},
		{
			name:     "invalid existing version",
			content:  "1.0 // x-releaser-pleaser-version",
			filename: "version.txt",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "1.0 // x-releaser-pleaser-version",
			wantErr: assert.NoError,
		},
		{
			name:     "complicated line",
			content:  "version: v1.2.0-alpha.1 => Awesome, isnt it? x-releaser-pleaser-version foobar",
			filename: "version.txt",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "version: v1.2.0 => Awesome, isnt it? x-releaser-pleaser-version foobar",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUpdaterTest(t, Generic, tt)
		})
	}
}
