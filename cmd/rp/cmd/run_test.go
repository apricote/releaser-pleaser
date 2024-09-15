package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
