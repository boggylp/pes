package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func cmdEncrypt(args []string) int {
	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	toolsDir := fs.String("tools-dir", "", "directory containing decrypter21.exe and encrypter21.exe")
	outFile := fs.String("out", "", "output EDIT00000000 file path")
	force := fs.Bool("force", false, "allow overwriting an existing output file")
	allowLiveName := fs.Bool("allow-live-name", false, "allow --out to be named EDIT00000000 (the live-save filename)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outFile == "" {
		fmt.Fprintln(os.Stderr, "usage: editsave encrypt --tools-dir DIR --out FILE [--force] [--allow-live-name] INPUT_DIR")
		fs.PrintDefaults()
		return 2
	}
	inputDir := fs.Arg(0)

	tools, err := resolveTools(*toolsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := verifyDecryptedDir(inputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := ensureSafeOutFile(*outFile, *force, *allowLiveName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(*outFile), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir output parent: %v\n", err)
		return 1
	}

	if err := encrypterToFileAtomic(tools, inputDir, *outFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	info, err := os.Stat(*outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encrypter ran but output is missing: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "encrypted to %s (%d bytes)\n", *outFile, info.Size())
	return 0
}

// encrypterToFileAtomic invokes the encrypter into a sibling temp file,
// then os.Renames it onto outFile on success. If the encrypter crashes
// mid-write the temp file is removed and outFile is left untouched, so
// a partial corrupt EDIT save can never land at --out.
func encrypterToFileAtomic(tools toolPaths, inputDir, outFile string) error {
	tmpFile := outFile + ".staging"
	if err := os.Remove(tmpFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing stale staging file %s: %w", tmpFile, err)
	}
	if err := runEncrypter(tools, inputDir, tmpFile); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	if err := os.Rename(tmpFile, outFile); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("renaming staging file to final output: %w", err)
	}
	return nil
}

// ensureSafeOutFile refuses to overwrite an existing file unless force
// is set. It additionally refuses --out paths whose basename matches
// the canonical live-save filename EDIT00000000 unless allowLiveName
// is also set; --force alone is not enough to overwrite a live save.
// This is a defense-in-depth check: a user typing `--force --out <live>`
// would otherwise destroy the live save on every run.
func ensureSafeOutFile(outFile string, force, allowLiveName bool) error {
	if filepath.Base(outFile) == liveSaveBaseName && !allowLiveName {
		return fmt.Errorf("--out basename is %q (looks like a live save) and --allow-live-name was not set", liveSaveBaseName)
	}
	info, err := os.Stat(outFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", outFile, err)
	}
	if info.IsDir() {
		return fmt.Errorf("--out points at a directory, not a file: %s", outFile)
	}
	if force {
		return nil
	}
	return fmt.Errorf("--out already exists: %s (pass --force to overwrite)", outFile)
}

// liveSaveBaseName is the canonical name of the FL26 / PES 2021 EDIT
// save file as it lives on disk inside the game's save directories.
const liveSaveBaseName = "EDIT00000000"
