---
name: pes-kit-ftex
description: 'Convert a kit PNG to a PES FTEX texture when the user asks for a kitserver-ready file.'
metadata:
  trusted_sources:
    - https://github.com/Atvaark/FtexTool
---

# PES kit FTEX

## Steps

1. Confirm that ImageMagick `magick` is on `PATH`. The script downloads FtexTool v0.4.0 to `tools/bin/FtexTool-v0.4.0/` on first use.

2. Select the output form.

   ```powershell
   pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png
   pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png path\to\out\u.ftex
   pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -Slot 3
   pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamName Hajduk -Slot 3 -OutDir out\
   pwsh -File .\tools\kit-to-ftex.ps1 path\to\kit.png -TeamId 2525 -KitType g -Slot 1
   ```

3. Check every team-name match before conversion. `-TeamName` uses substring matching and can match more than one team.

4. Confirm the user-provided live-install path before writing to kitserver.

5. Put the generated FTEX in the existing kitserver slot folder.

6. Verify the texture in the game. The script produces DXT5, which is FTEX `PixelFormatType 4`. Some stock textures use another pixel format.

## Parameters

| Flag | Effect |
| --- | --- |
| `-TeamId <int>` | Name the output for a literal team ID. |
| `-TeamName <text>` | Match a team name in `-TeamsFile`. |
| `-TeamsFile <path>` | Select the team list. The default is the installed Football Life 2026 `FL26_teams.txt`. |
| `-Slot <1-9>` | Select the kit slot. The default is `1`. |
| `-KitType p\|g` | Select a player or goalkeeper kit. The default is `p`. |
| `-OutDir <path>` | Select the destination directory. The default is the input PNG directory. |

## Review

- Confirm the output name is `u<team-id><p|g><slot>.ftex` when kitserver naming is requested.
- Keep existing `config.txt`, `order.ini`, and `map.txt` files unchanged.
- Report that visual verification is required because pixel formats can differ.

## Boundaries

- Use `pes-kit-extract` to extract or migrate existing kits from a CPK.
- This skill creates one DXT5 FTEX texture.
- It does not create slot configuration or partial textures such as `_back`, `_leg`, or `_name`.
- Use `pes-faces-set` for player faces.
