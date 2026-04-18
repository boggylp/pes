# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- **Before every commit**, find and run the Makefile in the changed subdirectory (`make all`). Do not skip this.
- If `git status` shows tracked changes, ask `commit/push backup?`; on confirm, commit and push.
- When asked about PES/Football Life mods, patches, or community content, first check the AI Knowledge Base (`~/dev/priv/ai-knowledge-base/wiki/pes/`) for existing research, then check scraped data in `evoweb/data/` for freshness. **If scraped data is older than 1 week, rescrape the relevant evoweb thread before answering.** Update the AIKB article when findings are worth preserving.
- When the task is about the live game install, read the live workspace docs at the game root before doing anything else. For this machine, the current Football Life workspace is `C:\Program Files (x86)\SP Football Life 2026\AGENTS.md` and `README.md`.
- The canonical gameplay backup root on this machine is `%USERPROFILE%\MEGA\gaming\pes\gameplay\`.
- If the user gives an explicit install path, treat that path as authoritative. Do not search broader drives or user profile roots unless the user asks or the stated path fails verification.
- Evidence-first only: do not propose gameplay reset or install steps until the exact active files are verified from the live install, at minimum `SiderAddons\sider.ini`, relevant `Data\dt13/dt18` files, backups, and archive contents or source-thread instructions.
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

### faces (Go)

- Player face mapping and mismatch detection. Build: `go build .` from `faces/`.
- Subcommands: `go run . detect --faces-dir <path> --player-csv <file>`, `go run . map [flags]`.
- `detect` scans a livecpk faces folder for ID mismatches, orphans, non-numeric folders, and missing FPKs.
- `map` copies and remaps faces between game versions using name matching across CSV exports.
