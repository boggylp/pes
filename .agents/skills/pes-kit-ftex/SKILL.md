---
name: pes-kit-ftex
description: 'Invoke BEFORE converting a kit PNG to a PES `.ftex` file or producing a kitserver-named texture. Covers `tools/kit-to-ftex.ps1` (PNG -> DDS -> FTEX), the kitserver naming convention `u<team_id><p|g><slot>.ftex`, team-name resolution via `FL26_teams.txt`, the live-install destination, and the DXT5-vs-PixelFormatType-11 caveat. Triggers: "convert this kit", "make an FTEX for <team>", "kit from pesmaster.com", "produce u<id>p<slot>.ftex", "install this kit texture".'
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

| Flag | Effect |
| --- | --- |
| `-TeamId <int>` | Kitserver naming with literal team ID |
| `-TeamName <s>` | Substring match in `FL26_teams.txt` at the install root (override with `-TeamsFile`; install path varies per machine) |
| `-TeamsFile <p>` | Override the teams file path |
| `-Slot <1-9>` | Kit slot (default `1`) |
| `-KitType p\|g` | Player or goalkeeper (default `p`) |
| `-OutDir <path>` | Destination folder (default: next to the input PNG) |

## Pitfalls

- **Kits load via kitserver / livecpk; `SYSTEM00000000` is not the gate.** Its delete is hard-banned anyway (repo `AGENTS.md`).
- **DXT5 vs PixelFormatType 11 caveat.** The tool produces DXT5. If a specific slot stutters or renders wrong, the original may have been PixelFormatType 11 (likely BC7). Verify visually; the tool does not match Konami's per-slot preferred format.
- **Team-name match is a substring, not exact.** `Hajduk` will match `Hajduk Split` and `HNK Hajduk` if both exist; check `FL26_teams.txt` before relying on the auto-name.
- **Slot scaffolding is the user's job.** This tool produces the texture only; if the slot folder doesn't already have its `config.txt` / `order.ini` / `map.txt`, dropping in the FTEX alone may not be enough.
- **Verify the live install path** before writing; it varies per machine and the user-given path wins. Kitserver kits live under `<install>\sider*\content\kit-server\...`; a raw livecpk override would be `<install>\SiderAddons\livecpk\root\...`.

## Extracting existing kits from a PES cpk (repeatable)

Any patch's kits live in its **uniform archive** (PES2021 packaging: `dt34_g4.cpk`; confirm for the patch at hand with `cpk list <cpk> | grep -c uniform`), in two parts:

- **Textures:** `Asset/model/character/uniform/texture/#windx11/u<NNNN><slot>[_back|_leg|_name|_name_ex].ftex` — `NNNN` is the team ID zero-padded to 4 digits (272 -> `u0272`); slot is **`p1..p4` (player)** or **`g1..g3` (goalkeeper)**. These names match the kitserver `config.txt` `KitFile` value exactly, so no rename on import. (Watch the trap: `g` is goalkeeper, not "graphics" — extract both `p` and `g`, or you get GK kits only.)
- **Definitions:** `common/character0/model/character/uniform/team/<id>/<id>_DEF_{1st,2nd,3rd,GK1st,...}_realUni.bin` — colours / slot setup; reference when writing the kitserver `config.txt`.

Map team IDs to leagues via the sider **kit-server map** (`<install>/sider/content/kit-server/map.txt`, lines `team-id, "League\Team"`) or a team CSV.

Extract a subset repeatably with `cpk extract --prefix` (one pass; exits non-zero if a prefix matches nothing). Per team `<id>` (`PAD` = `%04d`); the `u<PAD>` texture prefix catches both player and GK slots:

```sh
cpk extract --prefix "common/character0/model/character/uniform/team/<id>/"  --out <dir> <uniform.cpk>
cpk extract --prefix "Asset/model/character/uniform/texture/#windx11/u<PAD>" --out <dir> <uniform.cpk>
```

Verify each extracted `.ftex` begins with `FTEX`.

### Kitserver folder layout (how a kit becomes self-contained)

A team's kitserver content is `<install>/sider*/content/kit-server/<League>/<Team>/` with per-slot subfolders `p1..p4` (player) + `g1..g3` (GK), each holding a `config.txt`, plus `order.ini` / `gk_order.ini`. `KitServer.lua` loads the texture as `<slot>/<KitFile>.ftex` (e.g. `p1/u0272p1.ftex`). A patch that keeps its textures in the cpk (like BPB) ships **config-only** slot folders and falls back to the native cpk texture; to move that kit to a different install, drop the extracted `u<id><slot>*.ftex` into the slot folder so it is self-contained. `map.txt` maps `team-id, "League\Team"`; the kit only applies if the destination game actually has that team ID.

## Out of scope

- Kitserver slot folder creation, `config.txt` / `order.ini` / `map.txt` editing
- Multi-file kit packs (`_back`, `_leg`, `_name`)
- Producing non-DXT5 FTEX formats
- Any work on player faces: see pes-faces-install
