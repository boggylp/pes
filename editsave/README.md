# editsave

Headless wrapper around `decrypter21.exe` and `encrypter21.exe` (zlac, 2020) for Konami PES 2021 / SP Football Life 2026 `EDIT00000000` saves. Orchestrates the two binaries, normalizes paths, verifies output structure, and refuses to overwrite live game state without an explicit `--force`. The tool never modifies save contents — it owns only the decrypt → modify → encrypt boundary so callers can edit the decrypted `data.dat` in between.

## Build

```sh
cd editsave
go build .
```

## Layout

The wrapped tools live in zlac's tactics importer bundle. On `BogambeDesktop`:

```text
$HOME/MEGA/gaming/pes/tactics/klashman-2025-26-jan/PC PES 2021 ... FL26 EDITION BETA/tools/
├── decrypter21.exe
└── encrypter21.exe
```

Both tools must exist as non-zero `.exe` files in the same directory.

## Commands

```sh
# Decrypt a save. Refuses a non-empty --out unless --force.
editsave decrypt --tools-dir DIR --out DIR [--force] INPUT

# Encrypt a previously-decrypted directory back into an EDIT save.
# Refuses to overwrite --out unless --force, so a typo cannot
# silently destroy the live game save.
editsave encrypt --tools-dir DIR --out FILE [--force] INPUT_DIR

# Decrypt → re-encrypt → re-decrypt and report whether the round-trip
# is byte-perfect (strict) and / or content-perfect (data.dat equal).
# Exit 0 iff content matches; with --require-strict, also require strict.
editsave roundtrip --tools-dir DIR [--keep-work] [--require-strict] INPUT

# Apply Klashman/Zlac per-team tactics into a save. Reads id_list.txt
# (target_team_id, source_filename), writes each source file's payload
# into the matching team's 628-byte tactics slot, then re-encrypts.
# Source files' internal team_id is overwritten with the target team_id
# so the slot header stays consistent.
editsave apply-tactics \
    --tools-dir DIR --tactics-dir DIR --id-list FILE \
    --out FILE [--force] INPUT
```

## Roundtrip semantics

- **strict PASS**: ciphertext of the input equals the re-encryption. Decrypter and encrypter are a perfect inverse pair on this input.
- **content PASS**: decrypted `data.dat` of the input equals the decrypted `data.dat` of the re-encryption. The save is functionally reconstructible even if ciphertext drifts.

Default exit policy: 0 iff content is PASS. The wrapped tools are empirically a perfect inverse pair (strict PASS on `BogambeDesktop`); the looser default only covers a future tool revision drifting ciphertext. Pass `--require-strict` for tighter scripted guarantees.

## Status streams

All status and diagnostics go to stderr. stdout is reserved for future machine-readable output (none today).

## apply-tactics format notes

`.PES2021_tactics` files are 628 bytes. Byte 0..3 is `team_id` LE uint32; byte 4 is always `0x00`; bytes 5..6 encode the formation variant and vary per team (e.g. England `01 01`, Czech Republic `01 03`, Real Madrid `04 01`). The on-disk team-tactics section in `data.dat` is a contiguous array of 628-byte slots starting at the same byte layout. The section start was empirically `0x00a09880` on the FL26 v2.2 save and contained 749 slots ending at `0x00a7c5e4`.

The slot-finder anchors on the strict 3-byte pattern `00 01 01` (the FL26 default at byte 6), then walks the section in 628-byte strides using only `byte[4] == 0x00` plus a plausible team_id, which keeps working after tactics imports have shifted byte 6 per team.

`id_list.txt` syntax: one `target_team_id, source_filename [# comment]` per line. Blank lines and `#` lines are ignored. The source filename's internal team_id may differ from the target — common when reusing tactics across PES versions where the same real team has a different internal id.

## Safety

- Refuses to overwrite an existing `--out` file (`encrypt`) or non-empty `--out` directory (`decrypt`) without `--force`, so a typo cannot destroy the live save.
- All paths are normalized to absolute form before invoking the wrapped tools; `decrypter21.exe` is a 2020-era 32-bit binary sensitive to relative paths and trailing separators on Windows.
- `verifyDecryptedDir` checks every expected file is present AND that `data.dat` is non-empty, so a botched decrypt cannot pass into re-encryption and destroy the save.
