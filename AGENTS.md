# PES repo

@README.md

## Repo-specific

- Monorepo with independent tools in subdirectories, each with its own language and build system.
- After making changes, ask `commit/push backup?`; if user confirms, commit and push immediately.

## Tools

### evoweb (Go)

- XenForo forum scraper. Build: `go build ./src/` from `evoweb/`.
- Run: `go run ./src/ [flags] <thread-url>`.
- Requires cookies for authenticated forums. Pass via `--cookie` or `--cookie-file`.

### faces (Python)

- Player face mapping utility. Uses uv for package management.
- Setup: `uv sync` from `faces/`.
- Run: `uv run src/mapFaces.py [args]` from `faces/`.
