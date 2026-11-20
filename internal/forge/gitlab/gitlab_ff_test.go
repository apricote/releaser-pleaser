package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apricote/releaser-pleaser/internal/forge"
	"github.com/apricote/releaser-pleaser/internal/releasepr"
)

// Commit graph of the scenario, which is ordinary practice on a project using
// the fast-forward merge method:
//
//	A   ── base of the Release MR when releaser-pleaser opened it
//	 \
//	  A' ── the release commit, HEAD of releaser-pleaser--branches--main
//
//	A ── B          another MR merges to main while the Release MR is open
//	      \
//	       B'       merging the Release MR fast-forwards: GitLab REBASES A'
//	                onto B, main becomes B'
//
// After that merge GitLab reports, on the merge request:
//
//	sha               = A'   (the PRE-rebase head — never updated)
//	merge_commit_sha  = null (fast-forward: no merge commit)
//	squash_commit_sha = null (squash not required on the project)
//
// B' is the commit that actually landed on main. A' is on no branch at all.
const (
	branchMain = "main"

	commitPreRebaseReleaseHead = "1c17efd18edcf465c3355653b2cd171d67f1bf2e" // A'
	commitOnMainAfterMerge     = "dd4b82efa9e9c24f192da30a970376e53333b27b" // B'

	releaseVersion = "v0.12.1"
)

// fakeGitLab serves the handful of endpoints PendingReleases and CreateRelease
// touch, reproducing the API responses a real fast-forward merge leaves behind.
// The SHAs above are the real ones from gitlab.com/phpboyscout/go/config!94.
func fakeGitLab(t *testing.T, squashCommitSHA, recordedHead string) (*httptest.Server, *string) {
	t.Helper()

	// Captures the ref releaser-pleaser asks GitLab to create the tag at.
	var createdReleaseRef string

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v4/projects/{project}/merge_requests", func(w http.ResponseWriter, _ *http.Request) {
		mr := map[string]any{
			"iid":               94,
			"title":             fmt.Sprintf("chore(main): release %s", releaseVersion),
			"description":       "",
			"state":             "merged",
			"labels":            []string{releasepr.LabelReleasePending.Name},
			"target_branch":     branchMain,
			"source_branch":     "releaser-pleaser--branches--main",
			"merged_at":         "2026-07-31T08:32:52.536Z",
			"sha":               recordedHead,
			"merge_commit_sha":  "", // fast-forward merge: no merge commit
			"squash_commit_sha": squashCommitSHA,
		}
		writeJSON(t, w, []any{mr})
	})

	// Commits on the target branch in the merge window. After an automatic
	// rebase the landed commit keeps the merge request's message.
	mux.HandleFunc("GET /api/v4/projects/{project}/repository/commits", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, []any{map[string]any{
			"id":             commitOnMainAfterMerge,
			"title":          fmt.Sprintf("chore(main): release %s", releaseVersion),
			"committed_date": "2026-07-31T08:32:51.000Z",
		}})
	})

	// GitLab associates the rebased commit with the merge request it came from.
	mux.HandleFunc("GET /api/v4/projects/{project}/repository/commits/{sha}/merge_requests", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, []any{map[string]any{"iid": 94, "title": fmt.Sprintf("chore(main): release %s", releaseVersion)}})
	})

	// Only commitOnMainAfterMerge is on a branch. The pre-rebase head is on
	// none, which is exactly the state GitLab leaves after rebasing to
	// fast-forward.
	mux.HandleFunc("GET /api/v4/projects/{project}/repository/commits/{sha}/refs", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("sha") == commitOnMainAfterMerge {
			writeJSON(t, w, []any{map[string]any{"type": "branch", "name": branchMain}})
			return
		}
		writeJSON(t, w, []any{})
	})

	mux.HandleFunc("POST /api/v4/projects/{project}/releases", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading release request body: %v", err)
		}
		var opts struct {
			Ref     string `json:"ref"`
			TagName string `json:"tag_name"`
		}
		if err := json.Unmarshal(body, &opts); err != nil {
			t.Errorf("decoding release request body: %v", err)
		}
		createdReleaseRef = opts.Ref
		writeJSON(t, w, map[string]any{"tag_name": opts.TagName})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request to fake GitLab: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv, &createdReleaseRef
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encoding response: %v", err)
	}
}

