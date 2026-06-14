package main

import (
	"os"
	"path/filepath"
	"testing"
)

// fixturePath returns the real face.fpk fixture if present (gitignored, local
// dev only). Tests that need it skip when it is absent so CI stays green.
func fixturePath(t *testing.T) string {
	t.Helper()
	p := filepath.Join("testdata", "boggy-face.fpk")
	if _, err := os.Stat(p); err != nil {
		t.Skipf("fixture %s absent; skipping", p)
	}
	return p
}

func TestFpkParseFixture(t *testing.T) {
	b, err := os.ReadFile(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	f, err := parseFpk(b)
	if err != nil {
		t.Fatalf("parseFpk: %v", err)
	}
	if len(f.entries) == 0 {
		t.Fatal("no entries parsed")
	}
	for i, e := range f.entries {
		t.Logf("entry %d: name=%q nameOff=%d dataOff=%d dataSize=%d", i, e.name, e.nameOffset, e.dataOffset, e.dataSize)
		if e.dataOffset+e.dataSize > uint64(len(b)) {
			t.Errorf("entry %d data [%d,%d) past EOF %d", i, e.dataOffset, e.dataOffset+e.dataSize, len(b))
		}
	}
}
