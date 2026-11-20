package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
	"github.com/apricote/releaser-pleaser/internal/versioning"
)

func Test_parseExtraFiles(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: ``,
			want:  []string{},
		},
		{
			name:  "empty quoted",
			input: `""`,
			want:  []string{},
		},
		{
			name:  "single",
			input: `foo.txt`,
			want:  []string{"foo.txt"},
		},
		{
			name:  "single quoted",
			input: `"foo.txt"`,
			want:  []string{"foo.txt"},
		},
		{
			name: "multiple",
			input: `foo.txt
dir/Chart.yaml`,
			want: []string{"foo.txt", "dir/Chart.yaml"},
		},
		{
			name: "multiple quoted",
			input: `"foo.txt
dir/Chart.yaml"`,
			want: []string{"foo.txt", "dir/Chart.yaml"},
		},
		{
			name:  "multiple with broken new lines",
			input: `"action.yml\ntemplates/run.yml\n"`,
			want:  []string{"action.yml", "templates/run.yml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExtraFiles(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseUpdaters(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "empty",
			input: []string{},
			want:  []string{"changelog", "generic"},
		},
		{
			name:  "remove defaults",
			input: []string{"-changelog", "-generic"},
			want:  []string{},
		},
		{
			name:  "remove unknown is ignored",
			input: []string{"-fooo"},
			want:  []string{"changelog", "generic"},
		},
		{
			name:  "add new entry",
			input: []string{"bar"},
			want:  []string{"bar", "changelog", "generic"},
		},
		{
			name:  "duplicates are removed",
			input: []string{"bar", "bar", "changelog"},
			want:  []string{"bar", "changelog", "generic"},
		},
		{
			name:  "remove empty entries",
			input: []string{""},
			want:  []string{"changelog", "generic"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUpdaters(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseVersioning(t *testing.T) {
	tests := []struct {
		name         string
		versioning   string
		calVerFormat string
		now          string
		wantSemVer   bool
		wantErr      string
	}{
		{name: "defaults", wantSemVer: true},
		{name: "semver", versioning: "semver", wantSemVer: true},
		{name: "calver", versioning: "calver"},
		{name: "calver with format", versioning: "calver", calVerFormat: "YYYY.0M.MICRO", now: "2026-11-18T12:00:00Z"},
		{name: "unknown versioning", versioning: "foo", wantErr: "unknown --versioning: foo"},
		{name: "invalid calver format", versioning: "calver", calVerFormat: "YYYY.QQ.MICRO", wantErr: `invalid --calver-format: invalid calver format "YYYY.QQ.MICRO": unknown token "QQ"`},
		{name: "invalid now", versioning: "calver", now: "yesterday", wantErr: "invalid --now"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := parseVersioning(tt.versioning, tt.calVerFormat, tt.now)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSemVer, strategy == versioning.SemVer)
		})
	}
}

func Test_parseVersioning_CalVerNow(t *testing.T) {
	// The default format does not pad the week, 2026-02-11 is in ISO week 7.
	strategy, err := parseVersioning("calver", "", "2026-02-11T12:00:00Z")
	require.NoError(t, err)

	version, err := strategy.NextVersion(git.Releases{}, versioning.UnknownVersion, versioning.NextVersionTypeNormal)
	require.NoError(t, err)
	assert.Equal(t, "2026.7.0", version)
}
