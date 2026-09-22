package rp

import (
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	withoutExtraPatchTypes := New(nil, slog.Default(), "main", nil, nil, nil, nil)
	assert.Nil(t, withoutExtraPatchTypes.extraPatchTypes)

	pattern := regexp.MustCompile("^(docs|chore)$")
	withExtraPatchTypes := New(nil, slog.Default(), "main", nil, nil, nil, nil, pattern)
	assert.Same(t, pattern, withExtraPatchTypes.extraPatchTypes)
}
