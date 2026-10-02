---
name: pes-roster-extract
description: 'Extract or compare a PES or Football Life player roster when the answer requires Player.bin or EDIT-save evidence.'
---

# PES roster extract

Build each install's roster from its own effective source.

## Steps

1. Find the effective `Player.bin`.
   - An active livecpk `Player.bin` wins.
   - Otherwise take the first CPK in load order of `Data/download/dt80_*E_x64.cpk`, `Data/dt10_x64.cpk`, and `Data/dt00_x64.cpk`.
2. Without a livecpk override, extract it:

   ```sh
   cd cpk && go build . && ./cpk extract --file common/etc/pesdb/Player.bin --out /tmp/Player.bin "$INSTALL/Data/dt10_x64.cpk"
   ```

3. Confirm the file starts with the `\xff\x10\x81WESYS` magic, and treat a missing magic as a wrong source.
4. To include EDIT-save players, decrypt the live EDIT save with the configured `decrypter21.exe`, per `editsave/README.md`.
5. Build the `Id;Name;Shirt` CSV:

   ```sh
   cd pesdb && go build .
   ./pesdb roster --player-bin /tmp/Player.bin [--edit /tmp/edit-out/data.dat] --out /tmp/roster.csv
   ```

   - When EDIT players could not be merged, say so.
6. Check the row count and a few known players, because a large mismatch points at the wrong source or load order.
7. To compare installs, diff their normalized IDs and names.

## Record layout

The WESYS envelope holds a zlib stream of 312-byte player records.

| Source | ID | Name | Shirt |
| --- | --- | --- | --- |
| `Player.bin` | `+0x08` | `+0x44` | `+0x81` |
| EDIT `data.dat` | `+0x0C` | `+0x42` | `+0x7F` |

EDIT player records start at offset 112 and hold only edited or created players.
