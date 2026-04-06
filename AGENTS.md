# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- After making changes, ask `commit/push backup?`; if user confirms, commit and push immediately.

## Tools

### evoweb (Go)

- XenForo forum scraper. Build: `go build .` from `evoweb/`.
- Subcommands: `go run . login` and `go run . scrape [flags] <url>`.
- Credentials stored plaintext at `~/.secrets/evoweb/credentials`. Run `login` once to save them.
- `--cookie` and `--cookie-file` flags override stored credentials.

### faces (Python)

- Player face mapping utility. Uses uv for package management.
- Setup: `uv sync` from `faces/`.
- Run: `uv run src/mapFaces.py [args]` from `faces/`.
