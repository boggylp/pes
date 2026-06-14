package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// embeddedIDPattern matches the player ID inside a face FMDL's texture source
// paths, e.g. ".../face/real/21665/sourceimages/".
var embeddedIDPattern = regexp.MustCompile(`face/real/(\d+)/`)

// findEmbeddedID returns the player ID currently embedded in a face.fpk's
// packed data, or an error if none is present.
func findEmbeddedID(raw []byte) (string, error) {
	m := embeddedIDPattern.FindSubmatch(raw)
	if m == nil {
		return "", fmt.Errorf("no embedded face/real/<id> path found")
	}
	return string(m[1]), nil
}

// relinkFpkBytes rewrites every occurrence of oldID with newID inside the
// packed entry data of a foxfpk and repacks the archive. Unlike an in-place
// byte replace, it tolerates a length change: it recomputes each entry's data
// offset/size (16-byte aligned) and the header file size. This is sound because
// the IDs live in FMDL texture-path strings that nothing references by offset,
// and the FMDL carries no self-size field, so no FMDL-internal fixup is needed.
func relinkFpkBytes(raw []byte, oldID, newID string) ([]byte, error) {
	f, err := parseFpk(raw)
	if err != nil {
		return nil, err
	}
	if len(f.entries) == 0 {
		return nil, fmt.Errorf("foxfpk has no entries")
	}

	old, nw := []byte(oldID), []byte(newID)
	newData := make([][]byte, len(f.entries))
	changed := false
	for i, e := range f.entries {
		d := raw[e.dataOffset : e.dataOffset+e.dataSize]
		if bytes.Contains(d, old) {
			newData[i] = bytes.ReplaceAll(d, old, nw)
			changed = true
		} else {
			newData[i] = d
		}
	}
	if !changed {
		return nil, fmt.Errorf("id %q not found in any packed entry", oldID)
	}

	// Everything before the first data block (header, entry table, reference
	// table, name strings) is preserved; only entry-table offsets/sizes and the
	// header file size change.
	firstOff := f.entries[0].dataOffset
	for _, e := range f.entries {
		if e.dataOffset < firstOff {
			firstOff = e.dataOffset
		}
	}
	if int(firstOff) > len(raw) {
		return nil, fmt.Errorf("first data offset past EOF")
	}
	out := make([]byte, firstOff)
	copy(out, raw[:firstOff])

	order := make([]int, len(f.entries))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return f.entries[order[a]].dataOffset < f.entries[order[b]].dataOffset
	})

	cursor := int(firstOff)
	for _, i := range order {
		if pad := cursor % fpkAlign; pad != 0 {
			out = append(out, make([]byte, fpkAlign-pad)...)
			cursor += fpkAlign - pad
		}
		ent := fpkHeaderSize + i*fpkEntrySize
		binary.LittleEndian.PutUint64(out[ent+32:], uint64(cursor))
		binary.LittleEndian.PutUint64(out[ent+40:], uint64(len(newData[i])))
		out = append(out, newData[i]...)
		cursor += len(newData[i])
	}
	if pad := len(out) % fpkAlign; pad != 0 { // archives are padded to a 16-byte boundary
		out = append(out, make([]byte, fpkAlign-pad)...)
	}
	binary.LittleEndian.PutUint32(out[0x0A:], uint32(len(out)))
	return out, nil
}

// relinkFaceFolder relinks the face.fpk inside a face folder's #Win directory to
// newID. oldID is auto-detected. Equal-length IDs are handled by the same path.
func relinkFaceFolder(folder, newID string) error {
	fpkPath := filepath.Join(folder, "#Win", "face.fpk")
	raw, err := os.ReadFile(fpkPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil // textures-only face, nothing to relink
	}
	if err != nil {
		return err
	}
	oldID, err := findEmbeddedID(raw)
	if err != nil || oldID == newID {
		return nil // no embedded path, or already correct
	}
	out, err := relinkFpkBytes(raw, oldID, newID)
	if err != nil {
		return fmt.Errorf("%s: %w", fpkPath, err)
	}
	return os.WriteFile(fpkPath, out, 0o644)
}
