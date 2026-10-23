package updater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The local edition asks GitHub, whose tag carries the "v" the download URL and
// the signature check add back themselves.
func TestLatestReleaseFromGitHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.4.2","draft":false,"prerelease":false}`))
	}))
	t.Cleanup(srv.Close)
	old := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = old })
	t.Setenv("APP_ENV", "production")

	ver, err := fetchLatestRelease()
	if err != nil {
		t.Fatalf("fetchLatestRelease: %v", err)
	}
	if ver.LatestAppVersion != "1.4.2" {
		t.Fatalf("LatestAppVersion = %q, want 1.4.2", ver.LatestAppVersion)
	}
	if got := artifactURL(ver.LatestAppVersion, ver.ReleaseBaseURL); !strings.HasPrefix(got, "https://github.com/select-db/select/releases/download/v1.4.2/selectDb-") {
		t.Fatalf("artifactURL = %q", got)
	}
}

// staging-latest has no version to compare, so a local staging build stays put.
func TestLatestReleaseSkipsStaging(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	if _, err := fetchLatestRelease(); err == nil {
		t.Fatal("want an error on the staging channel")
	}
}
