package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// toolPaths holds resolved absolute paths to decrypter21.exe and
// encrypter21.exe. Both must exist as regular files with a .exe suffix
// and non-zero size before any invocation.
type toolPaths struct {
	decrypter string
	encrypter string
}

func resolveTools(toolsDir string) (toolPaths, error) {
	if toolsDir == "" {
		return toolPaths{}, errors.New("--tools-dir is required (directory containing decrypter21.exe and encrypter21.exe)")
	}
	absDir, err := filepath.Abs(toolsDir)
	if err != nil {
		return toolPaths{}, fmt.Errorf("resolving --tools-dir to absolute path: %w", err)
	}
	dec := filepath.Join(absDir, "decrypter21.exe")
	enc := filepath.Join(absDir, "encrypter21.exe")
	if err := checkToolFile(dec); err != nil {
		return toolPaths{}, err
	}
	if err := checkToolFile(enc); err != nil {
		return toolPaths{}, err
	}
	return toolPaths{decrypter: dec, encrypter: enc}, nil
}

func checkToolFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("tool not found: %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file, got a directory: %s", path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".exe") {
		return fmt.Errorf("expected .exe extension on %s", path)
	}
	if info.Size() == 0 {
		return fmt.Errorf("tool is empty: %s", path)
	}
	return nil
}

// runDecrypter invokes decrypter21.exe <inputFile> <outputDir> with
// stderr/stdout streamed to the parent process. Paths are normalized
// to absolute form because decrypter21.exe is a 2020-era 32-bit binary
// that mishandles relative paths and trailing separators on Windows.
func runDecrypter(tools toolPaths, inputFile, outputDir string) error {
	absIn, err := filepath.Abs(inputFile)
	if err != nil {
		return fmt.Errorf("resolving input file: %w", err)
	}
	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}
	cmd := exec.Command(tools.decrypter, absIn, absOut)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("decrypter21.exe %s %s: %w", absIn, absOut, err)
	}
	return nil
}

// runEncrypter invokes encrypter21.exe <inputDir> <outputFile> with
// stderr/stdout streamed to the parent process. Paths are absolutized
// for the same Windows-quirk reasons as decrypter.
func runEncrypter(tools toolPaths, inputDir, outputFile string) error {
	absIn, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("resolving input dir: %w", err)
	}
	absOut, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("resolving output file: %w", err)
	}
	cmd := exec.Command(tools.encrypter, absIn, absOut)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encrypter21.exe %s %s: %w", absIn, absOut, err)
	}
	return nil
}

// decryptedSaveFiles is the set of files decrypter21.exe produces and
// encrypter21.exe consumes. data.dat is the payload; the others are
// header and metadata.
var decryptedSaveFiles = []string{
	"data.dat",
	"description.dat",
	"encryptHeader.dat",
	"header.dat",
	"logo.png",
	"version.txt",
}

// verifyDecryptedDir checks that dir contains every file decrypter21.exe
// produces AND that data.dat is non-empty. A zero-byte data.dat from a
// failed decrypt would otherwise re-encrypt into a destroyed save.
// Returns nil on success or an error naming what is wrong.
func verifyDecryptedDir(dir string) error {
	var missing []string
	for _, name := range decryptedSaveFiles {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, name)
				continue
			}
			return fmt.Errorf("stat %s: %w", p, err)
		}
		if name == "data.dat" && info.Size() == 0 {
			return fmt.Errorf("data.dat in %s is zero bytes; decrypt likely failed", dir)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("decrypted directory is incomplete; missing: %v", missing)
	}
	return nil
}
