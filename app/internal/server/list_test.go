package server

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// A machine that has run the local edition has its database beside the
// servers'. The signed-in edition must not offer it as one.
func TestListServerDomainsSkipsLocalEdition(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("APP_ENV", "production")

	for _, domain := range []string{"api.us-east.select-db.com", LocalDomain} {
		path, err := ServerDBPath(domain)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	domains, err := ListServerDomains()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(domains, []string{"api.us-east.select-db.com"}) {
		t.Fatalf("domains = %v", domains)
	}
}
