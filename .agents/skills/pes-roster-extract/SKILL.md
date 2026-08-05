---
name: pes-roster-extract
description: 'Extract or compare a PES or Football Life player roster when the answer requires Player.bin or EDIT-save evidence.'
metadata:
  trusted_sources:
    - https://github.com/cri-mw/cpk
---

# PES roster extract

## Steps

1. Identify the effective `Player.bin`.

   - Check active livecpk roots first. Use a livecpk `Player.bin` when it overrides the archive path.
   - Otherwise, inspect CPK load order. Check `Data/download/dt80_*E_x64.cpk`, then `Data/dt10_x64.cpk`, then `Data/dt00_x64.cpk`.

2. When no livecpk override exists, extract the highest-priority CPK entry.

   ```sh
   cd cpk
   go build .
   ./cpk extract \
     --file common/etc/pesdb/Player.bin \
     --out /tmp/Player.bin \
     "$INSTALL/Data/dt10_x64.cpk"
   ```

3. Verify that the extracted file contains the `\xff\x10\x81WESYS` magic.

4. To include EDIT-save additions, decrypt the live EDIT save with the configured `decrypter21.exe`. Use the resulting `data.dat` only as input.

5. Build the roster CSV.

   ```sh
   cd pesdb
   go build .
   ./pesdb roster --player-bin /tmp/Player.bin --out /tmp/roster.csv
   ./pesdb roster --player-bin /tmp/Player.bin --edit /tmp/edit-out/data.dat --out /tmp/roster.csv
   ```

   The output columns are `Id;Name;Shirt`.

6. Verify the row count and inspect known players from the target install. A large mismatch usually indicates the wrong source or load order.

7. For a comparison, generate each CSV from its own effective source and compare normalized IDs and names.

## Record layout

The WESYS envelope contains a zlib stream. Decompressed player records use a 312-byte stride.

| Source | ID | Name | Shirt |
| --- | --- | --- | --- |
| `Player.bin` | `+0x08` | `+0x44` | `+0x81` |
| EDIT `data.dat` | `+0x0C` | `+0x42` | `+0x7F` |

EDIT player records start at offset 112. The EDIT data contains edited or created players, not the complete base roster.

## Review

- Use the effective live database for face and player-ID work.
- Generate a roster for each install. Do not reuse a roster from another patch.
- Treat the archive roster as incomplete when an active livecpk database overrides it.
- State when EDIT additions could not be merged.
- Treat a missing WESYS magic as an extraction or source error.

## Boundaries

- The `pesdb` tool is read-only.
- Use `pes-faces-install` to install faces from the CSV.
- Use `editsave/README.md` for EDIT-save decryption and encryption.
