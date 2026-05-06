package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
)

// File describes one entry in the CPK's table of contents.
type File struct {
	Dir         string // logical directory (may be empty)
	Name        string // file name within the CPK
	Offset      int64  // absolute file offset to the payload
	Size        int    // bytes stored in the CPK (compressed if CRILAYLA)
	ExtractSize int    // size after decompression
	UserString  string // optional user-assigned tag, used by some games for encryption
}

// Path returns the full inner path used by tools that index by inner path
// (e.g. "common/etc/pesdb/Player.bin"). When the CPK has no DirName column
// it equals Name.
func (f File) Path() string {
	if f.Dir == "" {
		return f.Name
	}
	return path.Join(f.Dir, f.Name)
}

// Reader holds a file handle to a CPK and the parsed table of contents.
// All file payloads are read on demand from the underlying file.
type Reader struct {
	f             *os.File
	contentOffset int64
	files         []File
}

// Open parses a CPK file's headers and returns a Reader. The caller must
// Close the Reader when done.
func Open(filePath string) (*Reader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	r := &Reader{f: f}
	if err := r.load(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return r, nil
}

// Close releases the underlying file handle.
func (r *Reader) Close() error { return r.f.Close() }

// Files returns the parsed TOC. The slice is shared; callers should not
// mutate it.
func (r *Reader) Files() []File { return r.files }

// FindFile returns the first TOC entry whose Path matches inner. Match is
// case-insensitive on the filename and tolerant of leading "/"s; this matches
// what most PES tooling expects when users paste paths.
func (r *Reader) FindFile(inner string) (*File, bool) {
	target := normalizePath(inner)
	for i := range r.files {
		if normalizePath(r.files[i].Path()) == target {
			return &r.files[i], true
		}
	}
	return nil, false
}

// ReadFile reads a single TOC entry's payload, decompressing CRILAYLA when
// the payload begins with that signature. Returned bytes are exactly
// ExtractSize for compressed entries and Size otherwise.
func (r *Reader) ReadFile(file File) ([]byte, error) {
	if file.Size == 0 {
		return nil, nil
	}
	raw, err := r.readAt(file.Offset, file.Size)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", file.Path(), err)
	}
	if isCriLayla(raw) {
		decompressed, err := criLaylaDecompress(raw)
		if err != nil {
			return nil, fmt.Errorf("decompressing %q: %w", file.Path(), err)
		}
		return decompressed, nil
	}
	return raw, nil
}

func (r *Reader) load() error {
	// CPK container at offset 0: 16 bytes -> "CPK ", marker, table-size, padding
	tbl, err := r.readTableContainer(0, "CPK ")
	if err != nil {
		return fmt.Errorf("reading CPK header: %w", err)
	}
	headerTable, err := parseTable(tbl)
	if err != nil {
		return fmt.Errorf("parsing CPK header table: %w", err)
	}

	// Header table has exactly one row with TocOffset and ContentOffset.
	row, err := headerTable.row(0)
	if err != nil {
		return err
	}
	tocOffset, err := readNamedNumber(row, "TocOffset")
	if err != nil {
		return fmt.Errorf("CPK header: %w", err)
	}
	contentOffset, err := readNamedNumber(row, "ContentOffset")
	if err != nil {
		return fmt.Errorf("CPK header: %w", err)
	}
	// Some CPKs place files relative to the TOC instead of the explicit
	// ContentOffset. CriFsV2Lib's TocFinder normalises this.
	if tocOffset < contentOffset {
		contentOffset = tocOffset
	}
	r.contentOffset = contentOffset

	tocBuf, err := r.readTableContainer(tocOffset, "TOC ")
	if err != nil {
		return fmt.Errorf("reading TOC: %w", err)
	}
	tocTable, err := parseTable(tocBuf)
	if err != nil {
		return fmt.Errorf("parsing TOC table: %w", err)
	}

	r.files = make([]File, 0, tocTable.rowCount)
	for i := 0; i < tocTable.rowCount; i++ {
		row, err := tocTable.row(i)
		if err != nil {
			return err
		}
		file, err := readTocRow(row, contentOffset)
		if err != nil {
			return fmt.Errorf("TOC row %d: %w", i, err)
		}
		r.files = append(r.files, file)
	}
	// Some PES season packs declare ContentOffset == TocOffset near the file
	// end so FileOffsets are already absolute. If every entry's computed
	// Offset+Size would land past EOF but the raw FileOffset would not,
	// re-add the entries with contentOffset=0.
	if st, err := r.f.Stat(); err == nil {
		size := st.Size()
		needFix := false
		for _, f := range r.files {
			if int64(f.Offset)+int64(f.Size) > size {
				needFix = true
				break
			}
		}
		if needFix && contentOffset > 0 {
			ok := true
			for _, f := range r.files {
				raw := int64(f.Offset) - contentOffset
				if raw < 0 || raw+int64(f.Size) > size {
					ok = false
					break
				}
			}
			if ok {
				r.contentOffset = 0
				for i := range r.files {
					r.files[i].Offset -= contentOffset
				}
			}
		}
	}
	return nil
}

