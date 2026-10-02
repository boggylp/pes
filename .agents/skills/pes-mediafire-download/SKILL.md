---
name: pes-mediafire-download
description: 'Download a PES or Football Life MediaFire folder with mdrs when the user provides a folder URL.'
metadata:
  trusted_sources:
    - https://github.com/NicKoehler/mediafire_rs
    - https://crates.io/crates/mediafire_rs
---

# PES MediaFire download

This workflow takes `mediafire.com/folder/` URLs only.

## Steps

1. Confirm `where.exe mdrs` finds the tool.
   - When it is absent, get approval, then run `cargo install mediafire_rs`.
2. Pick an absolute, empty output directory.
   - `mdrs` replaces same-named files without a prompt, so list what a populated directory would lose and get approval.
3. Download: `mdrs --tries 3 -o "<output-dir>" <mediafire-folder-url>`.
   - When MediaFire rate-limits, add `--max 3`.
   - When MediaFire shows a login or interstitial page, use the authenticated browser profile.
4. Check the file count and total size against the forum post or mod README.
5. Keep the download until the installed mod passes verification.
