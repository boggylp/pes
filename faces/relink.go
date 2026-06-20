package main

import (
	"bytes"
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

// detectEmbeddedID returns the player ID embedded in the first #Win package that
// carries a face/real/<id> path. The package is authoritative over any
// caller-supplied ID (e.g. a CSV folder name that has drifted from the FMDL
// contents), so the alias target and length decision both come from here.
func detectEmbeddedID(pkgs []string) (string, error) {
	for _, p := range pkgs {
		raw, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		if id, derr := findEmbeddedID(raw); derr == nil {
			return id, nil
		}
	}
	return "", errNoEmbeddedID
}

// rewriteFaceFolderID points a placed face folder at newID, deriving the current
// ID from the packages themselves.
//
// Equal-length IDs are swapped in place inside every #Win package that carries
// the face/real/<id>/ path: a size-preserving byte replace, format-agnostic
// across foxfpk and foxfpkd.
//
// A length change must NOT touch the packages. The ID sits in the FMDL's packed
// null-terminated string blob, reached through a string offset table; growing or
// shrinking the string shifts every following string while the offsets keep
// their old values, so the texture filenames the model reads come out off by the
// length delta and the face renders with missing textures. Instead the textures
// are aliased (see aliasFaceTextures): the FMDL keeps pointing at the original
// face/real/<oldID>/sourceimages, and that directory is provided. The model
// still loads, because binding is by folder name, not by the embedded path.
func rewriteFaceFolderID(folder, newID string) error {
	pkgs, err := winPackages(folder)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return errNoFpk
	}
	oldID, err := detectEmbeddedID(pkgs)
	if err != nil {
		return err
	}
	if oldID == newID {
		return nil // already correct
	}
	if len(oldID) != len(newID) {
		return aliasFaceTextures(folder, oldID)
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
		if err := os.WriteFile(p, bytes.ReplaceAll(raw, old, nw), 0o644); err != nil {
			return err
		}
		rewrote = true
	}
	if !rewrote {
		return errNoEmbeddedID
	}
	return nil
}

// aliasFaceTextures makes a length-mismatched face render without mutating its
// FMDL: it mirrors the folder's sourceimages into a sibling directory named for
// the ID still embedded in the package, so the embedded
// face/real/<oldID>/sourceimages path resolves against the same livecpk root.
// It refuses when that sibling is itself a real face folder (has #Win), to avoid
// clobbering another player's face.
func aliasFaceTextures(folder, oldID string) error {
	src := filepath.Join(folder, "sourceimages")
	if fi, err := os.Stat(src); err != nil || !fi.IsDir() {
		return fmt.Errorf("face folder %s has no sourceimages to alias for embedded id %s", folder, oldID)
	}
	aliasRoot := filepath.Join(filepath.Dir(folder), oldID)
	if _, err := os.Stat(filepath.Join(aliasRoot, "#Win")); err == nil {
		return fmt.Errorf("alias target %s is an existing face folder; refusing to overwrite", aliasRoot)
	}
	dst := filepath.Join(aliasRoot, "sourceimages")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return copyDir(src, dst)
}
