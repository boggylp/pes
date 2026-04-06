package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCookie_DirectValue(t *testing.T) {
	got := resolveCookie("xf_session=abc", "")
	if got != "xf_session=abc" {
		t.Errorf("resolveCookie() = %q, want %q", got, "xf_session=abc")
	}
}

func TestResolveCookie_DirectTakesPriority(t *testing.T) {
	got := resolveCookie("direct", "somefile.txt")
	if got != "direct" {
		t.Errorf("resolveCookie() = %q, want %q", got, "direct")
	}
}

func TestResolveCookie_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookie.txt")
	os.WriteFile(path, []byte("  xf_user=xyz  \n"), 0600)

	got := resolveCookie("", path)
	if got != "xf_user=xyz" {
		t.Errorf("resolveCookie() = %q, want %q", got, "xf_user=xyz")
	}
}

func TestResolveCookie_Empty(t *testing.T) {
	got := resolveCookie("", "")
	if got != "" {
		t.Errorf("resolveCookie() = %q, want empty", got)
	}
}
