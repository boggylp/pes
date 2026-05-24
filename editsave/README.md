# editsave

Headless wrapper around `decrypter21.exe` and `encrypter21.exe` (zlac, 2020) for
Konami PES 2021 / SP Football Life 2026 `EDIT00000000` saves. Orchestrates the
two binaries, normalizes paths, verifies output structure, and refuses to
overwrite live game state without an explicit `--force`. The tool never modifies
save contents — it owns only the decrypt → modify → encrypt boundary so callers
can edit the decrypted `data.dat` in between.

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
```

## Roundtrip semantics

- **strict PASS**: ciphertext of the input equals the re-encryption.
  Decrypter and encrypter are a perfect inverse pair on this input.
- **content PASS**: decrypted `data.dat` of the input equals the decrypted
  `data.dat` of the re-encryption. The save is functionally reconstructible
  even if ciphertext drifts.

Default exit policy: 0 iff content is PASS. On `BogambeDesktop` the wrapped
tools are empirically a perfect inverse pair (strict is PASS), so the looser
default exists for the case where a future tool revision drifts ciphertext;
it is not the observed behavior today. Pass `--require-strict` for tighter
scripted guarantees.

## Status streams

All status and diagnostics go to stderr. stdout is reserved for future
machine-readable output (none today).

## Safety

- Refuses to overwrite an existing `--out` file (`encrypt`) or non-empty
  `--out` directory (`decrypt`) without `--force`. The live FL26 save lives at
  `C:\Program Files (x86)\SP Football Life 2026\dataSP\EDIT00000000`; the
  refuse-overwrite default exists so a typo cannot destroy it.
- All paths are normalized to absolute form before invoking the wrapped tools;
  `decrypter21.exe` is a 2020-era 32-bit binary sensitive to relative paths and
  trailing separators on Windows.
- `verifyDecryptedDir` checks every expected file is present AND that
  `data.dat` is non-empty, so a botched decrypt cannot pass into re-encryption
  and destroy the save.
