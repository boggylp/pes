package main

import (
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

// TestFindEmbeddedID checks the texture-path ID is read out of packed data.
func TestFindEmbeddedID(t *testing.T) {
	raw := buildSyntheticFpk(t)
	old, err := findEmbeddedID(raw)
	if err != nil {
		t.Fatalf("findEmbeddedID on synthetic: %v", err)
	}
	if old != "21665" {
		t.Fatalf("synthetic embedded id = %q, want 21665", old)
	}
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
		// "mesh21665bytes" carries the bare ID outside the path; an anchored
		// rewrite must leave it untouched.
		{"b.fmdl", append([]byte("FMDL\x00\x00\x00\x00mesh21665bytes\x00"),
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
