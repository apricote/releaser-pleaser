package rp

import (
	"fmt"
	"testing"

	"github.com/leodido/go-conventionalcommits"
	"github.com/stretchr/testify/assert"
)

func TestReleases_NextVersion(t *testing.T) {
	type args struct {
		versionBump     conventionalcommits.VersionBump
		nextVersionType NextVersionType
	}
	tests := []struct {
		name     string
		releases Releases
		args     args
		want     string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "simple bump (major)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (minor)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (patch)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (major)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (minor)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (patch)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (major)",
			releases: Releases{
				Latest: &Tag{Name: "v2.0.0-rc.0"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (minor)",
			releases: Releases{
				Latest: &Tag{Name: "v1.2.0-rc.0"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (patch)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.2-rc.0"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (major)",
			releases: Releases{
				Latest: &Tag{Name: "v1.2.0-rc.0"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (minor)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.2-rc.0"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to prerelease",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-alpha.2"},
				Stable: &Tag{Name: "v1.1.0"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.1-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (explicit)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-alpha.2"},
				Stable: &Tag{Name: "v1.1.0"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeNormal,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (implicit)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-alpha.2"},
				Stable: &Tag{Name: "v1.1.0"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (major)",
			releases: Releases{
				Latest: nil,
				Stable: nil,
			},
			args: args{
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (minor)",
			releases: Releases{
				Latest: nil,
				Stable: nil,
			},
			args: args{
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.1.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (patch)",
			releases: Releases{
				Latest: nil,
				Stable: nil,
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.0.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (major)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-rc.0"},
				Stable: nil,
			},
			args: args{
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (minor)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-rc.0"},
				Stable: nil,
			},
			args: args{

				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil stable release (patch)",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-rc.0"},
				Stable: nil,
			},
			args: args{

				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			// TODO: Is this actually correct our should it be v1.1.1?
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "error on invalid tag semver",
			releases: Releases{
				Latest: &Tag{Name: "foodazzle"},
				Stable: &Tag{Name: "foodazzle"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid tag prerelease",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1-rc.foo"},
				Stable: &Tag{Name: "v1.1.1-rc.foo"},
			},
			args: args{
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid bump",
			releases: Releases{
				Latest: &Tag{Name: "v1.1.1"},
				Stable: &Tag{Name: "v1.1.1"},
			},
			args: args{

				versionBump:     conventionalcommits.UnknownVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.releases.NextVersion(tt.args.versionBump, tt.args.nextVersionType)
			if !tt.wantErr(t, err, fmt.Sprintf("Releases(%v, %v).NextVersion(%v, %v)", tt.releases.Latest, tt.releases.Stable, tt.args.versionBump, tt.args.nextVersionType)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Releases(%v, %v).NextVersion(%v, %v)", tt.releases.Latest, tt.releases.Stable, tt.args.versionBump, tt.args.nextVersionType)
		})
	}
}

func TestVersionBumpFromChangesets(t *testing.T) {
	tests := []struct {
		name       string
		changesets []Changeset
		want       conventionalcommits.VersionBump
	}{
		{
			name:       "no entries (unknown)",
			changesets: []Changeset{},
			want:       conventionalcommits.UnknownVersion,
		},
		{
			name:       "non-release type (unknown)",
			changesets: []Changeset{{ChangelogEntries: []AnalyzedCommit{{Type: "docs"}}}},
			want:       conventionalcommits.UnknownVersion,
		},
		{
			name:       "single breaking (major)",
			changesets: []Changeset{{ChangelogEntries: []AnalyzedCommit{{BreakingChange: true}}}},
			want:       conventionalcommits.MajorVersion,
		},
		{
			name:       "single feat (minor)",
			changesets: []Changeset{{ChangelogEntries: []AnalyzedCommit{{Type: "feat"}}}},
			want:       conventionalcommits.MinorVersion,
		},
		{
			name:       "single fix (patch)",
			changesets: []Changeset{{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}}}},
			want:       conventionalcommits.PatchVersion,
		},
		{
			name: "multiple changesets (major)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}}},
				{ChangelogEntries: []AnalyzedCommit{{BreakingChange: true}}},
			},
			want: conventionalcommits.MajorVersion,
		},
		{
			name: "multiple changesets (minor)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}}},
				{ChangelogEntries: []AnalyzedCommit{{Type: "feat"}}},
			},
			want: conventionalcommits.MinorVersion,
		},
		{
			name: "multiple changesets (patch)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "docs"}}},
				{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}}},
			},
			want: conventionalcommits.PatchVersion,
		},
		{
			name: "multiple entries in one changeset (major)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}, {BreakingChange: true}}},
			},
			want: conventionalcommits.MajorVersion,
		},
		{
			name: "multiple entries in one changeset (minor)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "fix"}, {Type: "feat"}}},
			},
			want: conventionalcommits.MinorVersion,
		},
		{
			name: "multiple entries in one changeset (patch)",
			changesets: []Changeset{
				{ChangelogEntries: []AnalyzedCommit{{Type: "docs"}, {Type: "fix"}}},
			},
			want: conventionalcommits.PatchVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, VersionBumpFromChangesets(tt.changesets), "VersionBumpFromChangesets(%v)", tt.changesets)
		})
	}
}
