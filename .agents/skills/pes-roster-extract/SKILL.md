---
name: pes-roster-extract
description: '**Invoke this skill BEFORE extracting a player roster from a PES / Football Life / BPB install, comparing rosters across versions, or answering any question about which players ship in a given CPK.** Covers `cpk extract --file common/etc/pesdb/Player.bin <some.cpk>` to pull the base, optional `decrypter21.exe` against `EDIT00000000` for custom adds, and `pesdb roster --player-bin Player.bin [--edit data.dat] --out roster.csv` to produce the `Id;Name;Shirt` CSV. Covers load order (`dt10_x64.cpk` overrides `dt00_x64.cpk`; `dt80_*E_x64.cpk` overrides both), the `\xff\x10\x81WESYS` + zlib envelope structure, and the BPB-vs-FL26 distinct-Player.bin fact. Triggers: "extract the player DB", "get the roster from this cpk", "what players ship in BPB?", "what is in Player.bin?", "compare BPB and FL26 rosters".'
metadata:
  trusted_sources:
    - https://github.com/cri-mw/cpk
---

# PES roster extract

Reads `Player.bin` out of a CPK, optionally merges custom adds from a decrypted EDIT save, emits a CSV. Pesdb files are zlib-wrapped, not encrypted.

## When to use

- Cross-version roster comparison (BPB vs FL26)
- Validating face mappings against the actual shipped roster (precondition for [[pes-faces-install]])
- Answering "is player X in this cpk?" with evidence
- Generating the CSV used by the `faces` tool

## Steps

1. **Identify the CPK that owns the effective `Player.bin`.** Load order:

    | CPK                         | Role                              |
    | --------------------------- | --------------------------------- |
    | `Data/dt00_x64.cpk`         | Base                              |
    | `Data/dt10_x64.cpk`         | Overrides `dt00`                  |
    | `Data/download/dt80_*E_x64.cpk` | DLC, overrides both           |

    The highest-priority CPK that contains `common/etc/pesdb/Player.bin` is the one in use. BPB 2026 ships an identical 1,751,422-byte `Player.bin` in both `dt00` and `dt10` (md5 `89938c15...`); FL26 ships its own distinct bytes.

2. **Extract `Player.bin`** with the repo's `cpk` tool:

    ```sh
    cd cpk && go build .
    ./cpk extract \
      --file common/etc/pesdb/Player.bin \
      --out /tmp/Player.bin \
      "/c/Program Files (x86)/SP Football Life 2026/Data/dt10_x64.cpk"
    ```

    Inner-path matching is case-insensitive and slash-agnostic.

3. **(Optional) decrypt the EDIT save** for custom adds. The repo `cpk` tool does **not** parse EDIT saves; use ejogc327's `decrypter21.exe`:

    ```sh
    "/d/apps/Pes 2020 Editor V0.12.10 by Ejogc327/Lib/decrypter21.exe" \
      "$USERPROFILE/Documents/KONAMI/eFootball PES 2021 SEASON UPDATE/2026/save/EDIT00000000" \
      /tmp/edit-out
    ```

    Produces `data.dat` containing the custom-adds records. BPB ships 20 BPB-specific custom adds on top of the archive baseline.

4. **Build the CSV** with `pesdb`:

    ```sh
    cd pesdb && go build .
    # Base only
    ./pesdb roster --player-bin /tmp/Player.bin --out /tmp/roster.csv
    # Base + EDIT save adds merged
    ./pesdb roster --player-bin /tmp/Player.bin --edit /tmp/edit-out/data.dat --out /tmp/roster.csv
    ```

    Output is `Id;Name;Shirt` ready for the `faces` tool.

5. **Verify** the row count against the expected roster size. BPB 2026 with the 20 custom adds gives a known total; FL26 a different known total. A wildly off count usually means the wrong CPK was extracted.

## Record layout (for verification)

Pesdb files are a `\xff\x10\x81WESYS` magic + 8-byte size header + zlib stream, prefixed by ~2 KB of opaque header. After decompression, records are 312 bytes:

| Source       | Offset | Field  |
| ------------ | ------ | ------ |
| `Player.bin` | `+0x08` | Id    |
| `Player.bin` | `+0x44` | Name  |
| `Player.bin` | `+0x81` | Shirt |
| EDIT `data.dat` | `+0x0C` | Id    |
| EDIT `data.dat` | `+0x42` | Name  |
| EDIT `data.dat` | `+0x7F` | Shirt |

EDIT records start at file offset 112.

## Pitfalls

- **Player.bin is not encrypted.** Earlier "encrypted Player.bin" / "zero-filled Player.bin" claims were both wrong. The first read only the opaque front section; the second hit an offset bug in `cpk` that was fixed 2026-05-06. The file is zlib inside a WESYS envelope; the front ~2 KB is opaque metadata, not encryption.
- **Pick the right CPK.** Running `extract` against `dt00` when `dt10` overrides it gives the base roster, not the effective one. Verify with [[pes-gameplay-status]] which CPK actually exists in `Data/`.
- **BPB and FL26 are distinct.** Their `Player.bin` files differ. Do not assume a roster CSV from one is valid against the other; regenerate per install.
- **EDIT save decryption is third-party.** `decrypter21.exe` is the only known working decrypter; if it's not available, the EDIT-save adds cannot be merged and the CSV is base-only.
- **The repo's `cpk` tool falls back to CRILAYLA decompression but PES 2017-2021 cpks usually store uncompressed.** A "compression unknown" error usually means the inner path is wrong, not that the file is encrypted.

## cpk extraction ContentOffset bug (FIXED 2026-06-07)

A `reader.go` bug silently corrupted extracted face packs. `load()` reused one `rowReader` for two `readNamedNumber` calls (`TocOffset` then `ContentOffset`); that walker advances the row pointer across every column, so the second read started past the row end and returned garbage. The bogus ContentOffset (a huge value, then forced down to TocOffset by the `tocOffset < contentOffset` clamp) fed the absolute-offset rebase, so every file was read one `Align` block (2048 B) too early — a 2 KB garbage prefix plus a 2048 B tail truncation.

Symptom: extracted `face.fpk` had `foxfpk` at offset 2048 instead of 0 → **crashes the game** on load. `Player.bin` survived only because `pesdb` searches for the WESYS magic and tolerated the junk prefix.

Fix: use a fresh `headerTable.row(0)` per `readNamedNumber`. Verified on `BogambeDesktop`: fa26_bpb `17003` → `foxfpk@0` size-correct; dt10 `Player.bin` → WESYS@0, 24981 rows.

Quick post-extract sanity check (no python; first 6 bytes of a face must be the magic):

```sh
head -c6 "<dir>/#Win/face.fpk"   # must print: foxfpk
```

## Out of scope

- Editing player records: the `pesdb` tool is read-only.
- Installing faces against the resulting CSV: see [[pes-faces-install]].
- Decrypting the EDIT save itself; that needs `decrypter21.exe` from ejogc327.
