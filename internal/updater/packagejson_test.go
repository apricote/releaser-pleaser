package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageJsonUpdater_Files(t *testing.T) {
	assert.Equal(t, []string{"package.json"}, PackageJson().Files())
}

func TestPackageJsonUpdater_CreateNewFiles(t *testing.T) {
	assert.False(t, PackageJson().CreateNewFiles())
}

func TestPackageJsonUpdater_Update(t *testing.T) {
	tests := []updaterTestCase{
		{
			name:    "simple package.json",
			content: `{"name":"test","version":"1.0.0"}`,
			info: ReleaseInfo{
				Version: "v2.0.5",
			},
			want:    `{"name":"test","version":"2.0.5"}`,
			wantErr: assert.NoError,
		},
		{
			name:    "complex package.json",
			content: "{\n  \"name\": \"test\",\n  \"version\": \"1.0.0\",\n  \"dependencies\": {\n    \"foo\": \"^1.0.0\"\n  }\n}",
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    "{\n  \"name\": \"test\",\n  \"version\": \"2.0.0\",\n  \"dependencies\": {\n    \"foo\": \"^1.0.0\"\n  }\n}",
			wantErr: assert.NoError,
		},
		{
			name:    "invalid json",
			content: `not json`,
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `not json`,
			wantErr: assert.NoError,
		},
		{
			name:    "json without version",
			content: `{"name":"test"}`,
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `{"name":"test"}`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUpdaterTest(t, PackageJson(), tt)
		})
	}
}
