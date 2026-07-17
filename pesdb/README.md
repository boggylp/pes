# pesdb

PES 2021 player roster extractor. Parses a `Player.bin` previously extracted from a cpk and emits a `Id;Name;Shirt` CSV. Optionally merges a decrypted EDIT save's `data.dat` so manual roster edits override the base table.

## Build

```sh
cd pesdb
go build .
```

## Usage

```sh
# 1. Extract Player.bin from the game's cpk (uses repo's cpk tool)
cd cpk && go build .
./cpk extract --file common/etc/pesdb/Player.bin \
  --out /tmp/Player.bin \
  "/d/SteamLibrary/steamapps/common/eFootball PES 2021/Data/dt10_x64.cpk"

# 2. Optional: decrypt the EDIT save (using ejogc327's decrypter21.exe)
"/d/apps/Pes 2020 Editor V0.12.10 by Ejogc327/Lib/decrypter21.exe" \
  "$USERPROFILE/Documents/KONAMI/eFootball PES 2021 SEASON UPDATE/2026/save/EDIT00000000" \
  /tmp/edit-out

# 3. Build the roster CSV
cd pesdb
./pesdb roster --player-bin /tmp/Player.bin --out roster.csv
# or with EDIT save merged
./pesdb roster --player-bin /tmp/Player.bin --edit /tmp/edit-out/data.dat --out roster.csv
```

## Format reference

Konami's pesdb files inside cpks are not encrypted — they're a thin envelope around a zlib stream:

```text
+0000  ?  unknown front section (~2 KB; index/hashes — not yet decoded)
+????  8  magic                  \xff\x10\x81WESYS
+????  8  size metadata          (uncompressed-ish + payload-ish; not strictly used)
+????  …  zlib stream            (no end marker — read with stream decoder)
```

After zlib decompression, the payload is a sequence of fixed-size player records. PES 2021's `Player.bin` uses a 312-byte stride from the start of the decompressed buffer:

| Offset  | Type                  | Field       |
| ------- | --------------------- | ----------- |
| `+0x08` | uint32 LE             | Player ID   |
| `+0x44` | UTF-8, 32 B, NUL-term | Player name |
| `+0x81` | UTF-8, 16 B, NUL-term | Shirt name  |

Per-player stats sit between the ID and the name; this tool does not decode them.

The PES 2021 EDIT save (`EDIT00000000`, decrypted with ejogc327's `decrypter21.exe`) uses the same 312-byte stride but a different field layout in `data.dat`:

| Offset  | Type                  | Field                                  |
| ------- | --------------------- | -------------------------------------- |
| `+0x0C` | uint32 LE             | Player ID (also duplicated at `+0x10`) |
| `+0x14` | uint16 LE             | Country code                           |
| `+0x42` | UTF-8, 32 B, NUL-term | Player name                            |
| `+0x7F` | UTF-8, 16 B, NUL-term | Shirt name                             |

Player records start at file offset 112 (80-byte file header + 32-byte section header). The section ends at the first slot whose ID is zero, sentinel (`0xFFFFFFFF` / `0xFFFF6000`), or where the duplicate-ID check fails — after that `data.dat` continues with team/stadium/coach sections in different formats.

## Why this exists

The extracted bytes look opaque if you only inspect the front 2 KB; the WESYS magic and zlib stream sit further into the file. This tool documents and automates the unwrap so patch comparisons stay one command away.
