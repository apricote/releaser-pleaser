package updater

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageJsonUpdater(t *testing.T) {
	tests := []updaterTestCase{
		{
			name:     "simple package.json",
			content:  `{"name":"test","version":"1.0.0"}`,
			filename: "package.json",
			info: ReleaseInfo{
				Version: "v2.0.5",
			},
			want:    `{"name":"test","version":"2.0.5"}`,
			wantErr: assert.NoError,
		},
		{
			name:     "simple package.json, wrong name",
			content:  `{"name":"test","version":"1.0.0"}`,
			filename: "nopackage.json",
			info: ReleaseInfo{
				Version: "v2.0.5",
			},
			want:    `{"name":"test","version":"1.0.0"}`,
			wantErr: assert.NoError,
		},
		{
			name:     "complex package.json",
			content:  "{\n  \"name\": \"test\",\n  \"version\": \"1.0.0\",\n  \"dependencies\": {\n    \"foo\": \"^1.0.0\"\n  }\n}",
			filename: "package.json",
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    "{\n  \"name\": \"test\",\n  \"version\": \"2.0.0\",\n  \"dependencies\": {\n    \"foo\": \"^1.0.0\"\n  }\n}",
			wantErr: assert.NoError,
		},
		{
			name:     "invalid json",
			content:  `not json`,
			filename: "package.json",
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `not json`,
			wantErr: assert.NoError,
		},
		{
			name:     "json without version",
			content:  `{"name":"test"}`,
			filename: "package.json",
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `{"name":"test"}`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("Running updater test for PackageJson")
			runUpdaterTest(t, PackageJson, tt)
		})
	}
}
