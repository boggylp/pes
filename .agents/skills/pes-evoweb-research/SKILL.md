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

2. Check the date of the relevant files under `evoweb/data/`. Rescrape the forum listing and active threads for a what-is-new question. Rescrape other data when it is more than one week old.

3. Save each scrape to a dated JSON file.

   ```sh
   cd evoweb
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json "<thread-url>"
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json --last-pages 5 "<thread-url>"
   go run . forum --output data/<forum>-$(date +%Y-%m-%d).json "<forum-url>"
   ```

   Run `go run . login` once when stored credentials are absent. The tool reads `~/.secrets/evoweb/credentials`.

4. Query the nested `posts` array with DuckDB.

   ```sh
   duckdb -c "SELECT p.author, p.date, p.content FROM (SELECT unnest(posts) AS p FROM read_json('evoweb/data/<file>.json')) WHERE p.content LIKE '%<keyword>%' ORDER BY p.date DESC LIMIT 20;"
   ```

5. Retrieve an attachment URL with the authenticated browser profile because scrape JSON omits attachment links. Download the URL with the stored Evoweb session.

   ```sh
   go run . download --output <file> "https://evoweb.uk/attachments/<name>.<id>/"
   ```

6. Cross-check release claims against the original post and independent replies. Mark unsupported claims as `**Not verified**`.

7. Cite the thread URL, post date, and author in the answer.

8. Use `pes-wiki-log` when the finding is durable.

## Review

- Answer from gathered evidence. Do not send the user to search the thread.
- Keep scrape files under the ignored `evoweb/data/` directory.
- Use stored credentials for routine work. Use cookie overrides only to diagnose login behavior.
- Stop when current source evidence is unavailable. State which source could not be refreshed.

## Boundaries

- Use `pes-gameplay-status` for live-install state.
- Use `pes-gameplay-set` for a gameplay change.
- Use `pes-wiki-log` to write the durable finding.
