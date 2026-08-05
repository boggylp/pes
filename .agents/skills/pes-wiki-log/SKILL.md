---
name: pes-wiki-log
description: 'Write or update a PES or Football Life wiki article when the user asks to record a durable finding.'
---

# PES wiki log

Articles live at `~/memory/priv/wiki/pes/`. The universal `mem-wiki` skill owns the article format, the frontmatter, and the line cap. This skill adds the PES evidence rules.

## Steps

1. Select the article by topic.

   | Article | Scope |
   | --- | --- |
   | `gameplay-combos-fl26.md` | Football Life 2026 gameplay combinations and verdicts |
   | `bpb-2026.md` | BPB install state |
   | `bpb-2026-player-database.md` | BPB roster, EDIT save format, and ID procedures |
   | `livecpk-face-management.md` | Faces and livecpk workflows |
   | `evoweb-credibility-signals.md` | Evoweb author and mod evidence |
   | `ai-tweaks-twiggy.md` | AI Tweaks findings |
   | `soccer-revolution-revamped.md` | Soccer Revolution |
   | `simsnob-experience.md` | SimSnob |
   | `uml-2026.md` | Ultimate Master League install state |
   | `player-edits-automation.md` | EDIT save tooling and the headless pipeline |

   Create a kebab-case article under `~/memory/priv/wiki/pes/` when no existing article fits. Display-related stutter belongs in `~/memory/priv/wiki/windows/multi-monitor-lag-gaming.md`.

2. Run `hostname`. Add the verified hostname to each machine-specific heading.

3. Lead with the conclusion. Give concise evidence in short bullets. Mark an unverified hypothesis once with `**Not verified**`.

4. Cite gameplay files by hash. Run `pes-gameplay-status` when the required hashes are missing.

5. Link related articles with Obsidian wikilinks.

6. Ask `commit/push?` after the entry is complete. On confirmation, commit and push only the changed article.

   ```sh
   cd ~/MEGA/home/pkb
   git add memory/priv/wiki/pes/<file>.md
   git commit -m "🤖: Log <subject> from <hostname>"
   git push
   ```

## Review

- State each verdict as a standing judgment, so the article reads as a ranking record.
- Rewrite a superseded verdict in place, and delete the bullets it replaces.
- Quote one necessary sentence from an Evoweb post and include its URL.
- Replace marketing claims with the tested state, hashes, host, and verdict.
- Report the changed article path without status narration.

## Boundaries

- Use `pes-gameplay-status` to collect live gameplay evidence.
- Use `pes-evoweb-research` to collect Evoweb evidence.
- Live install state belongs in the `sider.ini` `; gameplay:` tracking line.
- A dated comparison or a buy decision belongs in `~/memory/priv/research/` through `mem-research`.
