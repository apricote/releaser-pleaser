package markdown

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/ast"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "heading spacing",
			input:   "# Foo\n## Bar\n### Baz",
			want:    "# Foo\n\n## Bar\n\n### Baz\n",
			wantErr: assert.NoError,
		},
		{
			name:    "no empty lines for list items",
			input:   "# Foo\n- 1\n- 2\n",
			want:    "# Foo\n\n- 1\n- 2\n",
			wantErr: assert.NoError,
		},
		{
			name:    "sections",
			input:   "# Foo\n<!-- section-start foobar -->\n- 1\n- 2\n<!-- section-end foobar -->\n",
			want:    "# Foo\n\n<!-- section-start foobar -->\n- 1\n- 2\n\n<!-- section-end foobar -->\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.input)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCodeBlockText(t *testing.T) {
	type args struct {
		source   []byte
		language string
	}
	tests := []struct {
		name      string
		args      args
		wantText  string
		wantFound bool
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "no code block",
			args: args{
				source:   []byte("# Foo"),
				language: "missing",
			},
			wantText:  "",
			wantFound: false,
			wantErr:   assert.NoError,
		},
		{
			name: "code block",
			args: args{
				source:   []byte("```test\nContent\n```"),
				language: "test",
			},
			wantText:  "Content",
			wantFound: true,
			wantErr:   assert.NoError,
		},
		{
			name: "code block with other language",
			args: args{
				source:   []byte("```unknown\nContent\n```"),
				language: "test",
			},
			wantText:  "",
			wantFound: false,
			wantErr:   assert.NoError,
		},
		{
			name: "multiple code blocks with different languages",
			args: args{
				source:   []byte("```unknown\nContent\n```\n\n```test\n1337\n```"),
				language: "test",
			},
			wantText:  "1337",
			wantFound: true,
			wantErr:   assert.NoError,
		},
		{
			name: "multiple code blocks with same language returns first one",
			args: args{
				source:   []byte("```test\nContent\n```\n\n```test\n1337\n```"),
				language: "test",
			},
			wantText:  "Content",
			wantFound: true,
			wantErr:   assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotText string
			var gotFound bool

			err := WalkAST(tt.args.source,
				GetCodeBlockText(tt.args.source, tt.args.language, &gotText, &gotFound),
			)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.wantText, gotText)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestGetSectionText(t *testing.T) {
	type args struct {
		source []byte
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no section",
			args: args{
				source: []byte("# Foo"),
				name:   "missing",
			},
			want:    "",
			wantErr: assert.NoError,
		},
		{
			name: "section",
			args: args{
				source: []byte("<!-- section-start test -->\nContent\n<!-- section-end test -->"),
				name:   "test",
			},
			want:    "Content\n",
			wantErr: assert.NoError,
		},
		{
			name: "section with other name",
			args: args{
				source: []byte("<!-- section-start unknown -->\nContent\n<!-- section-end unknown -->"),
				name:   "test",
			},
			want:    "",
			wantErr: assert.NoError,
		},
		{
			name: "multiple sections with different names",
			args: args{
				source: []byte("<!-- section-start unknown -->\nContent\n<!-- section-end unknown -->\n\n<!-- section-start test -->\n1337\n<!-- section-end test -->"),
				name:   "test",
			},
			want:    "1337\n",
			wantErr: assert.NoError,
		},
		{
			name: "multiple sections with same name returns first one",
			args: args{
				source: []byte("<!-- section-start test -->\nContent\n<!-- section-end test -->\n\n<!-- section-start test -->\n1337\n<!-- section-end test -->"),
				name:   "test",
			},
			want:    "Content\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string

			err := WalkAST(tt.args.source,
				GetSectionText(tt.args.source, tt.args.name, &got),
			)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWalkAST(t *testing.T) {
	type args struct {
		source  []byte
		walkers []ast.Walker
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty walker",
			args: args{
				source: []byte("# Foo"),
				walkers: []ast.Walker{
					func(_ ast.Node, _ bool) (ast.WalkStatus, error) {
						return ast.WalkStop, nil
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "returns walker error",
			args: args{
				source: []byte("# Foo"),
				walkers: []ast.Walker{
					func(_ ast.Node, _ bool) (ast.WalkStatus, error) {
						return ast.WalkStop, errors.New("test")
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "runs all walkers",
			args: args{
				source: []byte("# Foo"),
				walkers: []ast.Walker{
					func(_ ast.Node, _ bool) (ast.WalkStatus, error) {
						return ast.WalkStop, nil
					},
					func(_ ast.Node, _ bool) (ast.WalkStatus, error) {
						return ast.WalkStop, errors.New("test")
					},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WalkAST(tt.args.source, tt.args.walkers...)
			if !tt.wantErr(t, err) {
				return
			}
		})
	}
}
