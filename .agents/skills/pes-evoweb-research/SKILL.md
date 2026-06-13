---
name: pes-evoweb-research
description: '**Invoke this skill BEFORE answering any question about PES / Football Life community releases, mod versions, author activity, or thread state on evoweb.uk.** Covers the `evoweb/` Go scraper (`scrape` for thread pages, `forum` for forum indices), the `--output data/<name>.json` mandate, the freshness rule (rescrape if `evoweb/data/*.json` is older than one week for the question), the `duckdb` query pattern over scraped JSON, the AIKB check-first rule, and the never-tell-the-user-to-search-evoweb-themselves rule. Triggers: "what is new in F4L?", "did the patch update?", "is the author still active?", "check evoweb for X", "any new release of Y", any phrase that needs evidence from a XenForo thread.'
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

2. **Check scraped data freshness.** `evoweb/data/<name>-<YYYY-MM-DD>.json` is the convention. If the latest scrape relevant to the question is older than **one week**, rescrape before answering. See repo `AGENTS.md`.

3. **Scrape with `--output`, never to stdout.**

    ```sh
    cd evoweb
    go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json "<thread-url>"
    # Limit pages when the thread is long and only the tail matters
    go run . scrape --output data/<topic>-$(date +%Y-%m-%d).json --last-pages 5 "<thread-url>"
    # Forum index page (lists threads)
    go run . forum --output data/<forum>-$(date +%Y-%m-%d).json "<forum-url>"
    ```

    Credentials at `~/.secrets/evoweb/credentials` are loaded automatically; run `go run . login` once to save them.

4. **Query with duckdb** over the resulting JSON:

    ```sh
    duckdb -c "SELECT author, posted_at, content FROM read_json_auto('evoweb/data/<file>.json', maximum_object_size=104857600) WHERE content LIKE '%<keyword>%' ORDER BY posted_at DESC LIMIT 20;"
    ```

    For very large threads, project only the columns you need; the full text payload can be hundreds of MB.

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
