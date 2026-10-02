---
name: pes-wiki-log
description: 'Write or update a PES or Football Life wiki article when the user asks to record a durable finding.'
---

# PES wiki log

Load `mem-wiki`, which owns the article format. This skill adds the PES evidence rules.

## Steps

1. List `~/memory/priv/wiki/pes/`, pick the article whose name covers the topic, and create a kebab-case one when none fits.
   - Display stutter belongs in `~/memory/priv/wiki/windows/multi-monitor-lag-gaming.md`.
2. Run `hostname`, and put its result in each machine-specific heading.
3. Lead with the verdict as a standing judgment, then the evidence that settled it in short bullets.
   - Mark an unverified hypothesis once with `**Not verified**`.
   - Replace a marketing claim with the tested state, hashes, host, and verdict.
4. Cite gameplay files by hash, loading `pes-gameplay-status` when a hash is missing.
5. Quote at most one sentence from an Evoweb post, with its URL.
6. Keep only numbers that stay true, such as hashes, sizes, versions, and parameter values.
   - A rank, count, or dated comparison belongs in a `mem-study` study, linked from the article.
7. Rewrite a superseded verdict in place, and delete the bullets it replaces.
8. Link related articles with Obsidian wikilinks, and report the changed article path.
