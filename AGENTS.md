# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- After making changes, ask `commit/push backup?`; if user confirms, commit and push immediately.
- When asked about PES/Football Life mods, patches, or community content, first check the AI Knowledge Base (`~/dev/priv/ai-knowledge-base/wiki/pes/`) for existing research, then use the evoweb scraper to fetch fresh thread content. Update the AIKB article when findings are worth preserving.
- Scraped JSON data lives in `evoweb/data/`. Use `duckdb` to query it for analysis across large datasets.

## Tools

### evoweb (Go)

- XenForo forum scraper. Build: `go build .` from `evoweb/`.
- Subcommands: `go run . login`, `go run . scrape [flags] <url>`, `go run . forum [flags] <url>`.
- `scrape` extracts posts from a thread. `forum` lists threads from a forum index page.
- Credentials stored plaintext at `~/.secrets/evoweb/credentials`. Run `login` once to save them.
- `--cookie` and `--cookie-file` flags override stored credentials.

### faces (Python)

- Player face mapping utility. Uses uv for package management.
- Setup: `uv sync` from `faces/`.
- Run: `uv run src/mapFaces.py [args]` from `faces/`.
