package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFPKInternalID(t *testing.T) {
	// Create a temp file with an embedded path
	dir := t.TempDir()
	fpkPath := filepath.Join(dir, "face.fpk")

	content := []byte("foxfpk\x00win\x00\x00\x00/Assets/pes16/model/character/face/real/65342/sourceimages/\x00\x00")
	if err := os.WriteFile(fpkPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	got := extractFPKInternalID(fpkPath)
	if got != "65342" {
		t.Errorf("extractFPKInternalID() = %q, want %q", got, "65342")
	}
}

func TestExtractFPKInternalID_NoMatch(t *testing.T) {
	dir := t.TempDir()
	fpkPath := filepath.Join(dir, "face.fpk")

	content := []byte("foxfpk\x00win\x00\x00\x00some random binary data\x00\x00")
	if err := os.WriteFile(fpkPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	got := extractFPKInternalID(fpkPath)
	if got != "" {
		t.Errorf("extractFPKInternalID() = %q, want empty", got)
	}
}

func TestToASCII(t *testing.T) {
	input := []byte("hello\x00world\xff!")
	got := toASCII(input)
	if got != "hello world !" {
		t.Errorf("toASCII() = %q, want %q", got, "hello world !")
	}
}
