package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// tacticsSlotSize is the size of one team's tactics block inside data.dat.
	// The team tactics section is a contiguous array of these slots.
	tacticsSlotSize = 628

	// tacticsTeamIDMax is the largest plausible team_id; any uint32 above
	// this in the team_id field rules out a slot candidate as a false positive.
	tacticsTeamIDMax = 100000

	// tacticsRegionMinSlots is the minimum slot count we expect the section
	// to have. FL26 ships 749 on this machine; well below that means our
	// scanner anchored on a false positive.
	tacticsRegionMinSlots = 100

	// tacticsScanStartOffset is the earliest byte we scan for the tactics
	// section. The section lives near the tail of data.dat (~10.5 MB on
	// the current FL26 v2.2 save); scanning from byte 0 wastes time and
	// risks anchoring on a coincidental match earlier in the file.
	tacticsScanStartOffset = 10_000_000
)

// tacticsHeaderMagic is the two-byte fixed prefix at offsets 4..5 of
// every tactics slot. Byte 6 varies across teams (it encodes the
// formation type or tactic variant), so it is NOT part of the magic.
// Examples observed in shipped .PES2021_tactics files:
//   - 5-ENGLAND:        00 01 01 ...
//   - 13-CZECH REPUBLIC: 00 01 03 ...
//   - 24-CROATIA:       00 01 01 ...
var tacticsHeaderMagic = [2]byte{0x00, 0x01}

// tacticsScanExtraByte is the third byte we require during data.dat
// scanning. The default FL26 save has byte 6 == 0x01 for every team's
// tactics slot (FL26 ships a uniform default formation/variant). A
// 2-byte-only scan picks up false positives elsewhere in data.dat, so
// the scan uses 3 bytes while the source-file magic stays at 2.
// If a future scan needs to handle a save where users imported tactics
// directly into data.dat (mixing variant bytes), drop this constant and
// add a longest-chain heuristic instead.
const tacticsScanExtraByte = 0x01

