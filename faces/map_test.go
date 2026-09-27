package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDirMirrorsNestedTree(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	files := map[string]string{
		"face.fpk":                   "top",
		filepath.Join("#Win", "a"):   "nested",
		filepath.Join("x", "y", "z"): "deep",
	}
	for rel, body := range files {
		p := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	for rel, want := range files {
		got, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			t.Fatalf("reading copied %s: %v", rel, err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", rel, got, want)
		}
	}
}
