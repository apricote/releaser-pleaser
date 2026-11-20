package versioning

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/apricote/releaser-pleaser/internal/git"
)

// DefaultCalVerFormat is the format used when the user does not specify one.
const DefaultCalVerFormat = "YYYY.WW.MICRO"

const calVerMicroToken = "MICRO"

type calVerCategory int

const (
	calVerYear calVerCategory = iota
	calVerMonth
	calVerWeek
	calVerDay
)

// calVerToken describes a single date segment in a CalVer format. See https://calver.org/#scheme for the tokens.
type calVerToken struct {
	category calVerCategory
	// pattern matches valid values of this token in a version string.
	pattern string
	// render formats the value of the date component for this token.
	render func(value int) string
}

func calVerUnpadded(value int) string { return strconv.Itoa(value) }
func calVerPadded(value int) string   { return fmt.Sprintf("%02d", value) }

// YY only matches years since 2010, so that SemVer tags without a "v" prefix (e.g. "1.5.0") are not valid versions
// of a "YY.MM.MICRO" format.
var calVerTokens = map[string]calVerToken{
	"YYYY": {calVerYear, `\d{4}`, calVerUnpadded},
	"YY":   {calVerYear, `[1-9]\d+`, func(year int) string { return calVerUnpadded(year - 2000) }},
	"0Y":   {calVerYear, `\d{2}|[1-9]\d{2,}`, func(year int) string { return calVerPadded(year - 2000) }},
	"MM":   {calVerMonth, `[1-9]|1[0-2]`, calVerUnpadded},
	"0M":   {calVerMonth, `0[1-9]|1[0-2]`, calVerPadded},
	"WW":   {calVerWeek, `[1-9]|[1-4]\d|5[0-3]`, calVerUnpadded},
	"0W":   {calVerWeek, `0[1-9]|[1-4]\d|5[0-3]`, calVerPadded},
	"DD":   {calVerDay, `[1-9]|[12]\d|3[01]`, calVerUnpadded},
	"0D":   {calVerDay, `0[1-9]|[12]\d|3[01]`, calVerPadded},
}

type calVer struct {
	format   string
	segments []calVerToken
	usesWeek bool
	regex    *regexp.Regexp
	now      func() time.Time
}

type calVerVersion struct {
	date   []string
	micro  uint64
	pre    string
	preNum uint64
}

func (v calVerVersion) core() string {
	return strings.Join(v.date, ".") + "." + strconv.FormatUint(v.micro, 10)
}

func (v calVerVersion) String() string {
	if v.pre == "" {
		return v.core()
	}
	return fmt.Sprintf("%s-%s.%d", v.core(), v.pre, v.preNum)
}

// NewCalVer returns a Calendar Versioning Strategy for the given format, for example "YYYY.WW.MICRO".
//
// The format consists of date tokens (YYYY, YY, 0Y, MM, 0M, WW, 0W, DD, 0D) separated by ".", and it must end with
// the MICRO token. All dates are in UTC. If the format contains a week token, the year tokens use the ISO 8601
// week-numbering year.
//
// now is used to get the current date, if it is nil time.Now is used.
func NewCalVer(format string, now func() time.Time) (Strategy, error) {
	if now == nil {
		now = time.Now
	}

	parts := strings.Split(format, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid calver format %q: needs at least one date token and %s, e.g. %q", format, calVerMicroToken, DefaultCalVerFormat)
	}

	c := &calVer{format: format, now: now}
	seen := map[calVerCategory]string{}
	patterns := make([]string, 0, len(parts))

	for i, part := range parts {
		if part == calVerMicroToken {
			if i != len(parts)-1 {
				return nil, fmt.Errorf("invalid calver format %q: %s must be the last token", format, calVerMicroToken)
			}
			continue
		}
		if i == len(parts)-1 {
			return nil, fmt.Errorf("invalid calver format %q: last token must be %s, got %q", format, calVerMicroToken, part)
		}

		token, ok := calVerTokens[part]
		if !ok {
			return nil, fmt.Errorf("invalid calver format %q: unknown token %q", format, part)
		}
		if previous, ok := seen[token.category]; ok {
			return nil, fmt.Errorf("invalid calver format %q: token %q conflicts with %q", format, part, previous)
		}
		seen[token.category] = part

		c.segments = append(c.segments, token)
		patterns = append(patterns, "("+token.pattern+")")
	}

	_, hasWeek := seen[calVerWeek]
	_, hasMonth := seen[calVerMonth]
	_, hasDay := seen[calVerDay]
	if hasWeek && (hasMonth || hasDay) {
		return nil, fmt.Errorf("invalid calver format %q: week tokens can not be combined with month or day tokens", format)
	}
	c.usesWeek = hasWeek

	patterns = append(patterns, `(0|[1-9]\d*)`)
	c.regex = regexp.MustCompile(`^` + strings.Join(patterns, `\.`) + `(?:-(alpha|beta|rc)\.(0|[1-9]\d*))?$`)

	return c, nil
}

