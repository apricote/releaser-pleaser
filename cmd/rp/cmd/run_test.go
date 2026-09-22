package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_parseExtraPatchTypes(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		matches    []string
		nonMatches []string
		wantErr    bool
	}{
		{name: "empty", input: ""},
		{name: "alternatives", input: "refactor|perf|chore|docs", matches: []string{"refactor", "perf", "chore", "docs"}, nonMatches: []string{"test", "myrefactor", "docs-extra"}},
		{name: "whole type only", input: "ref", matches: []string{"ref"}, nonMatches: []string{"refactor"}},
		{name: "wildcard", input: "ref.*", matches: []string{"refactor"}, nonMatches: []string{"perf", "myrefactor"}},
		{name: "all types", input: ".*", matches: []string{"feat", "fix", "docs", "chore"}},
		{name: "explicit anchors", input: "^(docs|chore)$", matches: []string{"docs", "chore"}, nonMatches: []string{"mydocs", "chore-extra"}},
		{name: "invalid expression", input: "[", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRunCommand()
			require.NoError(t, cmd.ParseFlags([]string{"--extra-patch-types=" + tt.input}))
			input, err := cmd.PersistentFlags().GetString("extra-patch-types")
			require.NoError(t, err)
			got, err := parseExtraPatchTypes(input)
			if tt.wantErr {
				require.ErrorContains(t, err, "invalid --extra-patch-types pattern")
				return
			}
			require.NoError(t, err)
			if tt.input == "" {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			for _, commitType := range tt.matches {
				assert.True(t, got.MatchString(commitType), "pattern should match %q", commitType)
			}
			for _, commitType := range tt.nonMatches {
				assert.False(t, got.MatchString(commitType), "pattern should not match %q", commitType)
			}
		})
	}
}

func TestRunRejectsInvalidExtraPatchTypes(t *testing.T) {
	cmd := newRunCommand()
	cmd.SetArgs([]string{"--extra-patch-types=["})
	err := cmd.ExecuteContext(t.Context())
	require.ErrorContains(t, err, "invalid --extra-patch-types pattern")
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