func cmdApplyTactics(args []string) int {
	fs := flag.NewFlagSet("apply-tactics", flag.ContinueOnError)
	toolsDir := fs.String("tools-dir", "", "directory containing decrypter21.exe and encrypter21.exe")
	tacticsDir := fs.String("tactics-dir", "", "directory of .PES2021_tactics files (e.g. importer/imports)")
	idListPath := fs.String("id-list", "", "id_list.txt mapping team_id to tactics filename")
	outFile := fs.String("out", "", "output EDIT00000000 file")
	force := fs.Bool("force", false, "allow overwriting an existing output file")
	allowLiveName := fs.Bool("allow-live-name", false, "allow --out to be named EDIT00000000 (the live-save filename)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outFile == "" || *toolsDir == "" || *tacticsDir == "" || *idListPath == "" {
		fmt.Fprintln(os.Stderr, "usage: editsave apply-tactics --tools-dir DIR --tactics-dir DIR --id-list FILE --out FILE [--force] [--allow-live-name] INPUT")
		fs.PrintDefaults()
		return 2
	}
	input := fs.Arg(0)

	if err := runApplyTactics(input, *outFile, *toolsDir, *tacticsDir, *idListPath, *force, *allowLiveName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// runApplyTactics performs decrypt → patch tactics → encrypt entirely
// within one process. The decrypted intermediate is held in a temp dir
// that is unconditionally removed before return; never let the temp dir
// outlive the process because decrypter21 output is session-coupled and
// can't be re-encrypted by a later run.
func runApplyTactics(input, outFile, toolsDir, tacticsDir, idListPath string, force, allowLiveName bool) error {
	tools, err := resolveTools(toolsDir)
	if err != nil {
		return err
	}
	if err := ensureSafeOutFile(outFile, force, allowLiveName); err != nil {
		return err
	}
	entries, err := parseIDList(idListPath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", idListPath, err)
	}
	if len(entries) == 0 {
		return errors.New("id-list contains no entries")
	}

	work, err := os.MkdirTemp("", "editsave-apply-*")
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

	applied, skipped, missing, err := applyTacticsToData(data, slots, entries, tacticsDir)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "applied: %d, skipped (team_id absent from save): %d, missing tactics files: %d\n",
		applied, skipped, missing)
	if applied == 0 {
		return fmt.Errorf("no tactics applied (skipped=%d, missing=%d); refusing to encrypt an unchanged save", skipped, missing)
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

// parseIDList reads zlac's id_list.txt format:
//
//	team_id, tactics_filename [# optional trailing comment]
//
// Lines starting with `#` and blank lines are ignored. Inline comments
// after `#` are stripped.
type tacticsListEntry struct {
	teamID   uint32
	fileName string
}

func parseIDList(path string) ([]tacticsListEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []tacticsListEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		// Skip whole-line comments and blanks BEFORE splitting on
		// comma, so the line-level `#` strip never bites into a
		// filename that contains a `#`.
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected `team_id, filename`, got %q", raw)
		}
		tid64, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("bad team_id in %q: %w", raw, err)
		}
		fname := parts[1]
		// Strip an inline trailing comment from the filename field,
		// but only when the `#` is preceded by whitespace; that lets a
		// filename legitimately containing `#` survive.
		if i := indexCommentDelim(fname); i >= 0 {
			fname = fname[:i]
		}
		fname = strings.TrimSpace(fname)
		if fname == "" {
			return nil, fmt.Errorf("empty filename in %q", raw)
		}
		entries = append(entries, tacticsListEntry{teamID: uint32(tid64), fileName: fname})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// indexCommentDelim returns the index of the first `#` that is
// preceded by whitespace (treating it as a comment delimiter), or -1 if
// none. This preserves `#` inside an unquoted filename — e.g.
// `5, foo#bar.PES2021_tactics` keeps the whole filename, but
// `5, foo.PES2021_tactics # comment` strips the comment.
func indexCommentDelim(s string) int {
	for i, r := range s {
		if r == '#' && (i == 0 || isASCIISpace(s[i-1])) {
			return i
		}
	}
	return -1
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// tacticsRunGapTolerance is the maximum number of consecutive
// non-conforming slots we tolerate when walking out from the anchor.
// After Klashman tactics are applied byte 6 (formation variant) varies
// per slot, so the strict 3-byte anchor scan stops at the first variant;
// we then walk in 628-byte strides with the looser 2-byte magic and
// only abandon the walk after this many failures in a row. Tuned for
// FL26 v2.2 plus typical tactics packs.
const tacticsRunGapTolerance = 5

// scanTacticsSection finds the team-tactics array inside data and
// returns a team_id → offset map plus the section bounds.
//
// Algorithm:
//  1. Anchor: scan forward from tacticsScanStartOffset for the first
//     position that satisfies the strict 3-byte slot check. The strict
//     check pins us inside the real tactics section because the byte-6
//     variant byte is 0x01 in the FL26 default state.
//  2. Walk back and forward in tacticsSlotSize strides using the looser
//     2-byte slot check, which tolerates per-team variant-byte values
//     that appear once tactics have been imported. Stop a direction
//     only after tacticsRunGapTolerance consecutive failures.
func scanTacticsSection(data []byte) (map[uint32]int, int, int, error) {
	anchor := -1
	for off := tacticsScanStartOffset; off+tacticsSlotSize <= len(data); off++ {
		if isTacticsSlot(data, off) {
			anchor = off
			break
		}
	}
	if anchor < 0 {
		return nil, 0, 0, errors.New("no tactics slot found in scan window")
	}

	start := walkSlots(data, anchor, -tacticsSlotSize)
	end := walkSlots(data, anchor, +tacticsSlotSize) + tacticsSlotSize

	nSlots := (end - start) / tacticsSlotSize
	if nSlots < tacticsRegionMinSlots {
		return nil, 0, 0, fmt.Errorf("tactics section too small (%d slots); scan likely anchored on a false positive", nSlots)
	}

	slots := make(map[uint32]int, nSlots)
	for off := start; off < end; off += tacticsSlotSize {
		if !isLooseSlot(data, off) {
			continue
		}
		tid := binary.LittleEndian.Uint32(data[off : off+4])
		if existing, dup := slots[tid]; dup {
			return nil, 0, 0, fmt.Errorf("duplicate team_id %d at offsets 0x%x and 0x%x", tid, existing, off)
		}
		slots[tid] = off
	}
	return slots, start, end, nil
}

// walkSlots steps from anchor by step (positive or negative, must be a
// multiple of tacticsSlotSize). Returns the last position that looked
// like a slot under the loose check, allowing up to
// tacticsRunGapTolerance consecutive failures along the way.
//
// Returns lastValid even when the walk crossed gaps; the section bounds
// must always sit on a valid slot, not on the tail of a failed run.
func walkSlots(data []byte, anchor, step int) int {
	lastValid := anchor
	cur := anchor
	gaps := 0
	for {
		cur += step
		if cur < 0 || cur+tacticsSlotSize > len(data) {
			return lastValid
		}
		if isLooseSlot(data, cur) {
			lastValid = cur
			gaps = 0
			continue
		}
		gaps++
		if gaps > tacticsRunGapTolerance {
			return lastValid
		}
	}
}

// isLooseSlot is the slot check used while walking the section bounds.
// It checks only byte 4 == 0x00 and a plausible team_id; bytes 5 and 6
// vary per team once tactics are applied (e.g. England keeps 0x01,0x01;
// Czech Republic becomes 0x01,0x03; Real Madrid becomes 0x04,0x01).
// The probability of a false positive at a random 628-byte stride
// position is roughly 1/256 (byte 4 = 0x00) × 1/(2^32/100_000) so this
// is still a strong filter when combined with stride alignment.
func isLooseSlot(data []byte, off int) bool {
	if off < 0 || off+tacticsSlotSize > len(data) {
		return false
	}
	if data[off+4] != 0x00 {
		return false
	}
	tid := binary.LittleEndian.Uint32(data[off : off+4])
	return tid >= 1 && tid <= tacticsTeamIDMax
}

func isTacticsSlot(data []byte, off int) bool {
	if off < 0 || off+tacticsSlotSize > len(data) {
		return false
	}
	if data[off+4] != tacticsHeaderMagic[0] ||
		data[off+5] != tacticsHeaderMagic[1] ||
		data[off+6] != tacticsScanExtraByte {
		return false
	}
	tid := binary.LittleEndian.Uint32(data[off : off+4])
	return tid >= 1 && tid <= tacticsTeamIDMax
}

// applyTacticsToData mutates data in-place. For each id_list entry, it
// reads the .PES2021_tactics file and writes its tactics-slot-sized
// payload at the offset for that team_id. Returns counts of (applied,
// skipped-because-team-absent-from-save, missing-tactics-file).
//
// zlac's id_list format intentionally allows the source file's internal
// team_id to differ from the target — e.g. `1099, 43-HONDURAS.PES2021_tactics`
// reuses Honduras tactics exported from a game where Honduras was id 43
// for FL26's Honduras at id 1099. To keep the on-disk slot header
// consistent with its own slot index, the first 4 bytes are rewritten
// with the target team_id before the copy lands.
func applyTacticsToData(data []byte, slots map[uint32]int, entries []tacticsListEntry, tacticsDir string) (int, int, int, error) {
	var applied, skipped, missing int
	// Deterministic iteration for stable output / easier diff.
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].teamID < entries[j].teamID })
	for _, e := range entries {
		off, ok := slots[e.teamID]
		if !ok {
			skipped++
			fmt.Fprintf(os.Stderr, "  skip team_id=%d: not present in save\n", e.teamID)
			continue
		}
		path := filepath.Join(tacticsDir, e.fileName)
		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				missing++
				fmt.Fprintf(os.Stderr, "  missing tactics file: %s\n", e.fileName)
				continue
			}
			return applied, skipped, missing, fmt.Errorf("reading %s: %w", path, err)
		}
		if len(body) != tacticsSlotSize {
			return applied, skipped, missing, fmt.Errorf("%s: expected %d bytes, got %d", e.fileName, tacticsSlotSize, len(body))
		}
		// Defensive: source .PES2021_tactics files in the wild vary at
		// bytes 5..6 (formation variant / tactic count). The only
		// universal invariants are size == 628 (checked above) and a
		// plausible team_id at bytes 0..3. Reject obvious garbage.
		fileTID := binary.LittleEndian.Uint32(body[:4])
		if fileTID == 0 || fileTID > tacticsTeamIDMax {
			return applied, skipped, missing, fmt.Errorf("%s: implausible team_id=%d in file header", e.fileName, fileTID)
		}
		if body[4] != 0x00 {
			return applied, skipped, missing, fmt.Errorf("%s: expected byte 4 == 0x00, got 0x%02x", e.fileName, body[4])
		}
		// Stamp the target team_id over the file's internal one so the
		// slot header stays consistent with its position in data.dat.
		patched := make([]byte, tacticsSlotSize)
		copy(patched, body)
		binary.LittleEndian.PutUint32(patched[:4], e.teamID)
		copy(data[off:off+tacticsSlotSize], patched)
		applied++
	}
	return applied, skipped, missing, nil
}
