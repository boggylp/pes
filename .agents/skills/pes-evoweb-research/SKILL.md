---
name: pes-evoweb-research
description: 'Invoke BEFORE answering any question about PES / Football Life community releases, mod versions, author activity, or thread state on evoweb.uk. Covers the `evoweb/` Go scraper (`scrape`, `forum`, `download`), the `--output data/<name>.json` mandate, the freshness rule, the `duckdb` query pattern, and the AIKB check-first rule. Triggers: "what is new in F4L?", "did the patch update?", "is the author still active?", "check evoweb for X", any phrase that needs evidence from a XenForo thread.'
metadata:
  trusted_sources:
    - https://evoweb.uk/
---

# Evoweb research

Pull thread data with the scraper, query with duckdb, cross-check the AIKB. Never punt the search back to the user.

## When to use

- Any factual question about a PES / Football Life community release that hinges on what was posted on evoweb
- Verifying author activity, release dates, version numbers, install instructions
- Sourcing the basis for an AIKB update

## Steps

1. **Check the AIKB first** for the relevant article under `~/dev/priv/ai-knowledge-base/wiki/pes/`. If the answer is already there and not stale, use it.

2. **Check scraped data freshness.** `evoweb/data/<name>-<YYYY-MM-DD>.json` is the convention. What's-new questions always rescrape the forum listing + active threads first, regardless of scrape age; other questions rescrape when the relevant data is older than **one week**. See repo `AGENTS.md`.

3. **Scrape with `--output`, never to stdout.**

   ```sh
   cd evoweb
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json "<thread-url>"
   # Limit pages when the thread is long and only the tail matters
   go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json --last-pages 5 "<thread-url>"
   # Forum index page (lists threads)
   go run . forum --output data/<forum>-$(date +%Y-%m-%d).json "<forum-url>"
   # Login-walled attachment (hrefs never land in scrape JSON; extract them per repo AGENTS.md)
   go run . download --output <file> "https://evoweb.uk/attachments/<name>.<id>/"
   ```

   Credentials at `~/.secrets/evoweb/credentials` are loaded automatically; run `go run . login` once to save them.

4. **Query with duckdb** over the resulting JSON (posts sit in a nested `posts` array):

   ```sh
   duckdb -c "SELECT p.author, p.date, p.content FROM (SELECT unnest(posts) AS p FROM read_json('evoweb/data/<file>.json')) WHERE p.content LIKE '%<keyword>%' ORDER BY p.date DESC LIMIT 20;"
   ```

   For very large threads, project only the columns you need. `jq` works too for simple filters.

5. **Cite the source** in any answer: thread URL, post date, author. Evoweb is the canonical source for community releases; do not rely on second-hand summaries.

6. **Update the AIKB article** if findings are worth preserving. See pes-aikb-log.

## Pitfalls

- **Never tell the user "search evoweb"** or "check the thread". Scrape it or grep existing data yourself.
- **Never commit scraped data.** `evoweb/data/` is gitignored. Do not suggest committing the JSON files.
- **Always `--output data/<name>.json`.** Scraping to stdout drops the only durable record of the page state at scrape time.
- **Stale data lies.** A two-month-old scrape of the F4L thread will miss a release. Verify the latest post date against the live page before drawing conclusions from old JSON.
- **Cookie override is for debugging only.** `--cookie` / `--cookie-file` bypass stored credentials; the default login flow is the right one for routine scrapes.
- **Adversarial self-review.** Forum claims are user-generated. A single post claiming a patch "fixes everything" is a hypothesis, not a fact. Look for multiple independent confirmations or mark the claim **Not verified** in the AIKB. See pes-aikb-log.

## Out of scope

- Reading the live install state: see pes-gameplay-status.
- Performing a gameplay switch based on findings: see pes-gameplay-switch.
- Writing the durable finding to the AIKB: see pes-aikb-log.
