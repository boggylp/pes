package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// embeddedIDPattern matches the player ID inside a face FMDL's texture source
// paths, e.g. ".../face/real/21665/sourceimages/".
var embeddedIDPattern = regexp.MustCompile(`face/real/(\d+)/`)

// Sentinel results from relinkFaceFolder: a standalone `relink` treats these as
// failures; the map path tolerates them (the face copied, nothing to rewrite).
var (
	errNoFpk        = errors.New("no face package (#Win/*.fpk|*.fpkd) in face folder")
	errNoEmbeddedID = errors.New("no embedded face/real/<id> path in any face package")
)

// winPackages returns the #Win/*.fpk and #Win/*.fpkd files in a face folder, in
// deterministic order. The player ID lives in face.fpk; the companion face.fpkd
// is usually an ID-less dependency stub, but separate packages can also embed
// the path, so every package is processed rather than only face.fpk.
func winPackages(folder string) ([]string, error) {
	winDir := filepath.Join(folder, "#Win")
	var pkgs []string
	for _, ext := range []string{"*.fpk", "*.fpkd"} {
		m, err := filepath.Glob(filepath.Join(winDir, ext))
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, m...)
	}
	sort.Strings(pkgs)
	return pkgs, nil
}

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

	// Anchor on the path so a short numeric ID can't rewrite incidental byte
	// runs (vertex data, other strings) that happen to match the bare digits.
	old := []byte("face/real/" + oldID + "/")
	nw := []byte("face/real/" + newID + "/")
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
	entTableEnd := fpkHeaderSize + len(f.entries)*fpkEntrySize
	if int(firstOff) < entTableEnd || int(firstOff) > len(raw) {
		return nil, fmt.Errorf("first data offset 0x%X overlaps entry table or past EOF", firstOff)
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

// rewriteFaceFolderID rewrites the embedded face/real/<oldID>/ path to newID in
// every #Win face package that contains it. Equal-length IDs use an in-place
// byte replace, which is format-agnostic and size-preserving (safe for foxfpk
// and foxfpkd alike). A length change needs the foxfpk repacker; a length change
// in a foxfpkd is refused rather than risk corrupting an unverified layout.
func rewriteFaceFolderID(folder, oldID, newID string) error {
	if oldID == newID {
		return nil // already correct
	}
	pkgs, err := winPackages(folder)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return errNoFpk
	}
	old := []byte("face/real/" + oldID + "/")
	nw := []byte("face/real/" + newID + "/")
	rewrote := false
	for _, p := range pkgs {
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if !bytes.Contains(raw, old) {
			continue // ID-less stub or unrelated package
		}
		var out []byte
		switch {
		case len(oldID) == len(newID):
			out = bytes.ReplaceAll(raw, old, nw)
		case isFoxFpk(raw):
			if out, err = relinkFpkBytes(raw, oldID, newID); err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
		default:
			return fmt.Errorf("%s: length-changing relink of a %s package is unsupported", p, containerKind(raw))
		}
		if err := os.WriteFile(p, out, 0o644); err != nil {
			return err
		}
		rewrote = true
	}
	if !rewrote {
		return errNoEmbeddedID
	}
	return nil
}

// relinkFaceFolder auto-detects the current embedded ID from a face folder's
// #Win packages and rewrites every package to newID. Used by the standalone
// `relink` command, where only the new ID is known.
func relinkFaceFolder(folder, newID string) error {
	pkgs, err := winPackages(folder)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return errNoFpk
	}
	oldID := ""
	for _, p := range pkgs {
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if id, derr := findEmbeddedID(raw); derr == nil {
			oldID = id
			break
		}
	}
	if oldID == "" {
		return errNoEmbeddedID
	}
	return rewriteFaceFolderID(folder, oldID, newID)
}
