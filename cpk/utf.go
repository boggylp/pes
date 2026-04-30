package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// CRI @UTF table parser. The layout is from CRI Middleware's CPK format.
//
// Table structure inside the table buffer:
//
//	0x00-0x03  "@UTF"
//	0x04-0x07  table data size (BE uint32)  — same value as the container size
//	0x08       (unused)
//	0x09       encoding flag: 0 = Shift-JIS, anything else = UTF-8
//	0x0A-0x0B  rows offset (BE uint16, relative to baseOffset)
//	0x0C-0x0F  string pool offset (BE int32, relative to baseOffset)
//	0x10-0x13  data pool offset (BE int32, relative to baseOffset)
//	0x14-0x17  table-name string offset (unused here)
//	0x18-0x19  column count (BE uint16)
//	0x1A-0x1B  row size in bytes (BE uint16)
//	0x1C-0x1F  row count (BE int32)
//	0x20+      column descriptors, then rows, then string pool, then data pool
//
// All offsets stored in the header are relative to byte 8 of the buffer
// (i.e. just after "@UTF" + size4); we add tableBaseOffset to get absolute
// indices into the buffer.
const tableBaseOffset = 0x08
const columnDataOffset = 0x20

type stringEncoding int

const (
	encShiftJIS stringEncoding = iota
	encUTF8
)

type columnFlag byte

const (
	flagHasName         columnFlag = 0x10
	flagHasDefaultValue columnFlag = 0x20
	flagIsRowStorage    columnFlag = 0x40
	typeMask            columnFlag = 0x0F
)

type columnType byte

const (
	typeByte    columnType = 0
	typeSByte   columnType = 1
	typeUInt16  columnType = 2
	typeInt16   columnType = 3
	typeUInt32  columnType = 4
	typeInt32   columnType = 5
	typeUInt64  columnType = 6
	typeInt64   columnType = 7
	typeSingle  columnType = 8
	typeDouble  columnType = 9
	typeString  columnType = 10
	typeRawData columnType = 11
	typeGUID    columnType = 12
)

func (t columnType) size() int {
	switch t {
	case typeByte, typeSByte:
		return 1
	case typeUInt16, typeInt16:
		return 2
	case typeUInt32, typeInt32, typeSingle, typeString:
		return 4
	case typeUInt64, typeInt64, typeDouble, typeRawData:
		return 8
	case typeGUID:
		return 16
	}
	return 0
}

type column struct {
	flags    columnFlag
	typ      columnType
	name     string
	defValue []byte // present iff flags & flagHasDefaultValue
}

func (c column) hasFlag(f columnFlag) bool { return c.flags&f == f }

type table struct {
	buf      []byte
	rowsOff  int
	stringsO int
	dataOff  int
	rowSize  int
	rowCount int
	enc      stringEncoding
	columns  []column
}

func parseTable(buf []byte) (*table, error) {
	if len(buf) < columnDataOffset {
		return nil, fmt.Errorf("table buffer too small: %d bytes", len(buf))
	}
	if string(buf[0:4]) != "@UTF" {
		return nil, fmt.Errorf("not an @UTF table (magic %q)", buf[0:4])
	}

	t := &table{
		buf:      buf,
		rowsOff:  int(binary.BigEndian.Uint16(buf[0x0A:0x0C])) + tableBaseOffset,
		stringsO: int(int32(binary.BigEndian.Uint32(buf[0x0C:0x10]))) + tableBaseOffset,
		dataOff:  int(int32(binary.BigEndian.Uint32(buf[0x10:0x14]))) + tableBaseOffset,
		rowSize:  int(binary.BigEndian.Uint16(buf[0x1A:0x1C])),
		rowCount: int(int32(binary.BigEndian.Uint32(buf[0x1C:0x20]))),
	}
	if buf[0x09] == 0 {
		t.enc = encShiftJIS
	} else {
		t.enc = encUTF8
	}
	colCount := int(binary.BigEndian.Uint16(buf[0x18:0x1A]))

	cur := columnDataOffset
	for i := 0; i < colCount; i++ {
		if cur >= len(buf) {
			return nil, fmt.Errorf("column %d: buffer overrun at %d", i, cur)
		}
		flagsByte := buf[cur]
		col := column{
			flags: columnFlag(flagsByte) &^ typeMask,
			typ:   columnType(flagsByte & byte(typeMask)),
		}
		cur++

		if col.hasFlag(flagHasName) {
			if cur+4 > len(buf) {
				return nil, fmt.Errorf("column %d: name offset overrun", i)
			}
			nameOff := int(int32(binary.BigEndian.Uint32(buf[cur:cur+4]))) + t.stringsO
			cur += 4
			name, err := t.readStringAt(nameOff)
			if err != nil {
				return nil, fmt.Errorf("column %d name: %w", i, err)
			}
			col.name = name
		}
		if col.hasFlag(flagHasDefaultValue) {
			n := col.typ.size()
			if cur+n > len(buf) {
				return nil, fmt.Errorf("column %d: default value overrun", i)
			}
			col.defValue = buf[cur : cur+n]
			cur += n
		}
		t.columns = append(t.columns, col)
	}
	return t, nil
}

