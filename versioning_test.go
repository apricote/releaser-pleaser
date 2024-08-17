package rp

import (
	"fmt"
	"testing"

	"github.com/leodido/go-conventionalcommits"
	"github.com/stretchr/testify/assert"
)

func TestReleases_NextVersion(t *testing.T) {
	type args struct {
		releases        Releases
		versionBump     conventionalcommits.VersionBump
		nextVersionType NextVersionType
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "simple bump (major)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (minor)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (patch)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (major)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (minor)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (patch)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (major)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v2.0.0-rc.0"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (minor)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.2.0-rc.0"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (patch)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.2-rc.0"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (major)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.2.0-rc.0"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (minor)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.2-rc.0"},
					Stable: &Tag{Name: "v1.1.1"},
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to prerelease",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-alpha.2"},
					Stable: &Tag{Name: "v1.1.0"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.1-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (explicit)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-alpha.2"},
					Stable: &Tag{Name: "v1.1.0"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeNormal,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (implicit)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-alpha.2"},
					Stable: &Tag{Name: "v1.1.0"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (major)",
			args: args{
				releases: Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (minor)",
			args: args{
				releases: Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.1.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (patch)",
			args: args{
				releases: Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.0.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (major)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (minor)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (patch)",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			// TODO: Is this actually correct our should it be v1.1.1?
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "error on invalid tag semver",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "foodazzle"},
					Stable: &Tag{Name: "foodazzle"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid tag prerelease",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1-rc.foo"},
					Stable: &Tag{Name: "v1.1.1-rc.foo"},
				},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid bump",
			args: args{
				releases: Releases{
					Latest: &Tag{Name: "v1.1.1"},
					Stable: &Tag{Name: "v1.1.1"},
				},

				versionBump:     conventionalcommits.UnknownVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SemVerNextVersion(tt.args.releases, tt.args.versionBump, tt.args.nextVersionType)
			if !tt.wantErr(t, err, fmt.Sprintf("SemVerNextVersion(Releases(%v, %v), %v, %v)", tt.args.releases.Latest, tt.args.releases.Stable, tt.args.versionBump, tt.args.nextVersionType)) {
				return
			}
			assert.Equalf(t, tt.want, got, "SemVerNextVersion(Releases(%v, %v), %v, %v)", tt.args.releases.Latest, tt.args.releases.Stable, tt.args.versionBump, tt.args.nextVersionType)
		})
	}
}

func TestVersionBumpFromCommits(t *testing.T) {
	tests := []struct {
		name            string
		analyzedCommits []AnalyzedCommit
		want            conventionalcommits.VersionBump
	}{
		{
			name:            "no entries (unknown)",
			analyzedCommits: []AnalyzedCommit{},
			want:            conventionalcommits.UnknownVersion,
		},
		{
			name:            "non-release type (unknown)",
			analyzedCommits: []AnalyzedCommit{{Type: "docs"}},
			want:            conventionalcommits.UnknownVersion,
		},
		{
			name:            "single breaking (major)",
			analyzedCommits: []AnalyzedCommit{{BreakingChange: true}},
			want:            conventionalcommits.MajorVersion,
		},
		{
			name:            "single feat (minor)",
			analyzedCommits: []AnalyzedCommit{{Type: "feat"}},
			want:            conventionalcommits.MinorVersion,
		},
		{
			name:            "single fix (patch)",
			analyzedCommits: []AnalyzedCommit{{Type: "fix"}},
			want:            conventionalcommits.PatchVersion,
		},
		{
			name:            "multiple entries (major)",
			analyzedCommits: []AnalyzedCommit{{Type: "fix"}, {BreakingChange: true}},
			want:            conventionalcommits.MajorVersion,
		},
		{
			name:            "multiple entries (minor)",
			analyzedCommits: []AnalyzedCommit{{Type: "fix"}, {Type: "feat"}},
			want:            conventionalcommits.MinorVersion,
		},
		{
			name:            "multiple entries (patch)",
			analyzedCommits: []AnalyzedCommit{{Type: "docs"}, {Type: "fix"}},
			want:            conventionalcommits.PatchVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, VersionBumpFromCommits(tt.analyzedCommits), "VersionBumpFromCommits(%v)", tt.analyzedCommits)
		})
	}
}
