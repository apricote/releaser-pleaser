package git

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestAuthor_signature(t *testing.T) {
	now := time.Now()

	tests := []struct {
		author Author
		want   *object.Signature
	}{
		{author: Author{Name: "foo", Email: "bar@example.com"}, want: &object.Signature{Name: "foo", Email: "bar@example.com", When: now}},
		{author: Author{Name: "bar", Email: "foo@example.com"}, want: &object.Signature{Name: "bar", Email: "foo@example.com", When: now}},
	}
	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			if got := tt.author.signature(now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("signature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthor_String(t *testing.T) {
	tests := []struct {
		author Author
		want   string
	}{
		{author: Author{Name: "foo", Email: "bar@example.com"}, want: "foo <bar@example.com>"},
		{author: Author{Name: "bar", Email: "foo@example.com"}, want: "bar <foo@example.com>"},
	}
	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			if got := tt.author.String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
