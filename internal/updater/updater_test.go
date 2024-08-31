package updater

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type updaterTestCase struct {
	name    string
	content string
	info    ReleaseInfo
	want    string
	wantErr assert.ErrorAssertionFunc
}

func runUpdaterTest(t *testing.T, constructor NewUpdater, tt updaterTestCase) {
	t.Helper()

	got, err := constructor(tt.info)(tt.content)
	if !tt.wantErr(t, err, fmt.Sprintf("Updater(%v, %v)", tt.content, tt.info)) {
		return
	}
	assert.Equalf(t, tt.want, got, "Updater(%v, %v)", tt.content, tt.info)
}
