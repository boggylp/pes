package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveToolsMissingDir(t *testing.T) {
	if _, err := resolveTools(""); err == nil {
		t.Fatal("expected error for empty tools-dir")
	}
}

func TestResolveToolsMissingFiles(t *testing.T) {
	if _, err := resolveTools(t.TempDir()); err == nil {
		t.Fatal("expected error when both tools absent")
	}
}

func TestResolveToolsRejectsEmptyExe(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"decrypter21.exe", "encrypter21.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := resolveTools(dir); err == nil {
		t.Fatal("expected error for zero-byte tool")
	}
}

func TestResolveToolsRejectsNonExeSuffix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "decrypter21"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "encrypter21"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveTools(dir); err == nil {
		t.Fatal("expected error for missing .exe suffix")
	}
}

func TestResolveToolsOK(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"decrypter21.exe", "encrypter21.exe"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("stub"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := resolveTools(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got.decrypter) || !filepath.IsAbs(got.encrypter) {
		t.Fatalf("paths should be absolute: %+v", got)
	}
}

func TestVerifyDecryptedDirComplete(t *testing.T) {
	dir := t.TempDir()
	for _, name := range decryptedSaveFiles {
		body := []byte("x")
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := verifyDecryptedDir(dir); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
}

func TestVerifyDecryptedDirRejectsEmptyDataDat(t *testing.T) {
	dir := t.TempDir()
	for _, name := range decryptedSaveFiles {
		body := []byte("x")
		if name == "data.dat" {
			body = nil
		}
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := verifyDecryptedDir(dir)
	if err == nil || !strings.Contains(err.Error(), "data.dat") {
		t.Fatalf("expected data.dat-specific error, got %v", err)
	}
}

func TestVerifyDecryptedDirMissingOne(t *testing.T) {
	dir := t.TempDir()
	for _, name := range decryptedSaveFiles[:len(decryptedSaveFiles)-1] {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := verifyDecryptedDir(dir)
	missing := decryptedSaveFiles[len(decryptedSaveFiles)-1]
	if err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("expected error mentioning %q, got %v", missing, err)
	}
}

func TestVerdict(t *testing.T) {
	if verdict(true) != "PASS" || verdict(false) != "FAIL" {
		t.Fatal("verdict wrong")
	}
}

func TestEnsureSafeOutFileRejectsExisting(t *testing.T) {
	dir := t.TempDir()
	// Use a name that is NOT the canonical live-save name; --allow-live-name
	// is exercised by its own dedicated test.
	f := filepath.Join(dir, "out.EDIT")
	if err := os.WriteFile(f, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureSafeOutFile(f, false, false); err == nil {
		t.Fatal("expected refusal when target exists and --force is off")
	}
	if err := ensureSafeOutFile(f, true, false); err != nil {
		t.Fatalf("--force should allow overwrite, got %v", err)
	}
}

func TestEnsureSafeOutFileRefusesLiveSaveName(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "EDIT00000000")
	// File doesn't even exist yet — refusal is solely on the basename.
	err := ensureSafeOutFile(f, true, false)
	if err == nil || !strings.Contains(err.Error(), "live-name") {
		t.Fatalf("expected live-name refusal even with --force, got %v", err)
	}
	if err := ensureSafeOutFile(f, true, true); err != nil {
		t.Fatalf("--allow-live-name should permit, got %v", err)
	}
}

func TestEnsureEmptyOutDirRejectsNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "leftover"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureEmptyOutDir(dir, false); err == nil {
		t.Fatal("expected error for non-empty dir without --force")
	}
	if err := ensureEmptyOutDir(dir, true); err != nil {
		t.Fatalf("--force should allow non-empty, got %v", err)
	}
}

// stubTools builds two no-op Go programs and parks them under a path
// containing spaces, then exercises runDecrypter/runEncrypter so we
// catch quoting regressions on real filesystems.
func TestRunDecrypterRunEncrypterWithSpacesInPath(t *testing.T) {
	if testing.Short() {
		t.Skip("builds child binaries")
	}
	if runtime.GOOS != "windows" {
		// The wrapped tools are .exe; on non-Windows we still build .exe
		// names but execution may behave oddly. Skip rather than chase
		// portability that doesn't exist in production.
		t.Skip("Windows-only smoke")
	}
	if _, err := exec.LookPath("go"); err != nil {
		// CI sandboxes / restricted shells may not have `go` on PATH;
		// the stub builder shells out to `go build`. Skip cleanly.
		t.Skip("`go` not on PATH; cannot build stub binaries")
	}

	root := t.TempDir()
	spaced := filepath.Join(root, "with spaces")
	if err := os.Mkdir(spaced, 0o755); err != nil {
		t.Fatal(err)
	}

	// Build stub decrypter: writes one zero byte to argv[2] / "data.dat"
	// and creates the other 5 files non-empty so verify passes.
	// Then a stub encrypter that just writes 1 byte to argv[2].
	if err := buildStub(t, spaced, "decrypter21.exe", decrypterStubSource); err != nil {
		t.Fatal(err)
	}
	if err := buildStub(t, spaced, "encrypter21.exe", encrypterStubSource); err != nil {
		t.Fatal(err)
	}

	tools, err := resolveTools(spaced)
	if err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(root, "out with space")
	if err := os.Mkdir(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dummyIn := filepath.Join(root, "input EDIT")
	if err := os.WriteFile(dummyIn, bytes.Repeat([]byte{0xAB}, 64), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runDecrypter(tools, dummyIn, outDir); err != nil {
		t.Fatalf("decrypter stub failed: %v", err)
	}
	if err := verifyDecryptedDir(outDir); err != nil {
		t.Fatalf("verify after stub decrypt: %v", err)
	}

	outFile := filepath.Join(root, "re encrypted EDIT")
	if err := runEncrypter(tools, outDir, outFile); err != nil {
		t.Fatalf("encrypter stub failed: %v", err)
	}
	if info, err := os.Stat(outFile); err != nil || info.Size() == 0 {
		t.Fatalf("encrypter stub did not produce expected output: stat=%v size=%v", err, info)
	}
}

func buildStub(t *testing.T, dir, name, source string) error {
	t.Helper()
	src := filepath.Join(t.TempDir(), "stub.go")
	if err := os.WriteFile(src, []byte(source), 0o644); err != nil {
		return err
	}
	out := filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", out, src)
	if buf, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("building stub %s: %w (%s)", name, err, buf)
	}
	return nil
}

const decrypterStubSource = `package main

import (
	"os"
	"path/filepath"
)

// Stub decrypter: writes the 6 files editsave expects. data.dat gets
// 1 byte so verifyDecryptedDir's non-empty check passes.
func main() {
	if len(os.Args) < 3 {
		os.Exit(2)
	}
	outDir := os.Args[2]
	files := map[string]int{
		"data.dat":         1,
		"description.dat":  1,
		"encryptHeader.dat": 1,
		"header.dat":       1,
		"logo.png":         1,
		"version.txt":      1,
	}
	for name, size := range files {
		buf := make([]byte, size)
		if err := os.WriteFile(filepath.Join(outDir, name), buf, 0o644); err != nil {
			os.Exit(1)
		}
	}
}
`

const encrypterStubSource = `package main

import "os"

// Stub encrypter: writes 1 byte to argv[2] regardless of input.
func main() {
	if len(os.Args) < 3 {
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[2], []byte{0x00}, 0o644); err != nil {
		os.Exit(1)
	}
}
`
