package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseIDList(t *testing.T) {
	// Use CRLF on the second line to confirm Windows-formatted files
	// also parse cleanly.
	input := "# header comment\r\n# another comment\n\n" +
		"3, 3-Scotland.PES2021_tactics\r\n" +
		"4, 4-Wales.PES2021_tactics # trailing comment\n" +
		"5, 5-ENGLAND.PES2021_tactics\n" +
		"1099, 43-HONDURAS.PES2021_tactics\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "id_list.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := parseIDList(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []tacticsListEntry{
		{teamID: 3, fileName: "3-Scotland.PES2021_tactics"},
		{teamID: 4, fileName: "4-Wales.PES2021_tactics"},
		{teamID: 5, fileName: "5-ENGLAND.PES2021_tactics"},
		{teamID: 1099, fileName: "43-HONDURAS.PES2021_tactics"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("entry %d: got %+v want %+v", i, got[i], want[i])
		}
	}
}

func TestParseIDListPreservesHashInFilename(t *testing.T) {
	// `#` not preceded by whitespace should be treated as part of the
	// filename, not a comment delimiter. This preserves filenames
	// that legitimately contain `#`.
	input := "5, foo#bar.PES2021_tactics\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "id_list.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := parseIDList(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].fileName != "foo#bar.PES2021_tactics" {
		t.Fatalf("expected single entry with hashed filename, got %+v", got)
	}
}

func TestParseIDListStripsHashAfterSpace(t *testing.T) {
	input := "5, foo.PES2021_tactics  # comment\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "id_list.txt")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := parseIDList(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].fileName != "foo.PES2021_tactics" {
		t.Fatalf("expected comment stripped, got %+v", got)
	}
}

func TestWalkSlotsHandlesGapsBelowTolerance(t *testing.T) {
	// Build a section where slot 5 is intentionally corrupted; walking
	// should skip over it and return the position of the last truly
	// valid slot.
	good := make([][]byte, 10)
	for i := range good {
		good[i] = buildSlot(uint32(i+1), byte(i+1))
	}
	data := buildSectionData(good)
	// Corrupt slot 5 in place (relative to section start).
	off := tacticsScanStartOffset + 1000 + 4*tacticsSlotSize
	data[off+4] = 0xFF
	// Walk forward from slot 0; should still reach the last valid slot.
	anchor := tacticsScanStartOffset + 1000
	last := walkSlots(data, anchor, +tacticsSlotSize)
	wantLast := tacticsScanStartOffset + 1000 + 9*tacticsSlotSize
	if last != wantLast {
		t.Errorf("walk forward: got 0x%x want 0x%x", last, wantLast)
	}
}

func TestWalkSlotsStopsAtToleranceExceeded(t *testing.T) {
	// Section with one good slot followed by garbage; walking forward
	// must stop and return the good slot's position, not the garbage.
	one := [][]byte{buildSlot(1, 0x01)}
	data := buildSectionData(one)
	// Force several invalid slots after the anchor — beyond tolerance.
	anchor := tacticsScanStartOffset + 1000
	last := walkSlots(data, anchor, +tacticsSlotSize)
	if last != anchor {
		t.Errorf("walk should return anchor when no further valid slots in tolerance window, got 0x%x", last)
	}
}

func TestIsLooseSlot(t *testing.T) {
	good := buildSlot(5, 0x42)
	if !isLooseSlot(good, 0) {
		t.Error("good slot rejected by loose check")
	}
	bad := buildSlot(5, 0x42)
	bad[4] = 0x01
	if isLooseSlot(bad, 0) {
		t.Error("byte 4 != 0x00 accepted")
	}
	zero := buildSlot(0, 0x42)
	if isLooseSlot(zero, 0) {
		t.Error("team_id 0 accepted")
	}
	huge := buildSlot(tacticsTeamIDMax+1, 0x42)
	if isLooseSlot(huge, 0) {
		t.Error("oversized team_id accepted")
	}
}

func TestParseIDListMalformed(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantSub string
	}{
		{"missing comma", "5 5-England.PES2021_tactics", "expected"},
		{"non-numeric id", "abc, foo.PES2021_tactics", "bad team_id"},
		{"empty filename", "5, ", "empty filename"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "id.txt")
			if err := os.WriteFile(p, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := parseIDList(p)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("got %q, want substring %q", err.Error(), tc.wantSub)
			}
		})
	}
}

// buildSlot returns a tacticsSlotSize-byte block whose first 4 bytes encode
// teamID LE u32 and bytes 4..7 hold tacticsHeaderMagic + payload[0]. The
// remaining bytes are filler so that the block is unique per team.
func buildSlot(teamID uint32, fill byte) []byte {
	b := make([]byte, tacticsSlotSize)
	binary.LittleEndian.PutUint32(b[:4], teamID)
	copy(b[4:6], tacticsHeaderMagic[:])
	b[6] = 0x01 // tactic variant byte; varies in real data but not checked
	for i := 7; i < tacticsSlotSize; i++ {
		b[i] = fill
	}
	return b
}

