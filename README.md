# PES

Pro Evolution Soccer / SP Football Life related utilities.

## Structure

```text
evoweb/     XenForo forum scraper (Go)
faces/      Player face mapping and mismatch detection (Go)
tools/      Local Windows helpers for PES and Football Life workflows
```

## tools

Repo-backed PowerShell helpers for recurring local workflows.

### Football Life gameplay helper

`tools/fl-gameplay.ps1` provides a minimal workflow for live-install gameplay checks and vanilla switches.

Defaults:

- Game root: `%ProgramFiles(x86)%\SP Football Life 2026`
- Gameplay backup root: `%USERPROFILE%\MEGA\gaming\pes\gameplay`

Usage:

```powershell
pwsh -File .\tools\fl-gameplay.ps1 status
pwsh -File .\tools\fl-gameplay.ps1 switch-dt13-vanilla
pwsh -File .\tools\fl-gameplay.ps1 switch-dt18-vanilla
```

What it does:

- `status` prints hashes for live `dt13`, `dt18`, and `FL_2026.exe`, the tracking comments from `SiderAddons\sider.ini`, active gameplay-related sider entries, and `SYSTEM` cache presence.
- `switch-dt13-vanilla` backs up the current live `dt13`, restores the vanilla `dt13` from the canonical gameplay backup root, updates the `sider.ini` tracking comment, and removes the current `SYSTEM` cache file.
- `switch-dt18-vanilla` does the same for `dt18`.

The helper prefers loose files under `%USERPROFILE%\MEGA\gaming\pes\gameplay\vanilla\` and falls back to the `dt13 & dt18 vanilla.rar` archive when needed.

## evoweb

Scrapes threads and forum listings from XenForo-based forums (e.g. evoweb.uk). Outputs structured JSON.

### Build

```sh
cd evoweb
go build .
```

### Usage

```sh
# Save evoweb.uk credentials (one-time setup, stored at ~/.secrets/evoweb/credentials)
go run . login

# Scrape a thread (auto-logs in with stored credentials)
go run . scrape --output data/example.json "https://evoweb.uk/threads/example.88633/"

# Limit pages
go run . scrape --output data/example.json --max-pages 3 "https://evoweb.uk/threads/example.88633/"

# Scrape only the last N pages of a thread
go run . scrape --output data/example.json --last-pages 5 "https://evoweb.uk/threads/example.88633/"

# List threads from a forum (default: 1 page)
go run . forum --output data/forum.json "https://evoweb.uk/forums/pes-2021.337/"

# List threads from multiple pages
go run . forum --output data/forum.json --max-pages 3 "https://evoweb.uk/forums/pes-2021.337/"

# Manual cookie override (skips stored credentials)
go run . scrape --output data/example.json --cookie "xf_session=abc; xf_user=def" "https://evoweb.uk/threads/example.88633/"
```

## faces

Player face mapping and mismatch detection for PES/Football Life. Matches player names across CSV exports and copies face asset directories.

### Build

```sh
cd faces
go build .
```

### Usage

```sh
# Detect mismatched/orphaned faces in livecpk folder
go run . detect \
  --faces-dir "/path/to/livecpk/.../face/real" \
  --player-csv "/path/to/UML 2026 - Player IDs.csv"

# Map faces between game versions
go run . map \
  --source-csv samples/BPB-2023-players.csv \
  --destination-csv samples/FL26_players.csv \
  --source-folder /path/to/source/faces \
  --dest-folder /path/to/destination/faces
```
