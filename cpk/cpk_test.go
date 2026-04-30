package main

import "testing"

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"common/etc/pesdb/Player.bin", "common/etc/pesdb/player.bin"},
		{"/common/etc/pesdb/Player.bin", "common/etc/pesdb/player.bin"},
		{`Common\Etc\Pesdb\Player.bin`, "common/etc/pesdb/player.bin"},
		{"", ""},
		{"/", ""},
	}
	for _, tc := range cases {
		got := normalizePath(tc.in)
		if got != tc.want {
			t.Errorf("normalizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDecodeNumberEndianness(t *testing.T) {
	// Big-endian uint16 0x0102 -> 258.
	v, err := decodeNumber([]byte{0x01, 0x02}, typeUInt16)
	if err != nil {
		t.Fatalf("decodeNumber uint16: %v", err)
	}
	if v != 258 {
		t.Errorf("uint16 BE 0x0102 = %d, want 258", v)
	}

	// Big-endian int32 0xFFFFFFFE -> -2.
	v, err = decodeNumber([]byte{0xFF, 0xFF, 0xFF, 0xFE}, typeInt32)
	if err != nil {
		t.Fatalf("decodeNumber int32: %v", err)
	}
	if v != -2 {
		t.Errorf("int32 BE 0xFFFFFFFE = %d, want -2", v)
	}

	// Signed byte 0xFF -> -1.
	v, err = decodeNumber([]byte{0xFF}, typeSByte)
	if err != nil {
		t.Fatalf("decodeNumber sbyte: %v", err)
	}
	if v != -1 {
		t.Errorf("sbyte 0xFF = %d, want -1", v)
	}
}

func TestColumnTypeSize(t *testing.T) {
	// Catches accidental drift from the CRI spec: numeric types must keep
	// their declared widths or row pointer arithmetic walks off the row.
	cases := map[columnType]int{
		typeByte:    1,
		typeSByte:   1,
		typeUInt16:  2,
		typeInt16:   2,
		typeUInt32:  4,
		typeInt32:   4,
		typeUInt64:  8,
		typeInt64:   8,
		typeSingle:  4,
		typeDouble:  8,
		typeString:  4,
		typeRawData: 8,
		typeGUID:    16,
	}
	for typ, want := range cases {
		if got := typ.size(); got != want {
			t.Errorf("type %d size = %d, want %d", typ, got, want)
		}
	}
}

func TestIsCriLayla(t *testing.T) {
	if !isCriLayla([]byte("CRILAYLA\x00\x00")) {
		t.Error("expected CRILAYLA magic to match")
	}
	if isCriLayla([]byte("CRILAYL\x00")) {
		t.Error("short prefix should not match")
	}
	if isCriLayla([]byte("CRIPAKnnn")) {
		t.Error("non-LAYLA prefix should not match")
	}
}
