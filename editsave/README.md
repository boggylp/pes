# EDIT save

Wrap `decrypter21.exe` and `encrypter21.exe` for PES 2021 and Football Life 2026 `EDIT00000000` saves. The tool normalizes paths, verifies decrypted output, and protects output files.

`decrypt`, `encrypt`, and `roundtrip` manage the decrypt and encrypt boundary. `apply-tactics` and `reset-squad-order` modify decrypted `data.dat` before encryption.

## Build

```sh
cd editsave
go build .
```

## Tools

Put non-empty `decrypter21.exe` and `encrypter21.exe` files in one directory. Pass that directory with `--tools-dir`.

## Commands

Decrypt a save:

```sh
editsave decrypt --tools-dir DIR --out DIR [--force] INPUT
```

Encrypt a decrypted directory:

```sh
editsave encrypt --tools-dir DIR --out FILE [--force] [--allow-live-name] INPUT_DIR
```

Verify a decrypt and encrypt round trip:

```sh
editsave roundtrip --tools-dir DIR [--keep-work] [--require-strict] INPUT
```

Apply per-team tactics from `id_list.txt`:

```sh
editsave apply-tactics \
  --tools-dir DIR \
  --tactics-dir DIR \
  --id-list FILE \
  --out FILE \
  [--force] [--allow-live-name] \
  INPUT
```

Reset each team's 32-byte squad-order array to identity after tactics import:

```sh
editsave reset-squad-order \
  --tools-dir DIR \
  --out FILE \
  [--force] [--allow-live-name] \
  INPUT
```

## Output protection

- `decrypt` rejects a non-empty output directory unless `--force` is set.
- Commands that write a file reject an existing output unless `--force` is set.
- An output named `EDIT00000000` also requires `--allow-live-name`.
- `--force` alone does not permit the live-save filename.
- The tool converts paths to absolute paths before it starts the wrapped Windows tools.
- Decryption must produce all expected files and a non-empty `data.dat`.

## Round-trip result

- Strict PASS means that input ciphertext equals the re-encryption.
- Content PASS means that the decrypted `data.dat` files are equal.
- The default exits with code 0 only when content passes.
- `--require-strict` also requires strict PASS.
- Status and diagnostics go to standard error. Standard output is reserved for machine-readable output.

## Tactics format

A `.PES2021_tactics` file is 628 bytes. Bytes 0 to 3 contain the little-endian team ID. Byte 4 is `0x00`. Bytes 5 and 6 contain the formation variant.

The tactics section in decrypted `data.dat` contains contiguous 628-byte slots. The scanner finds the section and walks it by slot size. It requires byte 4 to be `0x00` and the team ID to be plausible.

Each `id_list.txt` line has this form:

```text
<target-team-id>, <source-filename> # optional comment
```

Blank lines and lines that start with `#` are ignored. The tool changes the source file's internal team ID to the target team ID before it writes the slot.
