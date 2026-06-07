// uniparam edits PES / Football Life UniformParameter.bin and UniColor.bin
// (Konami WESYS + zlib envelopes).
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid number %q\n", s)
		os.Exit(2)
	}
	return n
}

var wesysTail = []byte("WESYS")

type envelope struct {
	header []byte // 16-byte envelope header (3-byte prefix + WESYS + sizes)
	body   []byte // inflated body
}

func readEnvelope(path string) (*envelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	w := bytes.Index(data, wesysTail)
	if w < 3 {
		return nil, errors.New("WESYS magic not found")
	}
	start := w - 3
	if len(data) < start+16 {
		return nil, errors.New("truncated envelope header")
	}
	zr, err := zlib.NewReader(bytes.NewReader(data[start+16:]))
	if err != nil {
		return nil, fmt.Errorf("zlib reader: %w", err)
	}
	body, err := io.ReadAll(zr)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("inflate: %w", err)
	}
	return &envelope{header: append([]byte(nil), data[start:start+16]...), body: body}, nil
}

type entry struct{ recOff, recSize, nameOff uint32 }

func parse(body []byte) (count, field4 uint32, idx []entry, err error) {
	if len(body) < 8 {
		return 0, 0, nil, errors.New("body shorter than header")
	}
	count = binary.LittleEndian.Uint32(body[0:4])
	field4 = binary.LittleEndian.Uint32(body[4:8])
	if uint64(len(body)) < uint64(8)+uint64(count)*12 {
		return 0, 0, nil, fmt.Errorf("body too short for %d index entries", count)
	}
	idx = make([]entry, count)
	for i := uint32(0); i < count; i++ {
		o := 8 + i*12
		idx[i] = entry{
			recOff:  binary.LittleEndian.Uint32(body[o : o+4]),
			recSize: binary.LittleEndian.Uint32(body[o+4 : o+8]),
			nameOff: binary.LittleEndian.Uint32(body[o+8 : o+12]),
		}
	}
	return count, field4, idx, nil
}

func cstr(body []byte, off uint32) string {
	end := off
	for end < uint32(len(body)) && body[end] != 0 {
		end++
	}
	return string(body[off:end])
}

func appendEntry(out []byte, e entry) []byte {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[0:4], e.recOff)
	binary.LittleEndian.PutUint32(b[4:8], e.recSize)
	binary.LittleEndian.PutUint32(b[8:12], e.nameOff)
	return append(out, b[:]...)
}

func serialize(body []byte, count uint32, idx []entry, insertAt int, newName string, newRec []byte) []byte {
	nameTableStart := 8 + count*12
	tail := body[nameTableStart:]

	newCount := count + 1
	newTailBase := 8 + newCount*12
	newNameOff := newTailBase + uint32(len(tail))
	newRecOff := newNameOff + uint32(len(newName)) + 1

	out := make([]byte, 0, len(body)+12+len(newName)+1+len(newRec))
	hdr := make([]byte, 8)
	binary.LittleEndian.PutUint32(hdr[0:4], newCount)
	binary.LittleEndian.PutUint32(hdr[4:8], 8)
	out = append(out, hdr...)

	ne := entry{recOff: newRecOff, recSize: uint32(len(newRec)), nameOff: newNameOff}
	for i, e := range idx {
		if i == insertAt {
			out = appendEntry(out, ne)
		}
		out = appendEntry(out, entry{recOff: e.recOff + 12, recSize: e.recSize, nameOff: e.nameOff + 12})
	}
	if insertAt >= len(idx) {
		out = appendEntry(out, ne)
	}
	out = append(out, tail...)
	out = append(out, []byte(newName)...)
	out = append(out, 0)
	out = append(out, newRec...)
	return out
}

func writeEnvelope(path string, magic, body []byte) error {
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	if _, err := zw.Write(body); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	out := make([]byte, 0, 16+z.Len())
	out = append(out, magic[:8]...)
	var sz [8]byte
	binary.LittleEndian.PutUint32(sz[0:4], uint32(z.Len()))
	binary.LittleEndian.PutUint32(sz[4:8], uint32(len(body)))
	out = append(out, sz[:]...)
	out = append(out, z.Bytes()...)
	return os.WriteFile(path, out, 0o644)
}

