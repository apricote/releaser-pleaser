package versioning

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apricote/releaser-pleaser/internal/commitparser"
	"github.com/apricote/releaser-pleaser/internal/git"
)

func TestSemVer_NextVersion(t *testing.T) {
	type args struct {
		releases        git.Releases
		versionBump     VersionBump
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
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (minor)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (patch)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (major)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (minor)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (patch)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (major)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v2.0.0-rc.0"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (minor)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.2.0-rc.0"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (patch)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.2-rc.0"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (major)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.2.0-rc.0"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (minor)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.2-rc.0"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to prerelease",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-alpha.2"},
					Stable: &git.Tag{Name: "v1.1.0"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.1-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (explicit)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-alpha.2"},
					Stable: &git.Tag{Name: "v1.1.0"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeNormal,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (implicit)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-alpha.2"},
					Stable: &git.Tag{Name: "v1.1.0"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (major)",
			args: args{
				releases: git.Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (minor)",
			args: args{
				releases: git.Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.1.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (patch)",
			args: args{
				releases: git.Releases{
					Latest: nil,
					Stable: nil,
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.0.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (major)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (minor)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (patch)",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-rc.0"},
					Stable: nil,
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			// TODO: Is this actually correct our should it be v1.1.1?
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "error on invalid tag semver",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "foodazzle"},
					Stable: &git.Tag{Name: "foodazzle"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid tag prerelease",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1-rc.foo"},
					Stable: &git.Tag{Name: "v1.1.1-rc.foo"},
				},
				versionBump:     PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid bump",
			args: args{
				releases: git.Releases{
					Latest: &git.Tag{Name: "v1.1.1"},
					Stable: &git.Tag{Name: "v1.1.1"},
				},

				versionBump:     UnknownVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SemVer.NextVersion(tt.args.releases, tt.args.versionBump, tt.args.nextVersionType)
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
		analyzedCommits []commitparser.AnalyzedCommit
		want            VersionBump
	}{
		{
			name:            "no entries (unknown)",
			analyzedCommits: []commitparser.AnalyzedCommit{},
			want:            UnknownVersion,
		},
		{
			name:            "non-release type (unknown)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "docs"}},
			want:            UnknownVersion,
		},
		{
			name:            "single breaking (major)",
			analyzedCommits: []commitparser.AnalyzedCommit{{BreakingChange: true}},
			want:            MajorVersion,
		},
		{
			name:            "single feat (minor)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "feat"}},
			want:            MinorVersion,
		},
		{
			name:            "single fix (patch)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "fix"}},
			want:            PatchVersion,
		},
		{
			name:            "multiple entries (major)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "fix"}, {BreakingChange: true}},
			want:            MajorVersion,
		},
		{
			name:            "multiple entries (minor)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "fix"}, {Type: "feat"}},
			want:            MinorVersion,
		},
		{
			name:            "multiple entries (patch)",
			analyzedCommits: []commitparser.AnalyzedCommit{{Type: "docs"}, {Type: "fix"}},
			want:            PatchVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, BumpFromCommits(tt.analyzedCommits), "BumpFromCommits(%v)", tt.analyzedCommits)
		})
	}
}