func (c *calVer) render(t time.Time) []string {
	t = t.UTC()
	isoYear, isoWeek := t.ISOWeek()

	year := t.Year()
	if c.usesWeek {
		year = isoYear
	}

	values := map[calVerCategory]int{
		calVerYear:  year,
		calVerMonth: int(t.Month()),
		calVerWeek:  isoWeek,
		calVerDay:   t.Day(),
	}

	date := make([]string, 0, len(c.segments))
	for _, segment := range c.segments {
		date = append(date, segment.render(values[segment.category]))
	}
	return date
}

func (c *calVer) parse(version string) (calVerVersion, bool) {
	matches := c.regex.FindStringSubmatch(version)
	if matches == nil {
		return calVerVersion{}, false
	}

	// matches: [full, date segments..., micro, preType, preNum]
	n := len(c.segments)
	v := calVerVersion{
		date: matches[1 : n+1],
		pre:  matches[n+2],
	}

	var err error
	v.micro, err = strconv.ParseUint(matches[n+1], 10, 64)
	if err != nil {
		return calVerVersion{}, false
	}

	if v.pre != "" {
		v.preNum, err = strconv.ParseUint(matches[n+3], 10, 64)
		if err != nil {
			return calVerVersion{}, false
		}
	}

	return v, true
}

func (c *calVer) parseTag(tag *git.Tag) (*calVerVersion, error) {
	if tag == nil {
		return nil, nil
	}

	v, ok := c.parse(tag.Name)
	if !ok {
		return nil, fmt.Errorf("failed to parse version %q with calver format %q", tag.Name, c.format)
	}
	return &v, nil
}

func (c *calVer) NextVersion(r git.Releases, _ VersionBump, nextVersionType NextVersionType) (string, error) {
	latest, err := c.parseTag(r.Latest)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest version: %w", err)
	}

	stable, err := c.parseTag(r.Stable)
	if err != nil {
		return "", fmt.Errorf("failed to parse stable version: %w", err)
	}

	// If there is a previous stable release, we use that as the version anchor. Falling back to any pre-releases
	// if they are the only tags in the repo.
	anchor := latest
	if stable != nil {
		anchor = stable
	}

	next := calVerVersion{date: c.render(c.now())}

	if anchor != nil && slices.Equal(anchor.date, next.date) {
		if anchor.pre == "" {
			next.micro = anchor.micro + 1
		} else {
			// Only pre-releases exist for this period, the stable release uses the same core version.
			next.micro = anchor.micro
		}
	}

	switch nextVersionType {
	case NextVersionTypeUndefined, NextVersionTypeNormal:
		// Stable release, no pre-release suffix
	case NextVersionTypeAlpha, NextVersionTypeBeta, NextVersionTypeRC:
		next.pre = nextVersionType.String()

		if latest != nil && latest.pre == next.pre && latest.core() == next.core() {
			next.preNum = latest.preNum + 1
		}
	default:
		return "", fmt.Errorf("unknown next version type %q", nextVersionType)
	}

	return next.String(), nil
}

func (c *calVer) IsPrerelease(version string) bool {
	v, ok := c.parse(version)
	return ok && v.pre != ""
}

func (c *calVer) Validate(tag string) error {
	if _, ok := c.parse(tag); !ok {
		return fmt.Errorf("does not match calver format %q", c.format)
	}
	return nil
}

// Compare sorts invalid versions before all valid versions.
func (c *calVer) Compare(a, b string) int {
	va, okA := c.parse(a)
	vb, okB := c.parse(b)
	switch {
	case !okA && !okB:
		return 0
	case !okA:
		return -1
	case !okB:
		return 1
	}

	for i := range va.date {
		// The patterns only match decimal numbers, padded and unpadded values compare the same.
		da, _ := strconv.Atoi(va.date[i])
		db, _ := strconv.Atoi(vb.date[i])
		if d := cmp.Compare(da, db); d != 0 {
			return d
		}
	}

	if d := cmp.Compare(va.micro, vb.micro); d != 0 {
		return d
	}

	// A stable release is newer than its pre-releases.
	if d := cmp.Compare(calVerPreOrder[va.pre], calVerPreOrder[vb.pre]); d != 0 {
		return d
	}

	return cmp.Compare(va.preNum, vb.preNum)
}

var calVerPreOrder = map[string]int{
	NextVersionTypeAlpha.String(): 0,
	NextVersionTypeBeta.String():  1,
	NextVersionTypeRC.String():    2,
	"":                            3,
}