func newTestGitLab(t *testing.T, apiURL string) *GitLab {
	t.Helper()

	// autodiscover() reads these from the environment and would otherwise
	// override the options below when the tests run inside GitLab CI.
	t.Setenv(EnvAPIURL, "")
	t.Setenv(EnvAPIToken, "")
	t.Setenv(EnvProjectURL, "")
	t.Setenv(EnvProjectPath, "")

	gl, err := New(slog.New(slog.DiscardHandler), &Options{
		Options:  forge.Options{Repository: "phpboyscout/go/config", BaseBranch: branchMain},
		Path:     "phpboyscout/go/config",
		APIURL:   apiURL + "/api/v4",
		APIToken: "test-token",
	})
	if err != nil {
		t.Fatalf("constructing GitLab forge: %v", err)
	}

	return gl
}

// TestPendingReleases_AutomaticRebaseRecoversTheLandedCommit is the regression
// test for the defect.
//
// GitLab 19.2 added "Enable automatic rebase prior to merge", which rebases a
// merge request that is behind its target branch as part of merging it. The
// rebase is not written back to the merge request, so its recorded head is the
// PRE-rebase commit — which is on no branch. #210 made that recorded head the
// release commit for fast-forward projects, so the tag lands on a commit that
// is not on the target branch and the release silently omits everything that
// merged while the release merge request was open.
//
// Nothing fails today: the merge request merges green, the tag is created, the
// release page appears, and only the tagged tree is wrong.
//
// The release must still be cut — refusing would block every release on a
// project using automatic rebase — but at the commit that actually landed.
func TestPendingReleases_AutomaticRebaseRecoversTheLandedCommit(t *testing.T) {
	srv, createdReleaseRef := fakeGitLab(t, "", commitPreRebaseReleaseHead)
	gl := newTestGitLab(t, srv.URL)

	prs, err := gl.PendingReleases(context.Background(), releasepr.LabelReleasePending)
	if err != nil {
		t.Fatalf("PendingReleases: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 pending release, got %d", len(prs))
	}
	if prs[0].ReleaseCommit == nil {
		t.Fatal("pending release has no release commit")
	}

	if err := gl.CreateRelease(
		context.Background(), *prs[0].ReleaseCommit, releaseVersion, "changelog", false, true,
	); err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	if *createdReleaseRef != commitOnMainAfterMerge {
		t.Errorf(
			"release %s was created at %s, but the commit that landed on %s is %s\n"+
				"the tag names a commit that is not on %s, so the released tree is missing\n"+
				"everything that merged while the release merge request was open",
			releaseVersion, *createdReleaseRef, branchMain, commitOnMainAfterMerge, branchMain,
		)
	}
}

// TestPendingReleases_FastForwardUpToDateMergeStillReleases guards the fix
// against over-reach: when the release merge request was NOT behind the target
// branch, a fast-forward merge lands its recorded head unchanged, that commit
// is on the branch, and the release must still be created there. This is the
// case #210 set out to support, and it must keep working.
func TestPendingReleases_FastForwardUpToDateMergeStillReleases(t *testing.T) {
	srv, createdReleaseRef := fakeGitLab(t, "", commitOnMainAfterMerge)
	gl := newTestGitLab(t, srv.URL)

	prs, err := gl.PendingReleases(context.Background(), releasepr.LabelReleasePending)
	if err != nil {
		t.Fatalf("PendingReleases rejected a fast-forward merge whose head IS on %s: %v", branchMain, err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 pending release, got %d", len(prs))
	}

	if err := gl.CreateRelease(
		context.Background(), *prs[0].ReleaseCommit, releaseVersion, "changelog", false, true,
	); err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	if *createdReleaseRef != commitOnMainAfterMerge {
		t.Errorf("release created at %s, want %s", *createdReleaseRef, commitOnMainAfterMerge)
	}
}

// TestPendingReleases_SquashedMergeTagsCommitOnTargetBranch is the control.
//
// With squashing enabled GitLab records a squash_commit_sha, which it creates
// on top of the CURRENT target head — so it is on the branch by construction
// and releaser-pleaser resolves it correctly. This passes both before and after
// the fix, and shows the defect is specific to the fast-forward-without-squash
// path.
func TestPendingReleases_SquashedMergeTagsCommitOnTargetBranch(t *testing.T) {
	srv, createdReleaseRef := fakeGitLab(t, commitOnMainAfterMerge, commitPreRebaseReleaseHead)
	gl := newTestGitLab(t, srv.URL)

	prs, err := gl.PendingReleases(context.Background(), releasepr.LabelReleasePending)
	if err != nil {
		t.Fatalf("PendingReleases: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 pending release, got %d", len(prs))
	}

	if err := gl.CreateRelease(
		context.Background(), *prs[0].ReleaseCommit, releaseVersion, "changelog", false, true,
	); err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	if *createdReleaseRef != commitOnMainAfterMerge {
		t.Errorf("release created at %s, want %s", *createdReleaseRef, commitOnMainAfterMerge)
	}
}
