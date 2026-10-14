package gitlab

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitAuthor(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantName  string
		wantEmail string
	}{
		{
			name:      "prefers commit email",
			body:      `{"username":"jdoe","name":"Jane Doe","email":"primary@example.com","public_email":"public@example.com","commit_email":"commit@example.com"}`,
			wantName:  "Jane Doe",
			wantEmail: "commit@example.com",
		},
		{
			name:      "falls back to primary email",
			body:      `{"username":"jdoe","name":"Jane Doe","email":"primary@example.com","public_email":"public@example.com"}`,
			wantName:  "Jane Doe",
			wantEmail: "primary@example.com",
		},
		{
			name:      "falls back to public email and username",
			body:      `{"username":"jdoe","public_email":"public@example.com"}`,
			wantName:  "jdoe",
			wantEmail: "public@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v4/user", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			forge, err := New(slog.New(slog.NewTextHandler(io.Discard, nil)), &Options{APIURL: server.URL, APIToken: "token"})
			require.NoError(t, err)

			author, err := forge.CommitAuthor(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, author.Name)
			assert.Equal(t, tt.wantEmail, author.Email)
		})
	}
}
