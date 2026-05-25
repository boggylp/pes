---
name: pes-kit-ftex
description: '**Invoke this skill BEFORE converting a kit PNG to a PES `.ftex` file or producing a kitserver-named texture.** Covers `tools/kit-to-ftex.ps1` (the PNG -> DDS (DXT5, via ImageMagick) -> FTEX (via Atvaark FtexTool 0.3.3) pipeline), the kitserver naming convention `u<team_id><p|g><slot>.ftex`, team-name resolution via `FL26_teams.txt` substring match, the live install destination at `C:\Program Files (x86)\SP Football Life 2026\SiderAddons\livecpk\root\...`, the DXT5-vs-PixelFormatType-11 caveat (DXT5 is broadly compatible but may not exactly match the engine-preferred format), and the kit-install scope rule for SYSTEM cache (don't propose the delete for kits — it's not warranted; see [[feedback_system_cache_scope]] for the universal propose-and-ask rule). Triggers: "convert this kit", "make an FTEX for <team>", "kit from pesmaster.com", "produce u<id>p<slot>.ftex", "install this kit texture".'
metadata:
  trusted_sources:
    - https://github.com/Atvaark/FtexTool
---

# PES kit PNG to FTEX

Wraps `tools/kit-to-ftex.ps1`. PNG -> DDS (DXT5, via ImageMagick `magick`) -> FTEX (via Atvaark's FtexTool 0.3.3, auto-fetched into `tools/bin/`).

## When to use

- User has a PNG kit (typically exported from `pesmaster.com/kit-creator`) and wants the kitserver-ready FTEX
- Producing the texture for a specific team and slot

## Steps

1. **Confirm prerequisites.** `magick` (ImageMagick) on `PATH`. FtexTool auto-fetches on first run to `tools/bin/FtexTool-v0.3.3/` (gitignored).

2. **Pick the invocation that matches the request.**

    Plain conversion next to the input:

    ```powershell
    pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png
    ```

    Explicit output path:

    ```powershell
    pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png path\to\out\u.ftex
    ```

    Kitserver naming by team ID + slot (player kit):

    ```powershell
    pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -Slot 3
    # produces u2525p3.ftex
    ```

    Kitserver naming by team-name substring match against `FL26_teams.txt` (default at install root; override with `-TeamsFile`):

    ```powershell
    pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamName Hajduk -Slot 3 -OutDir out\
    ```

    Goalkeeper kit:

    ```powershell
    pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -KitType g -Slot 1
    # produces u2525g1.ftex
    ```

3. **Drop the FTEX into the kitserver slot folder** under the live install. Slot folder scaffolding (`config.txt`, `order.ini`, `map.txt`) and partial-texture files (`_back`, `_leg`, `_name`) are **out of scope** for this skill.

4. **Verify in-game.** DXT5 (FTEX PixelFormatType 4) is broadly compatible but not necessarily what every stock kit slot uses; some use PixelFormatType 11 (unknown / likely BC7). Visual confirmation is the verification step. State this explicitly when reporting the conversion as done.

## Parameters

| Flag             | Effect                                                                                          |
| ---------------- | ----------------------------------------------------------------------------------------------- |
| `-TeamId <int>`  | Kitserver naming with literal team ID                                                            |
| `-TeamName <s>`  | Substring match in `FL26_teams.txt` (default `%ProgramFiles(x86)%\SP Football Life 2026\FL26_teams.txt`) |
| `-TeamsFile <p>` | Override the teams file path                                                                     |
| `-Slot <1-9>`    | Kit slot (default `1`)                                                                           |
| `-KitType p\|g`  | Player or goalkeeper (default `p`)                                                               |
| `-OutDir <path>` | Destination folder (default: next to the input PNG)                                              |

## Pitfalls

- **SYSTEM cache delete is not warranted for kits.** Kits load via kitserver / livecpk, not via the `SYSTEM00000000` cache gate. Don't propose the delete for a kit install. (Propose-and-ask is the universal rule per [[feedback_system_cache_scope]]; here the answer should be don't propose it.)
- **DXT5 vs PixelFormatType 11 caveat.** The tool produces DXT5. If a specific slot stutters or renders wrong, the original may have been PixelFormatType 11 (likely BC7). Verify visually; the tool does not match Konami's per-slot preferred format.
- **Team-name match is a substring, not exact.** `Hajduk` will match `Hajduk Split` and `HNK Hajduk` if both exist; check `FL26_teams.txt` before relying on the auto-name.
- **Slot scaffolding is the user's job.** This tool produces the texture only; if the slot folder doesn't already have its `config.txt` / `order.ini` / `map.txt`, dropping in the FTEX alone may not be enough.
- **Verify the live install path** before writing to a hard-coded location. Live destination is `C:\Program Files (x86)\SP Football Life 2026\SiderAddons\livecpk\root\...` per [[reference_fl26_install]].

## Out of scope

- Kitserver slot folder creation, `config.txt` / `order.ini` / `map.txt` editing
- Multi-file kit packs (`_back`, `_leg`, `_name`)
- Producing non-DXT5 FTEX formats
- Any work on player faces: see [[pes-faces-install]]