func (r *Reader) readAt(offset int64, size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := r.f.ReadAt(buf, offset); err != nil && err != io.EOF {
		return nil, err
	}
	return buf, nil
}

// readTableContainer reads a 16-byte CRI container (CPK / TOC / ETOC) at the
// given offset and returns the @UTF table buffer that follows. The container
// signature is verified against want (a 4-byte literal).
func (r *Reader) readTableContainer(offset int64, want string) ([]byte, error) {
	const containerSize = 16
	hdr, err := r.readAt(offset, containerSize)
	if err != nil {
		return nil, err
	}
	if string(hdr[0:4]) != want {
		return nil, fmt.Errorf("expected container signature %q, got %q at offset %d", want, hdr[0:4], offset)
	}
	// Container size at offset 8 is little-endian, while the @UTF table
	// itself is big-endian. The two magic numbers come from CRI's own writer.
	tableSize := int(binary.LittleEndian.Uint32(hdr[8:12]))
	if tableSize <= 0 {
		return nil, fmt.Errorf("invalid table size %d at offset %d", tableSize, offset)
	}
	return r.readAt(offset+containerSize, tableSize)
}

func readTocRow(r *rowReader, contentOffset int64) (File, error) {
	// Walk columns in declaration order so the row pointer advances correctly.
	var f File
	for _, c := range r.tbl.columns {
		switch c.name {
		case "DirName":
			s, err := readStringInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.Dir = s
		case "FileName":
			s, err := readStringInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.Name = s
		case "FileSize":
			n, err := readNumberInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.Size = int(n)
		case "ExtractSize":
			n, err := readNumberInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.ExtractSize = int(n)
		case "FileOffset":
			n, err := readNumberInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.Offset = n + contentOffset
		case "UserString":
			s, err := readStringInPlace(r, c)
			if err != nil {
				return File{}, err
			}
			f.UserString = s
		default:
			r.skip(c)
		}
	}
	if f.ExtractSize == 0 {
		f.ExtractSize = f.Size
	}
	return f, nil
}

func readNamedNumber(r *rowReader, name string) (int64, error) {
	// Walk all columns, pulling the named value out and skipping the rest.
	// A separate pass keeps the row pointer correctly positioned regardless
	// of which named columns the caller cares about.
	var (
		got     int64
		gotName bool
	)
	for _, c := range r.tbl.columns {
		if c.name == name {
			n, err := readNumberInPlace(r, c)
			if err != nil {
				return 0, err
			}
			got = n
			gotName = true
		} else {
			r.skip(c)
		}
	}
	if !gotName {
		return 0, fmt.Errorf("column %q not found", name)
	}
	return got, nil
}

func readNumberInPlace(r *rowReader, c column) (int64, error) {
	buf, advance := r.valueBytes(c)
	if buf == nil {
		return 0, fmt.Errorf("column %q has no value", c.name)
	}
	n, err := decodeNumber(buf, c.typ)
	if advance {
		r.rowPtr += c.typ.size()
	}
	return n, err
}

func readStringInPlace(r *rowReader, c column) (string, error) {
	if c.typ != typeString {
		return "", fmt.Errorf("column %q is not a string", c.name)
	}
	buf, advance := r.valueBytes(c)
	if buf == nil {
		return "", fmt.Errorf("column %q has no value", c.name)
	}
	off := int(int32(binary.BigEndian.Uint32(buf))) + r.tbl.stringsO
	s, err := r.tbl.readStringAt(off)
	if advance {
		r.rowPtr += 4
	}
	return s, err
}

func normalizePath(p string) string {
	for len(p) > 0 && (p[0] == '/' || p[0] == '\\') {
		p = p[1:]
	}
	out := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		c := p[i]
		if c == '\\' {
			c = '/'
		}
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}