func addSlot(inPath, outPath string, team, srcN, dstN int, recordPath string) error {
	env, err := readEnvelope(inPath)
	if err != nil {
		return err
	}
	count, _, idx, err := parse(env.body)
	if err != nil {
		return err
	}
	ord := map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th"}
	srcName := fmt.Sprintf("%d_DEF_%s_realUni.bin", team, ord[srcN])
	dstName := fmt.Sprintf("%d_DEF_%s_realUni.bin", team, ord[dstN])

	srcIdx := -1
	for i, e := range idx {
		switch cstr(env.body, e.nameOff) {
		case srcName:
			srcIdx = i
		case dstName:
			return fmt.Errorf("%s already exists at index %d", dstName, i)
		}
	}
	if srcIdx < 0 {
		return fmt.Errorf("source slot %s not found", srcName)
	}
	src := idx[srcIdx]
	var rec []byte
	if recordPath != "" {
		rec, err = os.ReadFile(recordPath)
		if err != nil {
			return err
		}
		if uint32(len(rec)) != src.recSize {
			return fmt.Errorf("donor record %d B, expected %d to match %s", len(rec), src.recSize, srcName)
		}
	} else {
		rec = append([]byte(nil), env.body[src.recOff:src.recOff+src.recSize]...)
		texSrc := fmt.Sprintf("u%04dp%d", team, srcN)
		texDst := fmt.Sprintf("u%04dp%d", team, dstN)
		rec = bytes.ReplaceAll(rec, []byte(texSrc), []byte(texDst))
	}

	body := serialize(env.body, count, idx, srcIdx+1, dstName, rec)
	if err := writeEnvelope(outPath, env.header, body); err != nil {
		return err
	}
	chk, err := readEnvelope(outPath)
	if err != nil {
		return fmt.Errorf("verify reopen: %w", err)
	}
	c2, _, idx2, err := parse(chk.body)
	if err != nil {
		return fmt.Errorf("verify parse: %w", err)
	}
	found := false
	for _, e := range idx2 {
		if cstr(chk.body, e.nameOff) == dstName && bytes.Equal(chk.body[e.recOff:e.recOff+e.recSize], rec) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("verify: %s missing or wrong after write", dstName)
	}
	fmt.Printf("wrote %s: count %d->%d, %s present (record %dB)\n", outPath, count, c2, dstName, len(rec))
	return nil
}

func unicolorAddSlot(inPath, outPath string, team int) error {
	env, err := readEnvelope(inPath)
	if err != nil {
		return err
	}
	body := env.body
	var tb [4]byte
	binary.LittleEndian.PutUint32(tb[:], uint32(team))
	off := bytes.Index(body, tb[:])
	if off < 0 {
		return fmt.Errorf("team %d not found in UniColor", team)
	}
	cntOff := off + 4
	count := int(body[cntOff])
	slots := off + 5
	gk, players := -1, 0
	for i := 0; i < count; i++ {
		switch m := body[slots+i*8]; {
		case m == 0x10:
			gk = i
		case m < 0x10:
			players++
		}
	}
	if gk < 0 {
		return fmt.Errorf("no GK1st (marker 0x10) slot for team %d", team)
	}
	if gk != count-1 {
		return fmt.Errorf("unexpected layout: GK at slot %d of %d (want last)", gk, count)
	}
	if players < 1 {
		return fmt.Errorf("team %d has no player slot to model", team)
	}
	if count >= 10 {
		return fmt.Errorf("team %d UniColor record full (%d slots)", team, count)
	}
	gkEntry := append([]byte(nil), body[slots+gk*8:slots+gk*8+8]...)
	newEntry := append([]byte(nil), body[slots+(players-1)*8:slots+(players-1)*8+8]...)
	newEntry[0] = byte(players) // next player marker (00,01 -> 02)
	copy(body[slots+gk*8:], newEntry)        // new player where GK was
	copy(body[slots+(gk+1)*8:], gkEntry)     // GK shifts into the padding slot
	body[cntOff] = byte(count + 1)
	if err := writeEnvelope(outPath, env.header, body); err != nil {
		return err
	}
	fmt.Printf("wrote %s: team %d count %d->%d, added player marker %d, GK moved to slot %d\n", outPath, team, count, count+1, players, gk+1)
	return nil
}

func retex(inPath, outPath, from, to string) error {
	if len(from) != len(to) {
		return fmt.Errorf("from/to length differ (%d vs %d)", len(from), len(to))
	}
	b, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	n := bytes.Count(b, []byte(from))
	if n == 0 {
		return fmt.Errorf("texture %q not found in %s", from, inPath)
	}
	b = bytes.ReplaceAll(b, []byte(from), []byte(to))
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s: %s->%s (%d), %d bytes\n", outPath, from, to, n, len(b))
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage:\n  uniparam unicolor-addslot <in> <out> <team>\n  uniparam add-slot <in> <out> <team> <srcSlot> <dstSlot> [donorRecord]\n  uniparam retex <src> <out> <fromTex> <toTex>\n  uniparam inflate <in> <out>")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "add-slot":
		if len(os.Args) != 7 && len(os.Args) != 8 {
			fmt.Fprintln(os.Stderr, "usage: uniparam add-slot <in> <out> <team> <srcSlot> <dstSlot> [donorRecord.bin]")
			os.Exit(2)
		}
		record := ""
		if len(os.Args) == 8 {
			record = os.Args[7]
		}
		err = addSlot(os.Args[2], os.Args[3], mustAtoi(os.Args[4]), mustAtoi(os.Args[5]), mustAtoi(os.Args[6]), record)
	case "unicolor-addslot":
		if len(os.Args) != 5 {
			fmt.Fprintln(os.Stderr, "usage: uniparam unicolor-addslot <in> <out> <team>")
			os.Exit(2)
		}
		err = unicolorAddSlot(os.Args[2], os.Args[3], mustAtoi(os.Args[4]))
	case "inflate":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: uniparam inflate <in> <out>")
			os.Exit(2)
		}
		var env *envelope
		env, err = readEnvelope(os.Args[2])
		if err == nil {
			err = os.WriteFile(os.Args[3], env.body, 0o644)
		}
	case "retex":
		if len(os.Args) != 6 {
			fmt.Fprintln(os.Stderr, "usage: uniparam retex <src> <out> <fromTex> <toTex>")
			os.Exit(2)
		}
		err = retex(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
