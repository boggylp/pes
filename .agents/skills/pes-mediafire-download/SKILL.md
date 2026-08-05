---
name: pes-mediafire-download
description: 'Download a PES or Football Life MediaFire folder with mdrs when the user provides a folder URL.'
metadata:
  trusted_sources:
    - https://github.com/NicKoehler/mediafire_rs
    - https://crates.io/crates/mediafire_rs
---

# PES MediaFire download

## Setup

1. Confirm that `mdrs` is available.

   ```sh
   where.exe mdrs
   ```

2. When it is absent, get approval before installing a Rust toolchain or package. Then install and verify it.

   ```sh
   cargo install mediafire_rs
   where.exe mdrs
   ```

## Steps

1. Use this workflow for `mediafire.com/folder/` URLs. Route a single-file URL to the selected single-file downloader.

2. Select an absolute output directory. Use a new or empty directory when possible.

3. For a populated output directory, list the files that can be replaced and get explicit approval.

4. Download with three retries per file.

   ```sh
   mdrs --tries 3 -o <output-dir> <mediafire-folder-url>
   ```

5. Reduce concurrency when the service rate-limits requests.

   ```sh
   mdrs --tries 3 --max 3 -o <output-dir> <mediafire-folder-url>
   ```

6. Verify file count and total size against the forum post or mod README.

7. Preserve the download until the installed mod passes verification.

## Options

| Flag | Default | Effect |
| --- | --- | --- |
| `-o, --output <DIR>` | `.` | Select the output directory. |
| `-m, --max <N>` | `10` | Set concurrent downloads. |
| `-t, --tries <N>` | `1` | Set retries per file. Use `3` for this workflow. |
| `-r, --reverse` | off | Download the largest files first. |
| `-p, --proxy <FILE>` | none | Use a proxy list for API calls. |
| `--proxy-download` | off | Use proxies for file downloads. |

## Review

- Quote Windows paths that contain spaces.
- A restarted run downloads interrupted files again from the start.
- `mdrs` can replace same-named files without a prompt.
- Use the authenticated browser profile when MediaFire presents a login or interstitial page.

## Boundaries

- Use `pes-gameplay-switch` to install gameplay files.
- Use `pes-faces-install` to install faces.
- Use `pes-wiki-log` to record durable source URLs or archive hashes.
