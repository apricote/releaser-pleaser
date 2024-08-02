package rp

import (
	"fmt"
	"testing"

	"github.com/leodido/go-conventionalcommits"
	"github.com/stretchr/testify/assert"
)

func TestNextVersion(t *testing.T) {
	type args struct {
		latestTag       *Tag
		stableTag       *Tag
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
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v2.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (minor)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.2.0",
			wantErr: assert.NoError,
		},
		{
			name: "simple bump (patch)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.2",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (major)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (minor)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "normal to prerelease  (patch)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (major)",
			args: args{
				latestTag:       &Tag{Name: "v2.0.0-rc.0"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (minor)",
			args: args{
				latestTag:       &Tag{Name: "v1.2.0-rc.0"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease bump (patch)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.2-rc.0"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.2-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (major)",
			args: args{
				latestTag:       &Tag{Name: "v1.2.0-rc.0"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v2.0.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease different bump (minor)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.2-rc.0"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.2.0-rc.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to prerelease",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1-alpha.2"},
				stableTag:       &Tag{Name: "v1.1.0"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "v1.1.1-rc.0",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (explicit)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1-alpha.2"},
				stableTag:       &Tag{Name: "v1.1.0"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeNormal,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "prerelease to normal (implicit)",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1-alpha.2"},
				stableTag:       &Tag{Name: "v1.1.0"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.1.1",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (major)",
			args: args{
				latestTag:       nil,
				stableTag:       nil,
				versionBump:     conventionalcommits.MajorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v1.0.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (minor)",
			args: args{
				latestTag:       nil,
				stableTag:       nil,
				versionBump:     conventionalcommits.MinorVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.1.0",
			wantErr: assert.NoError,
		},
		{
			name: "nil tag (patch)",
			args: args{
				latestTag:       nil,
				stableTag:       nil,
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "v0.0.1",
			wantErr: assert.NoError,
		},
		{
			name: "error on invalid tag semver",
			args: args{
				latestTag:       &Tag{Name: "foodazzle"},
				stableTag:       &Tag{Name: "foodazzle"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid tag prerelease",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1-rc.foo"},
				stableTag:       &Tag{Name: "v1.1.1-rc.foo"},
				versionBump:     conventionalcommits.PatchVersion,
				nextVersionType: NextVersionTypeRC,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "error on invalid bump",
			args: args{
				latestTag:       &Tag{Name: "v1.1.1"},
				stableTag:       &Tag{Name: "v1.1.1"},
				versionBump:     conventionalcommits.UnknownVersion,
				nextVersionType: NextVersionTypeUndefined,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextVersion(tt.args.latestTag, tt.args.stableTag, tt.args.versionBump, tt.args.nextVersionType)
			if !tt.wantErr(t, err, fmt.Sprintf("NextVersion(%v, %v, %v, %v)", tt.args.latestTag, tt.args.stableTag, tt.args.versionBump, tt.args.nextVersionType)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NextVersion(%v, %v, %v, %v)", tt.args.latestTag, tt.args.stableTag, tt.args.versionBump, tt.args.nextVersionType)
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
