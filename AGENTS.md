# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- **Before every commit**, find and run the Makefile in the changed subdirectory (`make all`). Do not skip this.
- If `git status` shows tracked changes, ask `commit/push?`; on confirm, commit and push.
- When asked about PES/Football Life mods, patches, or community content, first check the AI Knowledge Base (`~/dev/priv/ai-knowledge-base/wiki/pes/`) for existing research, then check scraped data in `evoweb/data/` for freshness. **If scraped data is older than 1 week, rescrape the relevant evoweb thread before answering.** Update the AIKB article when findings are worth preserving.
- AIKB notes are concise and punctual: lead with the conclusion, prefer tight bullets and short factual sentences over prose, mark hypotheses once with **Not verified** and move on. No multi-paragraph hedging or redundant restatements.
- When the task is about the live game install, read the live workspace docs at the game root before doing anything else. For this machine, the current Football Life workspace is `C:\Program Files (x86)\SP Football Life 2026\AGENTS.md` and `README.md`.
- The canonical gameplay backup root on this machine is `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- If the user gives an explicit install path, treat that path as authoritative. Do not search broader drives or user profile roots unless the user asks or the stated path fails verification.
- Evidence-first only: do not propose gameplay reset or install steps until the exact active files are verified from the live install, at minimum `SiderAddons\sider.ini`, relevant `Data\dt13/dt18` files, backups, and archive contents or source-thread instructions.
- Canonical reference for sider config (section layout, `lua.module` entries, livecpk roots, cache behavior): SOK Unleashed v9 thread at `https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/`. Check it before advising on any sider.ini change.
- Prefer file hashes over filename assumptions when identifying which gameplay is currently installed.
- For live Football Life/PES installs, do not make gameplay or config changes after verification unless the user explicitly approves the install/change step.
- Before any gameplay switch, inventory the full active gameplay stack, not just `dt18`: `dt13`, `dt18`, gameplay-related `livecpk` roots, gameplay-related `lua.module` entries, exe replacements, hook files, and cache files. Verify each component from file evidence or mod instructions.
- Never assume a file is unrelated just because it is named like an animation or visual addon. If a mod readme bundles it as part of gameplay, treat it as part of the gameplay stack until proven otherwise.
- Never perform a partial gameplay switch that leaves a mixed state unless the user explicitly asked for that exact mix.
- Prefer `tools/fl-gameplay.ps1` for recurring live-install gameplay status checks and vanilla `dt13` or `dt18` switches. It keeps the workflow consistent and repo-backed.
- Scraped JSON data lives in `evoweb/data/`. Use `duckdb` to query it for analysis across large datasets.
- On Windows, if a command needs elevation, spawn an elevated `pwsh` from the current session instead of stopping. Pattern: `Start-Process -FilePath (Get-Command pwsh.exe).Source -Verb RunAs -Wait -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File', <script>)` or pass `-Command` instead of `-File`.

## Tools

### evoweb (Go)

- XenForo forum scraper. Build: `go build .` from `evoweb/`.
- Subcommands: `go run . login`, `go run . scrape --output <file> [flags] <url>`, `go run . forum --output <file> [flags] <url>`.
- `scrape` extracts posts from a thread. `forum` lists threads from a forum index page.
- Credentials stored plaintext at `~/.secrets/evoweb/credentials`. Run `login` once to save them.
- `--cookie` and `--cookie-file` flags override stored credentials.
- **Always use `--output data/<name>.json`** when scraping. Never scrape to stdout only. All results must be persisted in `evoweb/data/` for future analysis.

### cpk (Go)

- CRI Middleware CPK archive reader. Build: `go build .` from `cpk/`.
- Subcommands: `go run . list [-l] <cpk>`, `go run . extract [--file inner-path] [--out path] <cpk>`.
- `list` prints inner paths; `-l` adds offsets and sizes. `extract` writes one file by inner path or dumps everything to a directory.
- Inner-path matching is case-insensitive and accepts both `/` and `\`. Players base lives at `common/etc/pesdb/Player.bin` inside `Data/dt00_x64.cpk`; `Data/dt10_x64.cpk` and `download/dt80_*E_x64.cpk` override in load order.
- BPB 2026 ships `Player.bin` zeroed in both `dt00` and `dt10` (1,751,422 bytes of `00`). Real player data lives in the encrypted EDIT save at `~/Documents/KONAMI/eFootball PES 2021 SEASON UPDATE/<SteamID>/save/EDIT00000000`. The cpk tool does not parse EDIT files; ejogc327's PES Editor or kisni07's PESDatabase do.

### faces (Go)

- Player face mapping and mismatch detection. Build: `go build .` from `faces/`.
- Subcommands: `go run . detect --faces-dir <path> --player-csv <file>`, `go run . map [flags]`.
- `detect` scans a livecpk faces folder for ID mismatches, orphans, non-numeric folders, and missing FPKs.
- `map` copies and remaps faces between game versions using name matching across CSV exports.
- **Live FL26 player and team databases live at the game root, not in `faces/samples/`.** Use `C:\Program Files (x86)\SP Football Life 2026\FL2621_players.txt` (current player IDs) and `FL262_teams.txt` (team IDs) as the authoritative source. The samples in `faces/samples/` are a snapshot and may lag the live install by weeks. Format conversion needed for the `faces` tool: live file is `<ID> - <Name>` with CRLF; faces tool expects a `Id;Name` semicolon CSV. Convert with `tr -d '\r' < FL2621_players.txt | sed '1iId;Name' | sed 's/ - /;/'`.
- The face install destination on this machine is `C:\Program Files (x86)\SP Football Life 2026\SiderAddons\livecpk\root\Asset\model\character\face\real\<player_id>\`. Sider is already configured (`cpk.root = .\livecpk\root`, `livecpk.enabled = 1`); dropping numeric ID folders there is sufficient, no `sider.ini` edit needed.
- The `map` command silently drops any pair where source and destination IDs have different character lengths (e.g. 5-digit `72284` to 6-digit `177929`). This is a safety check (`map.go:81`) -- the tool's hex replace is `strings.ReplaceAll`, which would change file size and corrupt FPK length-prefixed path offsets if lengths differed. Length-mismatch faces need a separate FPK-aware editor; do not force-install them.
- Salvage techniques for `map` near-misses (when source folder name doesn't fuzzy-match a destination CSV name):
  - **Direct ID match:** if the source folder ID literally exists in the destination CSV, copy the folder as-is to `dest-folder/<id>/` -- no hex remap needed.
  - **Source folder rename:** when the live name has fewer parts than the source name (e.g. live `Dion Beljo` vs source `Dion Drena Beljo`), the matcher fails on `len(targetParts) != len(candidateParts)` (`normalize.go:82`). Rename the source folder to match the live name's part count, then rerun. The internal ID inside `face.fpk` still gets remapped correctly.
- When installing faces to the live install, write a rollback record next to the source archives (e.g. `MEGA/gaming/pes/faces/_INSTALLED_<date>.txt`) listing every dest folder ID and the not-installed reasons. The user can clean up by deleting the listed numeric folders.
