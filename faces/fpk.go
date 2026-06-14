package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// foxfpk (PES 64-bit variant) container. Header is 0x40 bytes; each entry is
// 48 bytes; packed file data follows, 16-byte aligned. Offsets are absolute.
const (
	fpkHeaderSize = 0x40
	fpkEntrySize  = 48
	fpkAlign      = 16
)

var fpkMagic = []byte("foxfpk")

type fpkEntry struct {
	nameOffset uint64
	nameSize   uint64
	md5        [16]byte
	dataOffset uint64
	dataSize   uint64
	name       string
}

type fpkFile struct {
	raw     []byte
	version uint32
	entries []fpkEntry
}

func parseFpk(b []byte) (*fpkFile, error) {
	if len(b) < fpkHeaderSize || !bytes.HasPrefix(b, fpkMagic) {
		return nil, fmt.Errorf("not a foxfpk archive")
	}
	// PES 64-bit foxfpk: version at 0x0C, packed-file count at 0x20, reference
	// count at 0x24 (references are name-only, no packed data, so ignored here).
	f := &fpkFile{
		raw:     b,
		version: binary.LittleEndian.Uint32(b[0x0C:]),
	}
	fileCount := binary.LittleEndian.Uint32(b[0x20:])
	for i := 0; i < int(fileCount); i++ {
		off := fpkHeaderSize + i*fpkEntrySize
		if off+fpkEntrySize > len(b) {
			return nil, fmt.Errorf("entry %d past EOF", i)
		}
		var e fpkEntry
		e.nameOffset = binary.LittleEndian.Uint64(b[off:])
		e.nameSize = binary.LittleEndian.Uint64(b[off+8:])
		copy(e.md5[:], b[off+16:off+32])
		e.dataOffset = binary.LittleEndian.Uint64(b[off+32:])
		e.dataSize = binary.LittleEndian.Uint64(b[off+40:])
		if e.nameOffset+e.nameSize <= uint64(len(b)) {
			e.name = string(b[e.nameOffset : e.nameOffset+e.nameSize])
		}
		f.entries = append(f.entries, e)
	}
	return f, nil
}
