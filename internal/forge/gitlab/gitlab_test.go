package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
)

// page describes a single canned API response for the fake list function below.
type page struct {
	items      []int
	totalPages int64
	nextPage   int64
}

// fakeList returns a list function for all() that serves the canned responses and records
// which pages were requested.
func fakeList(t *testing.T, pages map[int64]page) (func(gitlab.ListOptions) ([]int, *gitlab.Response, error), *[]int64) {
	t.Helper()

	requested := make([]int64, 0)

	return func(listOptions gitlab.ListOptions) ([]int, *gitlab.Response, error) {
		requested = append(requested, listOptions.Page)

		p, ok := pages[listOptions.Page]
		if !ok {
			t.Fatalf("unexpected request for page %d", listOptions.Page)
		}

		return p.items, &gitlab.Response{
			TotalPages:  p.totalPages,
			CurrentPage: listOptions.Page,
			NextPage:    p.nextPage,
		}, nil
	}, &requested
}

func TestAll(t *testing.T) {
	tests := []struct {
		name          string
		pages         map[int64]page
		want          []int
		wantRequested []int64
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "empty repository",
			pages: map[int64]page{
				1: {items: []int{}, totalPages: 0, nextPage: 0},
			},
			want:          []int{},
			wantRequested: []int64{1},
			wantErr:       assert.NoError,
		},
		{
			name: "single page",
			pages: map[int64]page{
				1: {items: []int{1, 2, 3}, totalPages: 1, nextPage: 0},
			},
			want:          []int{1, 2, 3},
			wantRequested: []int64{1},
			wantErr:       assert.NoError,
		},
		{
			name: "multiple pages",
			pages: map[int64]page{
				1: {items: []int{1, 2}, totalPages: 3, nextPage: 2},
				2: {items: []int{3, 4}, totalPages: 3, nextPage: 3},
				3: {items: []int{5}, totalPages: 3, nextPage: 0},
			},
			want:          []int{1, 2, 3, 4, 5},
			wantRequested: []int64{1, 2, 3},
			wantErr:       assert.NoError,
		},
		{
			// https://github.com/apricote/releaser-pleaser/issues/488
			// The last page disagrees with the X-Total-Pages of the first page and has no
			// next page. This must terminate instead of falling back to page 0 forever.
			name: "inconsistent total pages on last page",
			pages: map[int64]page{
				1: {items: []int{1, 2}, totalPages: 2, nextPage: 2},
				2: {items: []int{}, totalPages: 1, nextPage: 0},
			},
			want:          []int{1, 2},
			wantRequested: []int64{1, 2},
			wantErr:       assert.NoError,
		},
		{
			name: "next page repeats the current page",
			pages: map[int64]page{
				1: {items: []int{1, 2}, totalPages: 2, nextPage: 1},
			},
			want:          nil,
			wantRequested: []int64{1},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorIs(t, err, ErrPaginationNotAdvancing)
			},
		},
		{
			name: "next page goes backwards",
			pages: map[int64]page{
				1: {items: []int{1, 2}, totalPages: 3, nextPage: 2},
				2: {items: []int{3, 4}, totalPages: 3, nextPage: 1},
			},
			want:          nil,
			wantRequested: []int64{1, 2},
			wantErr: func(t assert.TestingT, err error, _ ...any) bool {
				return assert.ErrorIs(t, err, ErrPaginationNotAdvancing)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, requested := fakeList(t, tt.pages)

			got, err := all(list)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantRequested, *requested)
			assert.NotContains(t, *requested, int64(0), "page 0 must never be requested")
		})
	}
}

func TestAllPropagatesError(t *testing.T) {
	wantErr := assert.AnError
	calls := 0

	got, err := all(func(_ gitlab.ListOptions) ([]int, *gitlab.Response, error) {
		calls++
		return nil, nil, wantErr
	})

	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
	assert.Equal(t, 1, calls)
}
