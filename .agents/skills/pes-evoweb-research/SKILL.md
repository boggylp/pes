---
name: pes-evoweb-research
description: 'Research PES or Football Life community releases on evoweb.uk when an answer needs current thread evidence.'
metadata:
  trusted_sources:
    - https://evoweb.uk/
---

# Evoweb research

## Steps

1. Search `~/memory/priv/wiki/pes/` for an existing answer.
2. Rescrape the forum listing and active threads for a what-is-new question, and any other scrape under `evoweb/data/` older than one week.
3. Save each scrape to a dated JSON file under `evoweb/data/`.

   ```sh
   cd evoweb
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json "<thread-url>"
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json --last-pages 5 "<thread-url>"
   go run . forum --output data/<forum>-$(date +%Y-%m-%d).json "<forum-url>"
   ```

   - When stored credentials are absent, run `go run . login` once.
   - Use `--cookie` only to diagnose login behavior.

4. Query the nested `posts` array with DuckDB.

   ```sh
   duckdb -c "SELECT p.author, p.date, p.content FROM (SELECT unnest(posts) AS p FROM read_json('evoweb/data/<file>.json')) WHERE p.content LIKE '%<keyword>%' ORDER BY p.date DESC LIMIT 20;"
   ```

5. Scrape JSON omits attachment links, so take an attachment URL from the authenticated browser profile and download it with the stored session.

   ```sh
   go run . download --output <file> "https://evoweb.uk/attachments/<name>.<id>/"
   ```

6. Cross-check each release claim against the original post and independent replies, and mark an unsupported one `**Not verified**`.
7. Answer from the gathered evidence, citing thread URL, post date, and author.
8. When a source cannot be refreshed, stop and name it.
9. Load `pes-wiki-log` when the finding is durable.
