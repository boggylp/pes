package main

import (
	"flag"
	"fmt"
	"os"
)

func cmdDecrypt(args []string) int {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	toolsDir := fs.String("tools-dir", "", "directory containing decrypter21.exe and encrypter21.exe")
	outDir := fs.String("out", "", "output directory (must not contain a pre-existing decrypted save)")
	force := fs.Bool("force", false, "allow writing into a non-empty output directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: editsave decrypt --tools-dir DIR --out DIR [--force] INPUT")
		fs.PrintDefaults()
		return 2
	}
	input := fs.Arg(0)

	tools, err := resolveTools(*toolsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := ensureEmptyOutDir(*outDir, *force); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := runDecrypter(tools, input, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := verifyDecryptedDir(*outDir); err != nil {
		fmt.Fprintf(os.Stderr, "decrypter ran but output is bad: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "decrypted to %s\n", *outDir)
	return 0
}

// ensureEmptyOutDir creates outDir if absent. If it exists and is
// non-empty, fails unless force is true. Refusing the overwrite prevents
// mixing stale files from a previous decrypt with the new run.
func ensureEmptyOutDir(outDir string, force bool) error {
	info, err := os.Stat(outDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(outDir, 0o755)
		}
		return fmt.Errorf("stat %s: %w", outDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("--out exists and is not a directory: %s", outDir)
	}
	if force {
		return nil
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("reading %s: %w", outDir, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("--out directory is not empty: %s (pass --force to overwrite)", outDir)
	}
	return nil
}
