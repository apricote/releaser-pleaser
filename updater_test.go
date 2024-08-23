package rp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type updaterTestCase struct {
	name    string
	content string
	info    ReleaseInfo
	want    string
	wantErr assert.ErrorAssertionFunc
}

func runUpdaterTest(t *testing.T, updater Updater, tt updaterTestCase) {
	t.Helper()

	got, err := updater.UpdateContent(tt.content, tt.info)
	if !tt.wantErr(t, err, fmt.Sprintf("UpdateContent(%v, %v)", tt.content, tt.info)) {
		return
	}
	assert.Equalf(t, tt.want, got, "UpdateContent(%v, %v)", tt.content, tt.info)
}

func TestGenericUpdater_UpdateContent(t *testing.T) {
	updater := &GenericUpdater{}

	tests := []updaterTestCase{
		{
			name:    "single line",
			content: "v1.0.0 // x-releaser-pleaser-version",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "v1.2.0 // x-releaser-pleaser-version",
			wantErr: assert.NoError,
		},
		{
			name:    "multiline line",
			content: "Foooo\n\v1.2.0\nv1.0.0 // x-releaser-pleaser-version\n",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "Foooo\n\v1.2.0\nv1.2.0 // x-releaser-pleaser-version\n",
			wantErr: assert.NoError,
		},
		{
			name:    "invalid existing version",
			content: "1.0 // x-releaser-pleaser-version",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "1.0 // x-releaser-pleaser-version",
			wantErr: assert.NoError,
		},
		{
			name:    "complicated line",
			content: "version: v1.2.0-alpha.1 => Awesome, isnt it? x-releaser-pleaser-version foobar",
			info: ReleaseInfo{
				Version: "v1.2.0",
			},
			want:    "version: v1.2.0 => Awesome, isnt it? x-releaser-pleaser-version foobar",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUpdaterTest(t, updater, tt)
		})
	}
}
