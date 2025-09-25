package testdata

import (
	"embed"
	"testing"
)

//go:embed *.txt
var testdata embed.FS

func MustReadFileString(t *testing.T, name string) string {
	t.Helper()

	content, err := testdata.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