// buildSectionData places three slots inside a 12 MB buffer at an offset
// greater than tacticsScanStartOffset so the scanner finds it.
func buildSectionData(slots [][]byte) []byte {
	const totalSize = 12_000_000
	buf := make([]byte, totalSize)
	// Fill with garbage that intentionally never satisfies isTacticsSlot.
	for i := 0; i < totalSize; i++ {
		buf[i] = 0xFF
	}
	// Place the section just past the scan start offset.
	off := tacticsScanStartOffset + 1000
	for _, s := range slots {
		copy(buf[off:], s)
		off += tacticsSlotSize
	}
	return buf
}

func TestIsTacticsSlot(t *testing.T) {
	good := buildSlot(5, 0x42)
	if !isTacticsSlot(good, 0) {
		t.Error("good slot rejected")
	}
	// bad header
	bad := buildSlot(5, 0x42)
	bad[4] = 0xFF
	if isTacticsSlot(bad, 0) {
		t.Error("bad header accepted")
	}
	// team_id zero is rejected
	zero := buildSlot(0, 0x42)
	if isTacticsSlot(zero, 0) {
		t.Error("team_id=0 accepted")
	}
	// team_id too high
	huge := buildSlot(tacticsTeamIDMax+1, 0x42)
	if isTacticsSlot(huge, 0) {
		t.Error("oversized team_id accepted")
	}
	// short buffer
	if isTacticsSlot(good[:50], 0) {
		t.Error("truncated buffer accepted")
	}
}

func TestScanTacticsSection(t *testing.T) {
	more := make([][]byte, 0, tacticsRegionMinSlots+5)
	for i := uint32(1); i <= uint32(tacticsRegionMinSlots+5); i++ {
		more = append(more, buildSlot(i, byte(i)))
	}
	data := buildSectionData(more)

	gotSlots, start, end, err := scanTacticsSection(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSlots) != len(more) {
		t.Errorf("slot count: got %d want %d", len(gotSlots), len(more))
	}
	if start != tacticsScanStartOffset+1000 {
		t.Errorf("start: got 0x%x want 0x%x", start, tacticsScanStartOffset+1000)
	}
	if end-start != len(more)*tacticsSlotSize {
		t.Errorf("end-start: got %d want %d", end-start, len(more)*tacticsSlotSize)
	}
	if gotSlots[uint32(5)] != tacticsScanStartOffset+1000+4*tacticsSlotSize {
		t.Errorf("team 5 offset wrong: 0x%x", gotSlots[uint32(5)])
	}
}

func TestScanTacticsSectionTooSmall(t *testing.T) {
	// Only a few slots — below minimum.
	short := make([][]byte, 0, 5)
	for i := uint32(1); i <= 5; i++ {
		short = append(short, buildSlot(i, byte(i)))
	}
	data := buildSectionData(short)
	if _, _, _, err := scanTacticsSection(data); err == nil {
		t.Fatal("expected error for too-small section")
	}
}

func TestScanTacticsSectionDuplicateID(t *testing.T) {
	// Two slots claim team_id 1.
	dup := make([][]byte, 0, tacticsRegionMinSlots+1)
	dup = append(dup, buildSlot(1, 0x01))
	for i := uint32(2); i <= uint32(tacticsRegionMinSlots); i++ {
		dup = append(dup, buildSlot(i, byte(i)))
	}
	dup = append(dup, buildSlot(1, 0xFE))
	data := buildSectionData(dup)
	_, _, _, err := scanTacticsSection(data)
	if err == nil || !strings.Contains(err.Error(), "duplicate team_id 1") {
		t.Fatalf("expected duplicate-team_id error, got %v", err)
	}
}

func TestApplyTacticsToData(t *testing.T) {
	slots := map[uint32]int{
		5: 1000,
		8: 1000 + tacticsSlotSize,
	}
	data := make([]byte, 1000+3*tacticsSlotSize)
	dir := t.TempDir()
	englandSlot := buildSlot(5, 0xAB)
	if err := os.WriteFile(filepath.Join(dir, "5-England.PES2021_tactics"), englandSlot, 0o644); err != nil {
		t.Fatal(err)
	}
	missingFile := buildSlot(8, 0xCD)
	_ = missingFile // referenced in id_list but not on disk

	entries := []tacticsListEntry{
		{teamID: 5, fileName: "5-England.PES2021_tactics"},
		{teamID: 8, fileName: "8-Belgium.PES2021_tactics"}, // file not present
		{teamID: 9999, fileName: "9999-Nope.PES2021_tactics"}, // team not in save
	}
	applied, skipped, missing, err := applyTacticsToData(data, slots, entries, dir)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 || skipped != 1 || missing != 1 {
		t.Errorf("got applied=%d skipped=%d missing=%d want 1/1/1", applied, skipped, missing)
	}
	// Verify England got copied to offset 1000.
	if !bytes.Equal(data[1000:1000+tacticsSlotSize], englandSlot) {
		t.Error("England tactics not written at expected offset")
	}
	// The other slot (1000+628) must remain zero.
	for i := 1000 + tacticsSlotSize; i < 1000+2*tacticsSlotSize; i++ {
		if data[i] != 0 {
			t.Errorf("byte %d should be untouched", i)
			break
		}
	}
}

