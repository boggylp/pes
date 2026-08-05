---
name: pes-aikb-log
description: 'Write or update a PES or Football Life AI Knowledge Base article when the user asks to record a durable finding.'
---

# PES AI Knowledge Base log

## Steps

1. Select the article by topic.

   | Article | Scope |
   | --- | --- |
   | `wiki/pes/gameplay-combos-fl26.md` | Football Life 2026 gameplay combinations and verdicts |
   | `wiki/pes/bpb-2026.md` | BPB install state |
   | `wiki/pes/livecpk-face-management.md` | Faces and livecpk workflows |
   | `wiki/pes/evoweb-credibility-signals.md` | Evoweb author and mod evidence |
   | `wiki/pes/ai-tweaks-twiggy.md` | AI Tweaks findings |
   | `wiki/pes/soccer-revolution-revamped.md` | Soccer Revolution |
   | `wiki/pes/simsnob-experience.md` | SimSnob |
   | `wiki/pes/multi-monitor-lag-gaming.md` | Display-related stutter |

   Create a kebab-case article when no existing article fits.

2. Run `hostname`. Add the verified hostname to each machine-specific heading.

3. Lead with the conclusion. Give concise evidence in short bullets. Mark an unverified hypothesis once with `**Not verified**`.

4. Cite gameplay files by hash. Run `pes-gameplay-status` when the required hashes are missing.

5. Link related articles with Obsidian wikilinks.

6. Ask `commit/push?` after the entry is complete. On confirmation, commit and push only the changed article.

   ```sh
   cd ~/dev/priv/ai-knowledge-base
   git add wiki/pes/<file>.md
   git commit -m "🤖: Log <subject> from <hostname>"
   git push
   ```

## Review

- Keep prior dated verdicts. Edit a historical entry in place when it is wrong.
- Quote one necessary sentence from an Evoweb post and include its URL.
- Replace marketing claims with the tested state, hashes, host, and verdict.
- Report the changed article path without status narration.

## Boundaries

- Use `pes-gameplay-status` to collect live gameplay evidence.
- Use `pes-evoweb-research` to collect Evoweb evidence.
- Write only under `wiki/pes/`.
