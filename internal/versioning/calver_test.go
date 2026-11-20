package versioning

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/apricote/releaser-pleaser/internal/git"
)

func fixedNow(t *testing.T, value string) func() time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return func() time.Time { return parsed }
}

func releases(latest, stable string) git.Releases {
	r := git.Releases{}
	if latest != "" {
		r.Latest = &git.Tag{Name: latest}
	}
	if stable != "" {
		r.Stable = &git.Tag{Name: stable}
	}
	return r
}

func TestNewCalVer_InvalidFormat(t *testing.T) {
	tests := []struct {
		format string
		errMsg string
	}{
		{format: "", errMsg: "needs at least one date token"},
		{format: "MICRO", errMsg: "needs at least one date token"},
		{format: "YYYY.0W", errMsg: `last token must be MICRO, got "0W"`},
		{format: "YYYY.MICRO.0W", errMsg: "MICRO must be the last token"},
		{format: "YYYY.MICRO.MICRO", errMsg: "MICRO must be the last token"},
		{format: "YYYY.QQ.MICRO", errMsg: `unknown token "QQ"`},
		{format: "YYYY.yy.MICRO", errMsg: `unknown token "yy"`},
		{format: "YYYY.0Y.MICRO", errMsg: `token "0Y" conflicts with "YYYY"`},
		{format: "YYYY.MM.0M.MICRO", errMsg: `token "0M" conflicts with "MM"`},
		{format: "YYYY.0W.0D.MICRO", errMsg: "week tokens can not be combined with month or day tokens"},
		{format: "YYYY-0W-MICRO", errMsg: "needs at least one date token"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			_, err := NewCalVer(tt.format, nil)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestCalVer_NextVersion(t *testing.T) {
	tests := []struct {
		name            string
		format          string
		now             string
		releases        git.Releases
		nextVersionType NextVersionType
		want            string
	}{
		// YYYY.0W.MICRO
		{
			name:     "first release",
			format:   "YYYY.0W.MICRO",
			now:      "2026-11-18T12:00:00Z",
			releases: releases("", ""),
			want:     "2026.47.0",
		},
		{
			name:     "same week",
			format:   "YYYY.0W.MICRO",
			now:      "2026-11-18T12:00:00Z",
			releases: releases("2026.47.0", "2026.47.0"),
			want:     "2026.47.1",
		},
		{
			name:     "same week, double digit micro",
			format:   "YYYY.0W.MICRO",
			now:      "2026-11-18T12:00:00Z",
			releases: releases("2026.47.9", "2026.47.9"),
			want:     "2026.47.10",
		},
		{
			name:     "next week",
			format:   "YYYY.0W.MICRO",
			now:      "2026-11-25T12:00:00Z",
			releases: releases("2026.47.3", "2026.47.3"),
			want:     "2026.48.0",
		},
		{
			name:     "zero padded week",
			format:   "YYYY.0W.MICRO",
			now:      "2026-02-11T12:00:00Z",
			releases: releases("", ""),
			want:     "2026.07.0",
		},
		{
			name:     "ISO week-year at year boundary",
			format:   "YYYY.0W.MICRO",
			now:      "2027-01-01T12:00:00Z",
			releases: releases("2026.52.0", "2026.52.0"),
			want:     "2026.53.0",
		},
		{
			name:     "ISO week-year at start of year",
			format:   "YYYY.WW.MICRO",
			now:      "2025-12-29T12:00:00Z",
			releases: releases("", ""),
			want:     "2026.1.0",
		},
		{
			name:     "UTC is used",
			format:   "YYYY.0M.0D.MICRO",
			now:      "2026-11-18T23:30:00-05:00",
			releases: releases("", ""),
			want:     "2026.11.19.0",
		},
		// Other formats
		{
			name:     "YY.0M.MICRO",
			format:   "YY.0M.MICRO",
			now:      "2026-03-05T12:00:00Z",
			releases: releases("26.03.0", "26.03.0"),
			want:     "26.03.1",
		},
		{
			name:     "0Y.MM.MICRO",
			format:   "0Y.MM.MICRO",
			now:      "2006-03-05T12:00:00Z",
			releases: releases("", ""),
			want:     "06.3.0",
		},
		{
			name:     "YYYY.MM.DD.MICRO",
			format:   "YYYY.MM.DD.MICRO",
			now:      "2026-03-05T12:00:00Z",
			releases: releases("2026.3.4.2", "2026.3.4.2"),
			want:     "2026.3.5.0",
		},
		// Pre-releases
		{
			name:            "first rc",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.46.0", "2026.46.0"),
			nextVersionType: NextVersionTypeRC,
			want:            "2026.47.0-rc.0",
		},
		{
			name:            "second rc",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0-rc.0", "2026.46.0"),
			nextVersionType: NextVersionTypeRC,
			want:            "2026.47.0-rc.1",
		},
		{
			name:            "rc after stable in same week",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0", "2026.47.0"),
			nextVersionType: NextVersionTypeRC,
			want:            "2026.47.1-rc.0",
		},
		{
			name:            "rc resets on new week",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-25T12:00:00Z",
			releases:        releases("2026.47.0-rc.3", "2026.46.0"),
			nextVersionType: NextVersionTypeRC,
			want:            "2026.48.0-rc.0",
		},
		{
			name:            "beta after rc",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0-rc.1", "2026.46.0"),
			nextVersionType: NextVersionTypeBeta,
			want:            "2026.47.0-beta.0",
		},
		{
			name:            "stable after rc",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0-rc.1", "2026.46.0"),
			nextVersionType: NextVersionTypeNormal,
			want:            "2026.47.0",
		},
		{
			name:            "stable after rc, no previous stable",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0-rc.1", ""),
			nextVersionType: NextVersionTypeNormal,
			want:            "2026.47.0",
		},
		{
			name:            "rc after rc, no previous stable",
			format:          "YYYY.0W.MICRO",
			now:             "2026-11-18T12:00:00Z",
			releases:        releases("2026.47.0-rc.1", ""),
			nextVersionType: NextVersionTypeRC,
			want:            "2026.47.0-rc.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewCalVer(tt.format, fixedNow(t, tt.now))
			require.NoError(t, err)

			got, err := s.NextVersion(tt.releases, UnknownVersion, tt.nextVersionType)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalVer_NextVersion_InvalidTag(t *testing.T) {
	s, err := NewCalVer(DefaultCalVerFormat, nil)
	require.NoError(t, err)

	_, err = s.NextVersion(releases("v1.2.3", "v1.2.3"), UnknownVersion, NextVersionTypeNormal)
	assert.ErrorContains(t, err, `failed to parse version "v1.2.3"`)
}

func TestCalVer_Validate(t *testing.T) {
	s, err := NewCalVer("YYYY.0W.MICRO", nil)
	require.NoError(t, err)

	tests := []struct {
		tag  string
		want bool
	}{
		{"2026.47.0", true},
		{"2026.07.12", true},
		{"2026.47.0-rc.0", true},
		{"2026.47.0-beta.3", true},
		{"2026.47.0-alpha.10", true},
		{"v2026.47.0", false},
		{"v1.2.3", false},
		{"1.2.3", false},
		{"2026.7.0", false},
		{"2026.54.0", false},
		{"2026.00.0", false},
		{"2026.47", false},
		{"2026.47.0.1", false},
		{"2026.47.01", false},
		{"2026.47.0-rc", false},
		{"2026.47.0-rc.01", false},
		{"2026.47.0-foo.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			err := s.Validate(tt.tag)
			if tt.want {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, `does not match calver format "YYYY.0W.MICRO"`)
			}
		})
	}
}

func TestCalVer_IsPrerelease(t *testing.T) {
	s, err := NewCalVer("YYYY.0W.MICRO", nil)
	require.NoError(t, err)

	assert.False(t, s.IsPrerelease("2026.47.0"))
	assert.True(t, s.IsPrerelease("2026.47.0-rc.0"))
	assert.False(t, s.IsPrerelease("invalid"))
}

func TestCalVer_Validate_ShortYear(t *testing.T) {
	s, err := NewCalVer("YY.MM.MICRO", nil)
	require.NoError(t, err)

	require.NoError(t, s.Validate("26.2.0"))
	require.NoError(t, s.Validate("126.2.0"))
	// SemVer tags without a "v" prefix
	require.Error(t, s.Validate("1.5.0"))
	require.Error(t, s.Validate("0.1.0"))
}

func TestCalVer_Compare(t *testing.T) {
	s, err := NewCalVer("YYYY.WW.MICRO", nil)
	require.NoError(t, err)

	tests := []struct {
		a, b string
		want int
	}{
		{"2026.47.0", "2026.47.0", 0},
		{"2026.9.0", "2026.10.0", -1},
		{"2026.10.0", "2026.9.0", 1},
		{"2025.53.3", "2026.1.0", -1},
		{"2026.48.9", "2026.48.10", -1},
		{"2026.48.1-rc.0", "2026.48.1", -1},
		{"2026.48.1-rc.0", "2026.48.0", 1},
		{"2026.48.1-alpha.3", "2026.48.1-beta.0", -1},
		{"2026.48.1-beta.3", "2026.48.1-rc.0", -1},
		{"2026.48.1-rc.10", "2026.48.1-rc.9", 1},
		{"invalid", "2026.1.0", -1},
		{"2026.1.0", "v1.2.3", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+" vs "+tt.b, func(t *testing.T) {
			assert.Equal(t, tt.want, s.Compare(tt.a, tt.b))
		})
	}

	padded, err := NewCalVer("YYYY.0W.MICRO", nil)
	require.NoError(t, err)
	assert.Equal(t, -1, padded.Compare("2026.09.0", "2026.10.0"))
}