// readStringAt reads a NUL-terminated string at an absolute offset in the
// table buffer, decoding using the table's encoding.
func (t *table) readStringAt(off int) (string, error) {
	if off < 0 || off >= len(t.buf) {
		return "", fmt.Errorf("string offset %d out of range (buf=%d)", off, len(t.buf))
	}
	end := off
	for end < len(t.buf) && t.buf[end] != 0 {
		end++
	}
	raw := t.buf[off:end]
	if t.enc == encUTF8 {
		return string(raw), nil
	}
	decoded, _, err := transform.Bytes(japanese.ShiftJIS.NewDecoder(), raw)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// rowReader walks one row of the table and exposes typed accessors keyed by
// column name. Values from columns flagged HasDefaultValue (no row storage)
// come from the column's default; columns with IsRowStorage advance the row
// pointer.
type rowReader struct {
	tbl    *table
	rowPtr int
}

func (t *table) row(idx int) (*rowReader, error) {
	if idx < 0 || idx >= t.rowCount {
		return nil, fmt.Errorf("row index %d out of range (count=%d)", idx, t.rowCount)
	}
	return &rowReader{tbl: t, rowPtr: t.rowsOff + idx*t.rowSize}, nil
}

// valueBytes returns the raw bytes the column refers to for this row, plus
// whether the value came from row storage (caller advances) or the column
// default (caller does not advance).
func (r *rowReader) valueBytes(c column) ([]byte, bool) {
	if c.hasFlag(flagIsRowStorage) {
		n := c.typ.size()
		if r.rowPtr+n > len(r.tbl.buf) {
			return nil, true
		}
		return r.tbl.buf[r.rowPtr : r.rowPtr+n], true
	}
	if c.hasFlag(flagHasDefaultValue) {
		return c.defValue, false
	}
	return nil, false
}

// readNumber returns the column value coerced to int64.
func (r *rowReader) readNumber(name string) (int64, error) {
	for _, c := range r.tbl.columns {
		if c.name != name {
			continue
		}
		buf, advance := r.valueBytes(c)
		if buf == nil {
			return 0, fmt.Errorf("column %q has no value for this row", name)
		}
		v, err := decodeNumber(buf, c.typ)
		if advance {
			r.rowPtr += c.typ.size()
		}
		return v, err
	}
	return 0, fmt.Errorf("column %q not found", name)
}

// readString returns the string referenced by a column of typeString.
func (r *rowReader) readString(name string) (string, error) {
	for _, c := range r.tbl.columns {
		if c.name != name {
			continue
		}
		if c.typ != typeString {
			return "", fmt.Errorf("column %q is not a string (type=%d)", name, c.typ)
		}
		buf, advance := r.valueBytes(c)
		if buf == nil {
			return "", fmt.Errorf("column %q has no value for this row", name)
		}
		off := int(int32(binary.BigEndian.Uint32(buf))) + r.tbl.stringsO
		s, err := r.tbl.readStringAt(off)
		if advance {
			r.rowPtr += 4
		}
		return s, err
	}
	return "", fmt.Errorf("column %q not found", name)
}

// advance walks past a column's row storage without reading it. Lets callers
// position the row pointer correctly when they only need a subset of fields.
func (r *rowReader) skip(c column) {
	if c.hasFlag(flagIsRowStorage) {
		r.rowPtr += c.typ.size()
	}
}

func decodeNumber(buf []byte, t columnType) (int64, error) {
	if len(buf) < t.size() {
		return 0, fmt.Errorf("buffer too small for type %d: %d", t, len(buf))
	}
	switch t {
	case typeByte:
		return int64(buf[0]), nil
	case typeSByte:
		return int64(int8(buf[0])), nil
	case typeUInt16:
		return int64(binary.BigEndian.Uint16(buf)), nil
	case typeInt16:
		return int64(int16(binary.BigEndian.Uint16(buf))), nil
	case typeUInt32:
		return int64(binary.BigEndian.Uint32(buf)), nil
	case typeInt32:
		return int64(int32(binary.BigEndian.Uint32(buf))), nil
	case typeUInt64, typeInt64:
		return int64(binary.BigEndian.Uint64(buf)), nil
	}
	return 0, fmt.Errorf("non-numeric column type %d", t)
}
