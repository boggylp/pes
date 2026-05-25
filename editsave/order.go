package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Empirically located on FL26 v2.2: each team's tactics slot stores a
// 32-byte squad-order array at slot offset 0x1e4..0x203 (bytes 484..515).
// The array is a permutation of [0..31] mapping each formation role to
// the squad-list index of the player who should fill it. FL26 ships
// with the identity permutation (00 01 02 ... 1f) for every team —
// "play squad slot N in role N". When a user toggles "by ability" in
// Edit Mode the engine writes a computed permutation here; when a
// .PES2021_tactics file is imported it overwrites this region with
// whatever permutation the source save had — usually a permutation
// computed against a DIFFERENT squad, which then maps the wrong
// players into the wrong roles in the target save.
const (
	squadOrderOffset = 0x1e4
	squadOrderLen    = 32
)

func cmdResetSquadOrder(args []string) int {
	fs := flag.NewFlagSet("reset-squad-order", flag.ContinueOnError)
	toolsDir := fs.String("tools-dir", "", "directory containing decrypter21.exe and encrypter21.exe")
	outFile := fs.String("out", "", "output EDIT00000000 file")
	force := fs.Bool("force", false, "allow overwriting an existing output file")
	allowLiveName := fs.Bool("allow-live-name", false, "allow --out to be named EDIT00000000 (the live-save filename)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outFile == "" || *toolsDir == "" {
		fmt.Fprintln(os.Stderr, "usage: editsave reset-squad-order --tools-dir DIR --out FILE [--force] [--allow-live-name] INPUT")
		fs.PrintDefaults()
		return 2
	}
	input := fs.Arg(0)

	if err := runResetSquadOrder(input, *outFile, *toolsDir, *force, *allowLiveName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// runResetSquadOrder decrypts INPUT, writes the identity squad-order
// array into every team's slot, re-encrypts to outFile. Same atomic +
// safe-out + within-session conventions as apply-tactics.
func runResetSquadOrder(input, outFile, toolsDir string, force, allowLiveName bool) error {
	tools, err := resolveTools(toolsDir)
	if err != nil {
		return err
	}
	if err := ensureSafeOutFile(outFile, force, allowLiveName); err != nil {
		return err
	}

	work, err := os.MkdirTemp("", "editsave-reset-*")
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(work); rmErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove work dir %s: %v\n", work, rmErr)
		}
	}()

	decDir := filepath.Join(work, "dec")
	if err := os.Mkdir(decDir, 0o755); err != nil {
		return err
	}
	if err := runDecrypter(tools, input, decDir); err != nil {
		return fmt.Errorf("stage 1 (decrypt): %w", err)
	}
	if err := verifyDecryptedDir(decDir); err != nil {
		return fmt.Errorf("stage 1 verify: %w", err)
	}

	dataPath := filepath.Join(decDir, "data.dat")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("reading data.dat: %w", err)
	}

	slots, sectionStart, sectionEnd, err := scanTacticsSection(data)
	if err != nil {
		return fmt.Errorf("scanning tactics section: %w", err)
	}
	fmt.Fprintf(os.Stderr, "tactics section: [0x%08x..0x%08x) = %d slots\n",
		sectionStart, sectionEnd, len(slots))

	reset := resetSquadOrderInData(data, slots)
	fmt.Fprintf(os.Stderr, "squad-order arrays reset to identity for %d slots\n", reset)
	if reset == 0 {
		return errors.New("no slots reset; refusing to encrypt an unchanged save")
	}

	if err := os.WriteFile(dataPath, data, 0o644); err != nil {
		return fmt.Errorf("writing modified data.dat: %w", err)
	}
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

// resetSquadOrderInData overwrites the squad-order array at every slot
// in slots with the identity permutation [0..31]. Returns the number of
// slots modified. Slot offsets that would write past the end of data
// are skipped silently; under normal invocation this never triggers
// because scanTacticsSection only emits offsets where a full slot fits.
func resetSquadOrderInData(data []byte, slots map[uint32]int) int {
	var identity [squadOrderLen]byte
	for i := range identity {
		identity[i] = byte(i)
	}
	count := 0
	for _, off := range slots {
		end := off + squadOrderOffset + squadOrderLen
		if off < 0 || end > len(data) {
			continue
		}
		copy(data[off+squadOrderOffset:end], identity[:])
		count++
	}
	return count
}
