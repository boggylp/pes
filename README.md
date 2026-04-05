# PES

Pro Evolution Soccer / SP Football Life related utilities.

## Structure

```text
evoweb/     XenForo forum scraper (Go)
faces/      Player face mapping between game versions (Python)
```

## evoweb

Scrapes threads from XenForo-based forums (e.g. evoweb.uk). Outputs structured JSON with thread title, posts, authors, and dates.

### Build

```sh
cd evoweb
go build ./src/
```

### Usage

```sh
# Save evoweb.uk credentials (one-time setup, stored at ~/.secrets/evoweb/credentials)
go run ./src/ login

# Scrape a thread (auto-logs in with stored credentials)
go run ./src/ scrape "https://evoweb.uk/threads/example.88633/"

# Limit pages
go run ./src/ scrape --max-pages 3 "https://evoweb.uk/threads/example.88633/"

# Scrape only the last N pages of a thread
go run ./src/ scrape --last-pages 5 "https://evoweb.uk/threads/example.88633/"

# Manual cookie override (skips stored credentials)
go run ./src/ scrape --cookie "xf_session=abc; xf_user=def" "https://evoweb.uk/threads/example.88633/"
```

## faces

Maps player faces between PES/Football Life versions by matching player names across CSV exports and copying face asset directories.

### Setup

```sh
cd faces
uv sync
```

### Usage

```sh
uv run src/mapFaces.py \
  --source-csv samples/BPB-2023-players.csv \
  --destination-csv samples/FL26_players.csv \
  --source-folder /path/to/source/faces \
  --dest-folder /path/to/destination/faces
```
