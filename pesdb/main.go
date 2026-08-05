// pesdb extracts a PES 2021 player roster (Id;Name;Shirt) from a Player.bin
// previously extracted from a cpk with the repo's `cpk extract` command.
//
// Pipeline:
//
//  1. Find Konami's pesdb container magic (\xff\x10\x81WESYS) inside the
//     Player.bin and skip the 16-byte container header.
//  2. zlib-decompress the payload — yields fixed 312-byte player records.
//  3. Per record: Id (uint32 LE @ +8), Name (UTF-8 @ +0x44, NUL-term, 32 B),
//     Shirt (NUL-term @ +0x81, 16 B).
//  4. Optionally merge an EDIT save's data.dat (decrypter21.exe output):
//     same 312-byte stride but Id @ +12, name @ +0x42, shirt @ +0x7F. EDIT
//     entries override base entries on the same Id.
//
// pesdb files inside cpks are NOT encrypted, just packed in a Konami WESYS
// envelope around a zlib stream. The "Player.bin is encrypted" myth came from
// reading only the front 2 KB (which is opaque) before reaching the WESYS
// magic at offset ~2048.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	recordStride = 312

	// pesdb Player.bin record offsets.
	playerIDOff    = 8
	playerNameOff  = 0x44
	playerShirtOff = 0x81

	// EDIT save data.dat record offsets.
	editIDOff    = 12
	editNameOff  = 0x42
	editShirtOff = 0x7F

	// EDIT save layout: 80-byte file header + 32-byte section header.
	editRecordsStart = 112
)

var wesysMagic = []byte{0xff, 0x10, 0x81, 'W', 'E', 'S', 'Y', 'S'}

type player struct {
	name  string
	shirt string
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "roster":
		if err := runRoster(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "pesdb:", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "pesdb: unknown subcommand %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `pesdb — parse PES 2021 player rosters from extracted pesdb files

Usage:
  pesdb roster --player-bin <Player.bin> --out <csv> [--edit <data.dat>]

The Player.bin must come from the repo's cpk tool:
  cpk extract --file common/etc/pesdb/Player.bin --out Player.bin <some.cpk>

The optional --edit data.dat is the decrypted output of ejogc327's
decrypter21.exe applied to a EDIT00000000 save file.
`)
}

func runRoster(args []string) error {
	fs := flag.NewFlagSet("roster", flag.ContinueOnError)
	playerBin := fs.String("player-bin", "", "Player.bin extracted from a cpk")
	outPath := fs.String("out", "", "output CSV path")
	editPath := fs.String("edit", "", "optional decrypted EDIT save data.dat to merge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *playerBin == "" || *outPath == "" {
		fs.Usage()
		return errors.New("--player-bin and --out are required")
	}

	base, err := parsePlayerBin(*playerBin)
	if err != nil {
		return fmt.Errorf("parse player.bin: %w", err)
	}
	fmt.Fprintf(os.Stderr, "base players: %d\n", len(base))

	combined := make(map[uint32]player, len(base))
	for k, v := range base {
		combined[k] = v
	}
	if *editPath != "" {
		edits, err := parseEditData(*editPath)
		if err != nil {
			return fmt.Errorf("parse edit save: %w", err)
		}
		overlap := 0
		for k := range edits {
			if _, ok := base[k]; ok {
				overlap++
			}
		}
		for k, v := range edits {
			combined[k] = v
		}
		fmt.Fprintf(os.Stderr, "EDIT-save adds: %d (overlap with base: %d)\n", len(edits), overlap)
	}

	if err := writeCSV(*outPath, combined); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %d rows -> %s\n", len(combined), *outPath)
	return nil
}

func parsePlayerBin(path string) (map[uint32]player, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	off := bytes.Index(data, wesysMagic)
	if off < 0 {
		return nil, errors.New("WESYS magic not found")
	}
	zlibStart := off + 16
	zr, err := zlib.NewReader(bytes.NewReader(data[zlibStart:]))
	if err != nil {
		return nil, fmt.Errorf("zlib reader: %w", err)
	}
	defer func() { _ = zr.Close() }()
	plain, err := io.ReadAll(zr)
	// PES pesdb zlib streams have no end-marker; an unexpected-EOF error
	// after we have substantial output is normal — accept it as long as we
	// got at least one record's worth of data.
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("zlib decompress: %w", err)
	}
	if len(plain) < recordStride {
		return nil, fmt.Errorf("decompressed payload too small (%d bytes)", len(plain))
	}
	return parseRecords(plain, 0, recordsAll, playerIDOff, playerNameOff, playerShirtOff), nil
}

func parseEditData(path string) (map[uint32]player, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < editRecordsStart+recordStride {
		return nil, fmt.Errorf("edit data too small (%d bytes)", len(data))
	}
	return parseRecords(data, editRecordsStart, recordsUntilSentinel, editIDOff, editNameOff, editShirtOff), nil
}

const (
	recordsAll int = iota
	recordsUntilSentinel
)

func parseRecords(buf []byte, start int, mode int, idOff, nameOff, shirtOff int) map[uint32]player {
	out := map[uint32]player{}
	for i := start; i+recordStride <= len(buf); i += recordStride {
		rec := buf[i : i+recordStride]
		pid := binary.LittleEndian.Uint32(rec[idOff:])
		if mode == recordsUntilSentinel {
			// EDIT save: stop at first invalid record so we don't run into
			// the team/stadium/coach sections that follow players.
			pid2 := binary.LittleEndian.Uint32(rec[idOff+4:])
			if pid == 0 || pid == 0xFFFFFFFF || pid != pid2 || pid > 10_000_000 {
				break
			}
		} else {
			if pid == 0 || pid > 10_000_000 {
				continue
			}
		}
		name := readCString(rec, nameOff, 32)
		if name == "" {
			continue
		}
		shirt := readCString(rec, shirtOff, 16)
		out[pid] = player{name: name, shirt: shirt}
	}
	return out
}

func readCString(rec []byte, off, max int) string {
	end := off + max
	if end > len(rec) {
		end = len(rec)
	}
	if i := bytes.IndexByte(rec[off:end], 0); i >= 0 {
		return strings.ToValidUTF8(string(rec[off:off+i]), "?")
	}
	return strings.ToValidUTF8(string(rec[off:end]), "?")
}

func writeCSV(path string, players map[uint32]player) error {
	if err := os.MkdirAll(parentDir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(f)
	w.Comma = ';'
	if err := w.Write([]string{"Id", "Name", "Shirt"}); err != nil {
		return err
	}
	ids := make([]uint32, 0, len(players))
	for id := range players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		p := players[id]
		if err := w.Write([]string{strconv.FormatUint(uint64(id), 10), p.name, p.shirt}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
