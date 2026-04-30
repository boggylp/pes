package main

import (
	"encoding/binary"
	"fmt"
)

// CRI Middleware uses an LZSS-derived codec named CRILAYLA. Files compressed
// with it begin with the literal 8-byte magic "CRILAYLA". Any other byte
// sequence at the start of a payload means the file is stored uncompressed.
//
// PES 2017–2021 cpks usually store files uncompressed, so CRILAYLA support
// is a fallback that older patches and some asset packs still rely on.

var criLaylaMagic = []byte("CRILAYLA")

func isCriLayla(buf []byte) bool {
	if len(buf) < len(criLaylaMagic) {
		return false
	}
	for i, b := range criLaylaMagic {
		if buf[i] != b {
			return false
		}
	}
	return true
}

// CRILAYLA payload layout after the 8-byte magic:
//
//	+0x00  "CRILAYLA"
//	+0x08  uncompressedSize  (LE int32)
//	+0x0C  compressedSize    (LE int32) — bytes between 0x10 and the trailing
//	                                       0x100-byte raw prefix block.
//	+0x10  ... compressed bitstream, written *backwards* from the end
//	+0x10 + compressedSize   uncompressed prefix (always 0x100 bytes)
//
// Decoding direction: the bitstream is consumed from end-to-start, the output
// is written from end-to-start. Once the bitstream is exhausted, the trailing
// 0x100 raw bytes are copied to the very start of the output.
func criLaylaDecompress(in []byte) ([]byte, error) {
	const headerSize = 0x10
	const rawPrefixSize = 0x100

	if len(in) < headerSize+rawPrefixSize {
		return nil, fmt.Errorf("CRILAYLA payload too small: %d bytes", len(in))
	}
	uncompressedSize := int(binary.LittleEndian.Uint32(in[0x08:0x0C]))
	compressedSize := int(binary.LittleEndian.Uint32(in[0x0C:0x10]))
	if uncompressedSize <= 0 {
		return nil, fmt.Errorf("CRILAYLA invalid uncompressed size %d", uncompressedSize)
	}
	if headerSize+compressedSize+rawPrefixSize > len(in) {
		return nil, fmt.Errorf("CRILAYLA truncated: hdr+%d+%d > %d", compressedSize, rawPrefixSize, len(in))
	}

	out := make([]byte, uncompressedSize+rawPrefixSize)
	copy(out[:rawPrefixSize], in[headerSize+compressedSize:headerSize+compressedSize+rawPrefixSize])

	// The bitstream lives in [headerSize, headerSize+compressedSize). We read
	// it MSB-first, byte-by-byte from the end. `bitPos` is the absolute bit
	// index from the *start* of the bitstream of the next bit to consume,
	// counting from the high end (so reading "the next bit" means consuming
	// the byte at headerSize + compressedSize - 1 - bitPos/8 and selecting
	// bit (bitPos & 7) from its low end).
	bitPos := 0
	totalBits := compressedSize * 8

	readBit := func() (byte, error) {
		if bitPos >= totalBits {
			return 0, fmt.Errorf("CRILAYLA bitstream underrun at bit %d", bitPos)
		}
		bytePos := headerSize + compressedSize - 1 - bitPos/8
		b := (in[bytePos] >> uint(bitPos&7)) & 1
		bitPos++
		return b, nil
	}
	readBits := func(n int) (uint32, error) {
		var v uint32
		for i := 0; i < n; i++ {
			b, err := readBit()
			if err != nil {
				return 0, err
			}
			v = (v << 1) | uint32(b)
		}
		return v, nil
	}

	// Variable-length copy length encoding: 2 then 3 then 5 then 8 bits, with
	// each "max" rolling forward into the next bucket.
	vleLengths := []int{2, 3, 5, 8}

	writePos := uncompressedSize + rawPrefixSize - 1
	for writePos >= rawPrefixSize {
		flag, err := readBit()
		if err != nil {
			return nil, err
		}
		if flag == 0 {
			// Literal byte: 8 bits.
			b, err := readBits(8)
			if err != nil {
				return nil, err
			}
			out[writePos] = byte(b)
			writePos--
			continue
		}
		// Copy: 13-bit offset (added to current write+1 in a sliding-window
		// scheme) plus a variable-length count >= 3.
		off, err := readBits(13)
		if err != nil {
			return nil, err
		}
		length := 3
		bucket := 0
		for {
			n := vleLengths[bucket]
			v, err := readBits(n)
			if err != nil {
				return nil, err
			}
			length += int(v)
			max := (1 << uint(n)) - 1
			if int(v) != max {
				break
			}
			if bucket+1 < len(vleLengths) {
				bucket++
			}
		}
		// Output is being filled backwards, so the back-reference reads from
		// later in the buffer (writePos+1+off) and copies forward in time
		// while writePos walks down.
		src := writePos + 1 + int(off)
		for i := 0; i < length; i++ {
			if src >= len(out) || writePos < 0 {
				return nil, fmt.Errorf("CRILAYLA copy out of bounds (src=%d writePos=%d len=%d)", src, writePos, length)
			}
			out[writePos] = out[src]
			writePos--
			src--
		}
	}

	return out, nil
}
