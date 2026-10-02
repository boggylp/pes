---
name: pes-kit-ftex
description: 'Convert a kit PNG to a PES FTEX texture when the user asks for a kitserver-ready file.'
metadata:
  trusted_sources:
    - https://github.com/Atvaark/FtexTool
---

# PES kit FTEX

## Steps

1. Confirm ImageMagick `magick` is on `PATH`.
2. Run `tools/kit-to-ftex.ps1` with the flags the README's Kit PNG to FTEX section lists.
   - `-TeamName` matches substrings, so check every match before converting.
3. Put the FTEX in the existing kitserver slot folder, and leave its `config.txt`, `order.ini`, and `map.txt` unchanged.
4. Ask the user to check the texture in game, because the script writes DXT5 (`PixelFormatType 4`) and some stock textures use another format.
