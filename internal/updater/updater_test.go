package updater

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type updaterTestCase struct {
	name     string
	content  string
	filename string
	info     ReleaseInfo
	want     string
	wantErr  assert.ErrorAssertionFunc
}

func runUpdaterTest(t *testing.T, constructor NewUpdater, tt updaterTestCase) {
	t.Helper()

	got, err := constructor(tt.info)(tt.content, tt.filename)
	if !tt.wantErr(t, err, fmt.Sprintf("Updater(%v, %v, %v)", tt.content, tt.filename, tt.info)) {
		return
	}
	assert.Equalf(t, tt.want, got, "Updater(%v, %v, %v)", tt.content, tt.filename, tt.info)
}
