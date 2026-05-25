---
name: pes-mediafire-download
description: '**Invoke this skill BEFORE downloading any MediaFire folder URL referenced in a pes / Football Life forum thread, mod README, or evoweb scraped post.** Covers `mdrs` (the `mediafire_rs` crate by NicKoehler) for bulk folder downloads, the `cargo install` setup, the retry flag for flaky links, and the split with Gopeed (single files) and yt-dlp (does not handle MediaFire at all). Triggers: a `mediafire.com/folder/` URL, "download this MediaFire folder", "grab this mod from MediaFire", any pes/FL mod link in evoweb data or forum threads where the destination is MediaFire.'
metadata:
  trusted_sources:
    - https://github.com/NicKoehler/mediafire_rs
    - https://crates.io/crates/mediafire_rs
---

# PES MediaFire bulk download

PES / Football Life mods are routinely hosted on MediaFire as multi-file folders. `mdrs` is the only viable bulk downloader for those: yt-dlp does not support MediaFire, and Gopeed resolves single files only.

## When to use

- A forum post or evoweb scrape (e.g. `evoweb/data/*.json`) links to `mediafire.com/folder/...`.
- A mod README ships a MediaFire folder for assets, dt files, faces, kits, etc.
- The user pastes a MediaFire folder URL.

For a `mediafire.com/file/...` single-file URL, use Gopeed or a direct browser download instead. `mdrs` accepts single files too, but for one file it is overkill.

## Setup (one-shot, per machine)

```sh
cargo install mediafire_rs
where.exe mdrs   # confirms `~/.cargo/bin/mdrs.exe` (or equivalent) is on PATH
```

Build prerequisites: a working Rust toolchain (`rustup` or scoop `rust`). Compile time: ~2 min on a warm cargo cache.

If the user does not have Rust installed and does not want it, stop and ask before installing the toolchain.

## Usage

```sh
mdrs -o <output-dir> --tries 3 <mediafire-folder-url>
```

Flags (verified from upstream README):

| Flag | Default | Purpose |
| --- | --- | --- |
| `-o, --output <DIR>` | `.` | Output directory. Pass an absolute path; `mdrs` creates it if missing. |
| `-m, --max <N>` | `10` | Concurrent downloads. Lower to 3-5 if MediaFire rate-limits. |
| `-t, --tries <N>` | `1` | Retries per file. **Always pass `--tries 3`**, MediaFire 503s are routine. |
| `-r, --reverse` | off | Largest files first. Useful when total size matters more than file count. |
| `-p, --proxy <FILE>` | none | Proxy list, one per line. For API-only by default. |
| `--proxy-download` | off | Route file downloads through the proxies too, not just metadata calls. |

### Common patterns

Download a mod folder to the canonical pes asset staging dir:

```sh
mdrs --tries 3 \
     -o "$USERPROFILE/MEGA/gaming/pes/<modtype>/<mod-name>" \
     "https://www.mediafire.com/folder/<id>/<name>"
```

The pes `AGENTS.md` names `%USERPROFILE%\MEGA\gaming\pes\gameplay\` as the canonical gameplay backup root. Use a parallel subdir (`faces`, `kits`, `dt18`, etc.) for non-gameplay mods. Confirm the destination with the user before starting a large download.

Quick scratch download (small, exploratory):

```sh
mdrs --tries 3 -o ~/tmp/mdrs-<short-name> "<url>"
```

## Pitfalls

- **MediaFire 503s are random.** Without `--tries`, one transient failure aborts that file. Always pass `--tries 3` at minimum.
- **No resume.** A killed `mdrs` run restarts each in-flight file from zero. For huge folders, prefer letting it complete or split the URL list.
- **Folder URL vs file URL.** `mediafire.com/folder/<id>` works. `mediafire.com/file/<id>` is a single file; mdrs handles it but Gopeed is the lighter tool.
- **Paths with spaces.** Always quote `-o` paths on Windows (`"$USERPROFILE/MEGA/..."`), `mdrs` does not re-quote them internally.
- **Output dir already populated.** `mdrs` overwrites without prompting. If re-downloading, the prior copy is gone the moment a same-named file lands.
- **Login walls.** MediaFire occasionally throws an interstitial for very large folders. `mdrs` cannot solve that; fall back to a logged-in browser via the [[playwright]] skill's persistent-profile mode.
- **Never `rm` the download** before verifying contents in-game. Mirrors [[pes-faces-install]] rollback discipline.

## After download

- Verify file count and size against the forum post / mod README before any install step.
- For gameplay mods: follow [[pes-gameplay-switch]] for the actual install.
- For faces: route the extracted folder through [[pes-faces-install]].
- Log non-obvious findings (working URL for a mod, hash of the canonical archive) in the AIKB via [[pes-aikb-log]].
