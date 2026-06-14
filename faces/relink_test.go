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

// assertRelink runs relinkFpkBytes and checks the result is a consistent foxfpk
// with the ID swapped and ".fmdl" entries still starting with their magic.
func assertRelink(t *testing.T, raw []byte, old, newID string) {
	t.Helper()
	out, err := relinkFpkBytes(raw, old, newID)
	if err != nil {
		t.Fatalf("relink %s->%s: %v", old, newID, err)
	}
	f, err := parseFpk(out)
	if err != nil {
		t.Fatalf("reparse after relink to %s: %v", newID, err)
	}
	if got := binary.LittleEndian.Uint32(out[0x0A:]); int(got) != len(out) {
		t.Errorf("%s: header fileSize %d != len %d", newID, got, len(out))
	}
	for _, e := range f.entries {
		if e.dataOffset+e.dataSize > uint64(len(out)) {
			t.Errorf("%s: entry %q data past EOF", newID, e.name)
			continue
		}
		d := out[e.dataOffset : e.dataOffset+e.dataSize]
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
	switch {
	case len(newID) == len(old) && len(out) != len(raw):
		t.Errorf("%s: equal-length relink changed size %d -> %d", newID, len(raw), len(out))
	case len(newID) > len(old) && len(out) <= len(raw):
		t.Errorf("%s: longer id did not grow file (%d -> %d)", newID, len(raw), len(out))
	}
}

// TestRelinkSynthetic exercises the repack on an in-code foxfpk, so CI covers
// the logic without the gitignored real fixture.
func TestRelinkSynthetic(t *testing.T) {
	raw := buildSyntheticFpk(t)
	old, err := findEmbeddedID(raw)
	if err != nil {
		t.Fatalf("findEmbeddedID on synthetic: %v", err)
	}
	if old != "21665" {
		t.Fatalf("synthetic embedded id = %q, want 21665", old)
	}
	assertRelink(t, raw, old, "2147483648") // length-changing
	assertRelink(t, raw, old, "99999")      // equal-length
}

// TestRelinkRoundTrip runs the same checks against the real fixture when present.
func TestRelinkRoundTrip(t *testing.T) {
	raw, err := os.ReadFile(fixturePath(t))
	if err != nil {
		t.Fatal(err)
	}
	old, err := findEmbeddedID(raw)
	if err != nil {
		t.Fatal(err)
	}
	assertRelink(t, raw, old, "2147483648")
	assertRelink(t, raw, old, "99999")
}

// buildSyntheticFpk constructs a minimal valid foxfpk: an "a.bin" entry with no
// embedded ID and a "b.fmdl" entry whose data carries a face/real/<id> path.
func buildSyntheticFpk(t *testing.T) []byte {
	t.Helper()
	type ent struct {
		name string
		data []byte
	}
	ents := []ent{
		{"a.bin", []byte("BINDATA\x00padding-no-id")},
		{"b.fmdl", append([]byte("FMDL\x00\x00\x00\x00mesh-bytes\x00"),
			[]byte("/Assets/pes16/model/character/face/real/21665/sourceimages/face_bsm.dds\x00tail\x00")...)},
	}

	align := func(n int) int {
		if r := n % fpkAlign; r != 0 {
			return n + (fpkAlign - r)
		}
		return n
	}

	nameBlob := []byte{}
	nameOff := make([]int, len(ents))
	for i, e := range ents {
		nameOff[i] = fpkHeaderSize + len(ents)*fpkEntrySize + len(nameBlob)
		nameBlob = append(nameBlob, []byte(e.name)...)
		nameBlob = append(nameBlob, 0)
	}
	dataStart := align(fpkHeaderSize + len(ents)*fpkEntrySize + len(nameBlob))

	out := make([]byte, dataStart)
	copy(out, fpkMagic)
	copy(out[6:], []byte(" win")) // type ' ', "win"
	binary.LittleEndian.PutUint32(out[0x0C:], 4)
	binary.LittleEndian.PutUint32(out[0x20:], uint32(len(ents)))
	copy(out[dataStart-len(nameBlob):], nameBlob)

	cursor := dataStart
	for i, e := range ents {
		cursor = align(cursor)
		for len(out) < cursor {
			out = append(out, 0)
		}
		ent := fpkHeaderSize + i*fpkEntrySize
		binary.LittleEndian.PutUint64(out[ent:], uint64(nameOff[i]))
		binary.LittleEndian.PutUint64(out[ent+8:], uint64(len(e.name)))
		binary.LittleEndian.PutUint64(out[ent+32:], uint64(cursor))
		binary.LittleEndian.PutUint64(out[ent+40:], uint64(len(e.data)))
		out = append(out, e.data...)
		cursor += len(e.data)
	}
	for len(out)%fpkAlign != 0 {
		out = append(out, 0)
	}
	binary.LittleEndian.PutUint32(out[0x0A:], uint32(len(out)))
	return out
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