func TestApplyTacticsCrossTeamIDIsAllowed(t *testing.T) {
	// zlac's id_list format intentionally lets a tactics file exported
	// from a different team_id be applied to another slot. The Go
	// importer must rewrite the header team_id so the slot stays
	// consistent.
	slots := map[uint32]int{1099: 0}
	data := make([]byte, tacticsSlotSize)
	dir := t.TempDir()
	src := buildSlot(43, 0xAB) // file says team_id=43
	if err := os.WriteFile(filepath.Join(dir, "43-Honduras.PES2021_tactics"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	applied, _, _, err := applyTacticsToData(data, slots, []tacticsListEntry{
		{teamID: 1099, fileName: "43-Honduras.PES2021_tactics"},
	}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("expected applied=1, got %d", applied)
	}
	got := binary.LittleEndian.Uint32(data[:4])
	if got != 1099 {
		t.Errorf("slot header team_id: got %d want 1099 (file's 43 must have been overwritten)", got)
	}
	// The rest of the slot should match the source bytes.
	if !bytes.Equal(data[4:tacticsSlotSize], src[4:tacticsSlotSize]) {
		t.Error("slot body should match source file beyond the team_id header")
	}
}

func TestApplyTacticsAcceptsVaryingVariantByte(t *testing.T) {
	// Real-world tactics files use different byte 6 values (formation
	// variant). 0x03 (Czech Republic) is one example. Magic check is
	// only on bytes 4..5, so the variant byte must not cause rejection.
	slots := map[uint32]int{13: 0}
	data := make([]byte, tacticsSlotSize)
	dir := t.TempDir()
	cz := buildSlot(13, 0xAA)
	cz[6] = 0x03 // formation variant differs
	if err := os.WriteFile(filepath.Join(dir, "13-Czech.PES2021_tactics"), cz, 0o644); err != nil {
		t.Fatal(err)
	}
	applied, _, _, err := applyTacticsToData(data, slots, []tacticsListEntry{
		{teamID: 13, fileName: "13-Czech.PES2021_tactics"},
	}, dir)
	if err != nil || applied != 1 {
		t.Fatalf("expected applied=1 nil err, got applied=%d err=%v", applied, err)
	}
	if data[6] != 0x03 {
		t.Errorf("variant byte not copied: data[6]=0x%02x want 0x03", data[6])
	}
}

func TestApplyTacticsRejectsBadByte4(t *testing.T) {
	slots := map[uint32]int{5: 0}
	data := make([]byte, tacticsSlotSize)
	dir := t.TempDir()
	bad := make([]byte, tacticsSlotSize)
	binary.LittleEndian.PutUint32(bad[:4], 5)
	bad[4] = 0xFF // byte 4 must be 0x00
	if err := os.WriteFile(filepath.Join(dir, "5-bad.PES2021_tactics"), bad, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := applyTacticsToData(data, slots, []tacticsListEntry{
		{teamID: 5, fileName: "5-bad.PES2021_tactics"},
	}, dir)
	if err == nil || !strings.Contains(err.Error(), "byte 4") {
		t.Fatalf("expected byte-4 error, got %v", err)
	}
}

func TestApplyTacticsRejectsImplausibleTeamID(t *testing.T) {
	slots := map[uint32]int{5: 0}
	data := make([]byte, tacticsSlotSize)
	dir := t.TempDir()
	bad := make([]byte, tacticsSlotSize)
	binary.LittleEndian.PutUint32(bad[:4], 999999999) // way above tacticsTeamIDMax
	bad[4] = 0x00
	bad[5] = 0x01
	if err := os.WriteFile(filepath.Join(dir, "5-bad.PES2021_tactics"), bad, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := applyTacticsToData(data, slots, []tacticsListEntry{
		{teamID: 5, fileName: "5-bad.PES2021_tactics"},
	}, dir)
	if err == nil || !strings.Contains(err.Error(), "implausible") {
		t.Fatalf("expected implausible-team_id error, got %v", err)
	}
}

func TestApplyTacticsRejectsWrongSize(t *testing.T) {
	slots := map[uint32]int{5: 0}
	data := make([]byte, tacticsSlotSize)
	dir := t.TempDir()
	short := make([]byte, 100)
	if err := os.WriteFile(filepath.Join(dir, "5-short.PES2021_tactics"), short, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := applyTacticsToData(data, slots, []tacticsListEntry{
		{teamID: 5, fileName: "5-short.PES2021_tactics"},
	}, dir)
	if err == nil || !strings.Contains(err.Error(), "expected 628") {
		t.Fatalf("expected size error, got %v", err)
	}
}
