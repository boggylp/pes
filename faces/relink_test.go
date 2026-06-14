package main

import (
	"bytes"
	"encoding/binary"
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

func TestRelinkRoundTrip(t *testing.T) {
	raw, err := os.ReadFile(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	old, err := findEmbeddedID(raw)
	if err != nil {
		t.Fatal(err)
	}

	for _, newID := range []string{"2147483648", "99999"} { // length-changing, then equal-length
		out, err := relinkFpkBytes(raw, old, newID)
		if err != nil {
			t.Fatalf("relink to %s: %v", newID, err)
		}
		f, err := parseFpk(out)
		if err != nil {
			t.Fatalf("reparse after relink to %s: %v", newID, err)
		}
		// header file size must equal actual length
		if got := binary.LittleEndian.Uint32(out[0x0A:]); int(got) != len(out) {
			t.Errorf("%s: header fileSize %d != len %d", newID, got, len(out))
		}
		for _, e := range f.entries {
			d := out[e.dataOffset : e.dataOffset+e.dataSize]
			if e.dataOffset+e.dataSize > uint64(len(out)) {
				t.Errorf("%s: entry %q data past EOF", newID, e.name)
			}
			if bytes.HasSuffix([]byte(e.name), []byte(".fmdl")) && !bytes.HasPrefix(d, []byte("FMDL")) {
				t.Errorf("%s: entry %q lost FMDL magic at new offset", newID, e.name)
			}
		}
		if bytes.Contains(out, []byte("real/"+old+"/")) {
			t.Errorf("%s: old id path still present", newID)
		}
		if !bytes.Contains(out, []byte("real/"+newID+"/")) {
			t.Errorf("%s: new id path absent", newID)
		}
		// equal-length must preserve total size; length change must grow it
		switch {
		case len(newID) == len(old) && len(out) != len(raw):
			t.Errorf("%s: equal-length relink changed size %d -> %d", newID, len(raw), len(out))
		case len(newID) > len(old) && len(out) <= len(raw):
			t.Errorf("%s: longer id did not grow file (%d -> %d)", newID, len(raw), len(out))
		}
	}
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
