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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outFile == "" {
		fmt.Fprintln(os.Stderr, "usage: editsave encrypt --tools-dir DIR --out FILE [--force] INPUT_DIR")
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

	if err := ensureSafeOutFile(*outFile, *force); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(*outFile), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir output parent: %v\n", err)
		return 1
	}

	if err := runEncrypter(tools, inputDir, *outFile); err != nil {
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

// ensureSafeOutFile refuses to overwrite an existing file unless force
// is set. Without this guard, pointing --out at the live EDIT00000000
// would silently destroy the user's save on every run.
func ensureSafeOutFile(outFile string, force bool) error {
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
