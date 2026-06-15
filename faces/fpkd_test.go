package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// foxfpkdStub builds the 48-byte ID-less foxfpkd dependency package that every
// FL26 face folder ships alongside face.fpk (magic "foxfpkd", tag "win0").
func foxfpkdStub() []byte {
	b := make([]byte, 48)
	copy(b, fpkdMagic)
	copy(b[7:], []byte("win0"))
	binary.LittleEndian.PutUint32(b[0x20:], 2)
	return b
}

func TestParseRejectsFoxFpkd(t *testing.T) {
	stub := foxfpkdStub()
	if !isFoxFpkd(stub) {
		t.Fatal("stub not recognized as foxfpkd")
	}
	if isFoxFpk(stub) {
		t.Error("foxfpkd misidentified as foxfpk")
	}
	if _, err := parseFpk(stub); err == nil {
		t.Error("parseFpk accepted a foxfpkd package")
	}
}

func writeWinPackage(t *testing.T, folder, name string, data []byte) {
	t.Helper()
	winDir := filepath.Join(folder, "#Win")
	if err := os.MkdirAll(winDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(winDir, name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRewriteWalksAllPackages proves the rewrite covers every #Win package that
// carries the ID (the separate-package gap), and leaves the ID-less fpkd stub
// untouched.
func TestRewriteWalksAllPackages(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "21665")
	writeWinPackage(t, folder, "face.fpk", buildSyntheticFpk(t))
	writeWinPackage(t, folder, "extra.fpk", buildSyntheticFpk(t))
	stub := foxfpkdStub()
	writeWinPackage(t, folder, "face.fpkd", stub)

	if err := rewriteFaceFolderID(folder, "21665", "2147483648"); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	for _, name := range []string{"face.fpk", "extra.fpk"} {
		raw, err := os.ReadFile(filepath.Join(folder, "#Win", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseFpk(raw); err != nil {
			t.Errorf("%s no longer a valid foxfpk: %v", name, err)
		}
		if bytes.Contains(raw, []byte("real/21665/")) {
			t.Errorf("%s: old id path still present", name)
		}
		if !bytes.Contains(raw, []byte("real/2147483648/")) {
			t.Errorf("%s: new id path absent", name)
		}
	}

	got, err := os.ReadFile(filepath.Join(folder, "#Win", "face.fpkd"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, stub) {
		t.Error("ID-less fpkd stub was modified")
	}
}

// TestRewriteFpkdEqualLength covers a (hypothetical) fpkd that does embed the
// path: an equal-length swap is a format-agnostic in-place replace.
func TestRewriteFpkdEqualLength(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "21665")
	pkg := append(foxfpkdStub(), []byte("face/real/21665/sourceimages/x.dds\x00")...)
	writeWinPackage(t, folder, "face.fpkd", pkg)

	if err := rewriteFaceFolderID(folder, "21665", "99999"); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(folder, "#Win", "face.fpkd"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("face/real/99999/")) || bytes.Contains(got, []byte("face/real/21665/")) {
		t.Error("equal-length fpkd rewrite did not swap the id")
	}
}

// TestRewriteFpkdLengthChangeRefused: a length-changing rewrite of a foxfpkd that
// carries the path is refused, not silently corrupted (no verified repack).
func TestRewriteFpkdLengthChangeRefused(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "21665")
	pkg := append(foxfpkdStub(), []byte("face/real/21665/sourceimages/x.dds\x00")...)
	writeWinPackage(t, folder, "face.fpkd", pkg)

	err := rewriteFaceFolderID(folder, "21665", "2147483648")
	if err == nil {
		t.Fatal("expected refusal for length-changing foxfpkd rewrite")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("foxfpkd")) {
		t.Errorf("error should name the foxfpkd format: %v", err)
	}
}
