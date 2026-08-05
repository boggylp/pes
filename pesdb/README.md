# PESDB

Extract a PES 2021 player roster from `Player.bin`. Optionally merge player changes from a decrypted EDIT save. The output columns are `Id;Name;Shirt`.

## Build

```sh
cd pesdb
go build .
```

## Usage

Extract the effective `Player.bin` with the CPK tool when no livecpk database overrides it:

```sh
./cpk/cpk extract \
  --file common/etc/pesdb/Player.bin \
  --out /tmp/Player.bin \
  "/path/to/Data/dt10_x64.cpk"
```

Optionally decrypt the EDIT save with the configured `decrypter21.exe`.

Build the roster:

```sh
./pesdb/pesdb roster --player-bin /tmp/Player.bin --out roster.csv
./pesdb/pesdb roster --player-bin /tmp/Player.bin --edit /tmp/edit-out/data.dat --out roster.csv
```

## Player.bin format

PESDB files use a WESYS envelope around a zlib stream. The parser finds the `\xff\x10\x81WESYS` magic and skips the 16-byte envelope header. It does not use the two size fields.

After decompression, PES 2021 player records use a 312-byte stride.

| Offset | Type | Field |
| --- | --- | --- |
| `+0x08` | little-endian uint32 | Player ID |
| `+0x44` | 32-byte, NUL-terminated UTF-8 | Player name |
| `+0x81` | 16-byte, NUL-terminated UTF-8 | Shirt name |

The tool does not decode the player statistics between the ID and name fields.

## EDIT format

A decrypted PES 2021 EDIT save uses a 312-byte player stride with different offsets.

| Offset | Type | Field |
| --- | --- | --- |
| `+0x0C` | little-endian uint32 | Player ID |
| `+0x10` | little-endian uint32 | Duplicate player ID |
| `+0x14` | little-endian uint16 | Country code |
| `+0x42` | 32-byte, NUL-terminated UTF-8 | Player name |
| `+0x7F` | 16-byte, NUL-terminated UTF-8 | Shirt name |

Player records start at offset 112 after the file and section headers. Parsing stops at the first zero ID, sentinel ID, invalid duplicate ID, or ID above the supported range. Other EDIT sections follow the player section.
