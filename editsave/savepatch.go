package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type slotPatch func(data []byte, slots map[uint32]int) error

// patchSave keeps decrypt, patch, and encrypt inside one process because
// decrypter21 output is session-coupled and a later run cannot re-encrypt it.
func patchSave(tools toolPaths, input, outFile, workPattern string, patch slotPatch) error {
	return withWorkDir(workPattern, func(work string) error {
		decDir := filepath.Join(work, "dec")
		if err := decryptStage(tools, input, decDir, "1", "decrypt"); err != nil {
			return err
		}
		dataPath := filepath.Join(decDir, "data.dat")
		data, err := os.ReadFile(dataPath)
		if err != nil {
			return fmt.Errorf("reading data.dat: %w", err)
		}
		slots, err := loadTacticsSlots(data)
		if err != nil {
			return err
		}
		if err := patch(data, slots); err != nil {
			return err
		}
		if err := os.WriteFile(dataPath, data, 0o644); err != nil {
			return fmt.Errorf("writing modified data.dat: %w", err)
		}
		return encryptAndReport(tools, decDir, outFile)
	})
}

func withWorkDir(pattern string, fn func(work string) error) error {
	work, err := os.MkdirTemp("", pattern)
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}
	defer removeWorkDir(work)
	return fn(work)
}

func removeWorkDir(work string) {
	if err := os.RemoveAll(work); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove work dir %s: %v\n", work, err)
	}
}

func loadTacticsSlots(data []byte) (map[uint32]int, error) {
	slots, sectionStart, sectionEnd, err := scanTacticsSection(data)
	if err != nil {
		return nil, fmt.Errorf("scanning tactics section: %w", err)
	}
	fmt.Fprintf(os.Stderr, "tactics section: [0x%08x..0x%08x) = %d slots\n",
		sectionStart, sectionEnd, len(slots))
	return slots, nil
}

func encryptAndReport(tools toolPaths, decDir, outFile string) error {
	if err := encrypterToFileAtomic(tools, decDir, outFile); err != nil {
		return fmt.Errorf("stage 2 (encrypt): %w", err)
	}
	info, err := os.Stat(outFile)
	if err != nil {
		return fmt.Errorf("encrypter ran but output missing: %w", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", outFile, info.Size())
	return nil
}
