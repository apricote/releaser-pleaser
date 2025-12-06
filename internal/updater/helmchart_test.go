package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmChartUpdater_Files(t *testing.T) {
	assert.Equal(t, []string{"Chart.yaml"}, HelmChart().Files())
}

func TestHelmChartUpdater_CreateNewFiles(t *testing.T) {
	assert.False(t, HelmChart().CreateNewFiles())
}

func TestHelmChartUpdater_Update(t *testing.T) {
	tests := []updaterTestCase{
		{
			name:    "simple Chart.yaml",
			content: "apiVersion: v2\nname: test\nversion: v1.0.0",
			info: ReleaseInfo{
				Version: "v2.0.5",
			},
			want:    "apiVersion: v2\nname: test\nversion: v2.0.5",
			wantErr: assert.NoError,
		},
		{
			name:    "complex Chart.yaml",
			content: "apiVersion: v2\nname: test\ndescription: testing version update against complex Chart.yaml\ntype: application\nkeywords:\n  - testing\n  - version\n  - update\nversion: 1.0.0\nhome: https://apricote.github.io/releaser-pleaser/\ndependencies:\n  - name: somechart\n    version: 1.2.3\n",
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    "apiVersion: v2\nname: test\ndescription: testing version update against complex Chart.yaml\ntype: application\nkeywords:\n  - testing\n  - version\n  - update\nversion: v2.0.0\nhome: https://apricote.github.io/releaser-pleaser/\ndependencies:\n  - name: somechart\n    version: 1.2.3\n",
			wantErr: assert.NoError,
		},
		{
			name:    "invalid yaml",
			content: `not yaml`,
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `not yaml`,
			wantErr: assert.NoError,
		},
		{
			name:    "yaml without version",
			content: `apiVersion: v2\nname: test`,
			info: ReleaseInfo{
				Version: "v2.0.0",
			},
			want:    `apiVersion: v2\nname: test`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUpdaterTest(t, HelmChart(), tt)
		})
	}
}
