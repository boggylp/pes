package main

import (
	"bytes"
	"testing"
)

func TestResetSquadOrderInDataWritesIdentity(t *testing.T) {
	// Build a data buffer with two slots populated with a non-identity
	// squad-order array; snapshot before and verify reset overwrites
	// EXACTLY bytes [484..516) per slot and nothing else.
	const totalSize = tacticsSlotSize * 3
	data := make([]byte, totalSize)

	slotA := buildSlot(5, 0xAA)
	for i := 0; i < squadOrderLen; i++ {
		slotA[squadOrderOffset+i] = byte(squadOrderLen - 1 - i)
	}
	copy(data[0:], slotA)

	slotB := buildSlot(7, 0xBB)
	for i := 0; i < squadOrderLen; i++ {
		slotB[squadOrderOffset+i] = 0xFF
	}
	copy(data[tacticsSlotSize:], slotB)

	snapshot := bytes.Clone(data)
	slots := map[uint32]int{5: 0, 7: tacticsSlotSize}

	got := resetSquadOrderInData(data, slots)
	if got != 2 {
		t.Errorf("expected reset count 2, got %d", got)
	}

	// Identity check on both slots.
	for slotIdx, slotOff := range []int{0, tacticsSlotSize} {
		for i := 0; i < squadOrderLen; i++ {
			want := byte(i)
			got := data[slotOff+squadOrderOffset+i]
			if got != want {
				t.Errorf("slot %d byte %d: got 0x%02x want 0x%02x", slotIdx, i, got, want)
			}
		}
	}

	// Verify ALL bytes outside the squad-order regions match the
	// snapshot byte-for-byte. Catches stride miscount, off-by-one,
	// or accidental writes into adjacent fields.
	for i := 0; i < len(data); i++ {
		inSquadA := i >= squadOrderOffset && i < squadOrderOffset+squadOrderLen
		inSquadB := i >= tacticsSlotSize+squadOrderOffset && i < tacticsSlotSize+squadOrderOffset+squadOrderLen
		if inSquadA || inSquadB {
			continue
		}
		if data[i] != snapshot[i] {
			t.Errorf("unexpected mutation at byte %d (0x%x): snapshot=0x%02x post=0x%02x",
				i, i, snapshot[i], data[i])
		}
	}
}

func TestResetSquadOrderInDataRefusesOutOfBoundsSlot(t *testing.T) {
	// A slot offset that would write past the end of data must be
	// skipped silently — never panic — even though scanTacticsSection
	// shouldn't normally produce one.
	data := make([]byte, tacticsSlotSize) // exactly one slot worth
	slots := map[uint32]int{
		5: 0,                              // valid: writes bytes 484..515
		9: tacticsSlotSize - squadOrderLen, // would write 484..515 past the slot end
	}
	got := resetSquadOrderInData(data, slots)
	if got != 1 {
		t.Errorf("expected exactly 1 valid write, got %d", got)
	}
}

func TestResetSquadOrderInDataIdempotent(t *testing.T) {
	// Running reset twice produces the same bytes as running it once.
	data := make([]byte, tacticsSlotSize)
	slots := map[uint32]int{5: 0}
	resetSquadOrderInData(data, slots)
	snapshot := bytes.Clone(data)
	resetSquadOrderInData(data, slots)
	if !bytes.Equal(data, snapshot) {
		t.Error("second reset modified bytes; reset should be idempotent")
	}
}

func TestResetSquadOrderInDataNoopOnEmptyMap(t *testing.T) {
	data := make([]byte, tacticsSlotSize)
	// Fill with a recognizable pattern
	for i := range data {
		data[i] = 0x42
	}
	snapshot := bytes.Clone(data)

	got := resetSquadOrderInData(data, map[uint32]int{})
	if got != 0 {
		t.Errorf("expected reset count 0 on empty slots map, got %d", got)
	}
	if !bytes.Equal(data, snapshot) {
		t.Error("data was modified despite empty slots map")
	}
}

func TestSquadOrderConstantsMatchObservedLayout(t *testing.T) {
	// Lock the empirically-found constants. Anyone changing these
	// without an accompanying RE update against a new FL26 patch
	// version should trip this test.
	if squadOrderOffset != 484 {
		t.Errorf("squadOrderOffset drift: got %d want 484", squadOrderOffset)
	}
	if squadOrderLen != 32 {
		t.Errorf("squadOrderLen drift: got %d want 32", squadOrderLen)
	}
}
