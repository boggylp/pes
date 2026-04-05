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
# Public thread (no auth)
go run ./src/ "https://xenforo.com/community/threads/example.12345/"

# Authenticated forum
go run ./src/ --cookie "xf_session=abc; xf_user=def" "https://evoweb.uk/threads/example.88633/"

# From cookie file
go run ./src/ --cookie-file ~/.secrets/evoweb-cookie "https://evoweb.uk/threads/example.88633/"

# Limit pages
go run ./src/ --max-pages 3 "https://evoweb.uk/threads/example.88633/"
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
