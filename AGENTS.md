# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- **Before every commit**, find and run the Makefile in the changed subdirectory (`make all`). Do not skip this.
- After making changes, ask `commit/push backup?`; if user confirms, commit and push immediately.
- When asked about PES/Football Life mods, patches, or community content, first check the AI Knowledge Base (`~/dev/priv/ai-knowledge-base/wiki/pes/`) for existing research, then use the evoweb scraper to fetch fresh thread content. Update the AIKB article when findings are worth preserving.
- Scraped JSON data lives in `evoweb/data/`. Use `duckdb` to query it for analysis across large datasets.

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
