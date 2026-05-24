package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// cmdRoundtrip decrypts INPUT, re-encrypts the result, decrypts the
// re-encryption, and reports whether each layer matches.
//
// Two assertions matter, independently:
//  1. strict — input.EDIT00000000 SHA256 == re-encrypted SHA256.
//     If true, decrypter21 and encrypter21 are perfect inverses.
//  2. content — decrypted INPUT data.dat SHA256 == decrypted
//     re-encryption's data.dat SHA256. If true, the round-trip preserves
//     logical content even if ciphertext differs.
//
// Exit code policy: 0 iff content matches; 1 otherwise. With
// --require-strict, also require strict match for exit 0. Default
// content-only because Konami's encryption is known to embed nonces;
// scripted callers that need byte-equality opt in explicitly.
func cmdRoundtrip(args []string) int {
	fs := flag.NewFlagSet("roundtrip", flag.ContinueOnError)
	toolsDir := fs.String("tools-dir", "", "directory containing decrypter21.exe and encrypter21.exe")
	keepWork := fs.Bool("keep-work", false, "keep the temp working directory for inspection")
	requireStrict := fs.Bool("require-strict", false, "require strict (ciphertext-equal) match for exit 0")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: editsave roundtrip --tools-dir DIR [--keep-work] [--require-strict] INPUT")
		fs.PrintDefaults()
		return 2
	}
	return runRoundtrip(fs.Arg(0), *toolsDir, *keepWork, *requireStrict)
}

// runRoundtrip is split out so cleanup runs through a single return
// path that os.Exit cannot bypass. Keeping the temp dir is opt-in.
func runRoundtrip(input, toolsDir string, keepWork, requireStrict bool) int {
	tools, err := resolveTools(toolsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	work, err := os.MkdirTemp("", "editsave-rt-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		return 1
	}
	defer func() {
		if keepWork {
			fmt.Fprintf(os.Stderr, "kept work dir: %s\n", work)
			return
		}
		if err := os.RemoveAll(work); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove work dir %s: %v\n", work, err)
		}
	}()

	decA := filepath.Join(work, "dec-a")
	if err := os.Mkdir(decA, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := runDecrypter(tools, input, decA); err != nil {
		fmt.Fprintf(os.Stderr, "stage 1 (decrypt input): %v\n", err)
		return 1
	}
	if err := verifyDecryptedDir(decA); err != nil {
		fmt.Fprintf(os.Stderr, "stage 1 verify: %v\n", err)
		return 1
	}

	reEnc := filepath.Join(work, "EDIT00000000")
	if err := runEncrypter(tools, decA, reEnc); err != nil {
		fmt.Fprintf(os.Stderr, "stage 2 (re-encrypt): %v\n", err)
		return 1
	}

	decB := filepath.Join(work, "dec-b")
	if err := os.Mkdir(decB, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := runDecrypter(tools, reEnc, decB); err != nil {
		fmt.Fprintf(os.Stderr, "stage 3 (decrypt re-encryption): %v\n", err)
		return 1
	}
	if err := verifyDecryptedDir(decB); err != nil {
		fmt.Fprintf(os.Stderr, "stage 3 verify: %v\n", err)
		return 1
	}

	encHashA, err := sha256File(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	encHashB, err := sha256File(reEnc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	dataHashA, err := sha256File(filepath.Join(decA, "data.dat"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	dataHashB, err := sha256File(filepath.Join(decB, "data.dat"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	strict := bytes.Equal(encHashA, encHashB)
	content := bytes.Equal(dataHashA, dataHashB)

	fmt.Fprintf(os.Stderr, "input:        %s  (%s)\n", input, hex.EncodeToString(encHashA))
	fmt.Fprintf(os.Stderr, "re-encrypted: %s  (%s)\n", reEnc, hex.EncodeToString(encHashB))
	fmt.Fprintf(os.Stderr, "decrypt(input)  data.dat sha256: %s\n", hex.EncodeToString(dataHashA))
	fmt.Fprintf(os.Stderr, "decrypt(re-enc) data.dat sha256: %s\n", hex.EncodeToString(dataHashB))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "strict (encrypted bytes identical):  %s\n", verdict(strict))
	fmt.Fprintf(os.Stderr, "content (data.dat identical):        %s\n", verdict(content))

	if !content {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "content FAIL: round-trip mutated the decrypted payload. Do not reuse this save.")
		return 1
	}
	if requireStrict && !strict {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "--require-strict was set but ciphertext drifted between runs.")
		return 1
	}
	return 0
}

func sha256File(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return h.Sum(nil), nil
}

func verdict(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
