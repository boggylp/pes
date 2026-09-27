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
	defer cleanupRoundtripWork(work, keepWork)

	hashes, err := runRoundtripStages(tools, input, work)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return reportRoundtrip(hashes, requireStrict)
}

func cleanupRoundtripWork(work string, keepWork bool) {
	if keepWork {
		fmt.Fprintf(os.Stderr, "kept work dir: %s\n", work)
		return
	}
	removeWorkDir(work)
}

type roundtripHashes struct {
	input, reEnc string
	encA, encB   []byte
	dataA, dataB []byte
}

func runRoundtripStages(tools toolPaths, input, work string) (roundtripHashes, error) {
	decA := filepath.Join(work, "dec-a")
	if err := decryptStage(tools, input, decA, "1", "decrypt input"); err != nil {
		return roundtripHashes{}, err
	}
	reEnc := filepath.Join(work, "EDIT00000000")
	if err := runEncrypter(tools, decA, reEnc); err != nil {
		return roundtripHashes{}, fmt.Errorf("stage 2 (re-encrypt): %w", err)
	}
	decB := filepath.Join(work, "dec-b")
	if err := decryptStage(tools, reEnc, decB, "3", "decrypt re-encryption"); err != nil {
		return roundtripHashes{}, err
	}

	paths := []string{input, reEnc, filepath.Join(decA, "data.dat"), filepath.Join(decB, "data.dat")}
	sums := make([][]byte, len(paths))
	for i, p := range paths {
		sum, err := sha256File(p)
		if err != nil {
			return roundtripHashes{}, err
		}
		sums[i] = sum
	}
	return roundtripHashes{input: input, reEnc: reEnc, encA: sums[0], encB: sums[1], dataA: sums[2], dataB: sums[3]}, nil
}

func decryptStage(tools toolPaths, in, dir, stage, label string) error {
	if err := os.Mkdir(dir, 0o755); err != nil {
		return err
	}
	if err := runDecrypter(tools, in, dir); err != nil {
		return fmt.Errorf("stage %s (%s): %w", stage, label, err)
	}
	if err := verifyDecryptedDir(dir); err != nil {
		return fmt.Errorf("stage %s verify: %w", stage, err)
	}
	return nil
}

func reportRoundtrip(h roundtripHashes, requireStrict bool) int {
	strict := bytes.Equal(h.encA, h.encB)
	content := bytes.Equal(h.dataA, h.dataB)

	fmt.Fprintf(os.Stderr, "input:        %s  (%s)\n", h.input, hex.EncodeToString(h.encA))
	fmt.Fprintf(os.Stderr, "re-encrypted: %s  (%s)\n", h.reEnc, hex.EncodeToString(h.encB))
	fmt.Fprintf(os.Stderr, "decrypt(input)  data.dat sha256: %s\n", hex.EncodeToString(h.dataA))
	fmt.Fprintf(os.Stderr, "decrypt(re-enc) data.dat sha256: %s\n", hex.EncodeToString(h.dataB))
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
	defer func() { _ = f.Close() }()
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
