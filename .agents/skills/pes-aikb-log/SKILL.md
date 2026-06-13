---
name: pes-aikb-log
description: '**Invoke this skill BEFORE writing or updating a PES / Football Life article in the AI Knowledge Base at `~/dev/priv/ai-knowledge-base/wiki/pes/`.** Covers the existing article set (`gameplay-combos-fl26.md`, `bpb-2026.md`, `livecpk-face-management.md`, `evoweb-credibility-signals.md`, and others), the mandatory hostname tag on every playtest / verdict entry (`BogambeDesktop` desktop / `DESKTOP-J0MDFMU` laptop), the AIKB house style (lead with conclusion, tight bullets, mark hypotheses once with `**Not verified**`), and the commit-and-push step that closes the loop. Triggers: "log this to the KB", "save the verdict", "record the install state in obsidian-kb", "update the AIKB", "this is the best combo, log it", "/obsidian-kb …" for PES content.'
metadata:
  trusted_sources:
    - https://keepachangelog.com/en/1.1.0/
---

# PES AI Knowledge Base log

The final step of most PES sessions: write durable findings into `wiki/pes/`. Hostname-tagged, dated, concise.

## When to use

- After a playtest the user wants recorded
- After pes-evoweb-research surfaces a fact worth keeping
- After pes-gameplay-status captures a noteworthy live install state
- When the user says any variant of "log this to the KB" / "save to obsidian-kb" / "diary it" with PES content

## Steps

1. **Pick the right article.** Match by topic, not by date.

    | Article                                       | Scope                                                   |
    | --------------------------------------------- | ------------------------------------------------------- |
    | `wiki/pes/gameplay-combos-fl26.md`            | FL26 desktop install gameplay combos and verdicts       |
    | `wiki/pes/bpb-2026.md`                        | BPB laptop install (the `DESKTOP-J0MDFMU` machine)      |
    | `wiki/pes/livecpk-face-management.md`         | Faces, livecpk, face install workflow                   |
    | `wiki/pes/evoweb-credibility-signals.md`      | Author / mod reputation signals from evoweb             |
    | `wiki/pes/ai-tweaks-twiggy.md`                | Twiggy / AI-tweaks specific findings                    |
    | `wiki/pes/soccer-revolution-revamped.md`      | Soccer Revolution patch                                 |
    | `wiki/pes/simsnob-experience.md`              | SimSnob mod experience                                  |
    | `wiki/pes/multi-monitor-lag-gaming.md`        | Multi-monitor stutter / gaming display issues           |

    If no article fits, create a new one with a kebab-case filename.

2. **Tag every machine-specific entry with the hostname.** The desktop is `BogambeDesktop`; the laptop is `DESKTOP-J0MDFMU`. Examples:

    ```markdown
    ## User's best combo verdict (2026-05-20, `BogambeDesktop`)
    ```

    Never write "this desktop" or "the laptop" without the hostname.

3. **Match the AIKB house style.** Lead with the conclusion, then evidence. Tight bullets and short factual sentences over prose. Mark hypotheses **once** with `**Not verified**` and move on; do not hedge again. No multi-paragraph qualifications, no redundant restatements. See repo `AGENTS.md` "AIKB notes are concise and punctual".

4. **Cite hashes, not filenames** for any gameplay claim. Run pes-gameplay-status first if hashes are not yet in hand.

5. **Cross-link related articles** using the vault's Obsidian wikilink syntax. The pkb is an Obsidian vault; backlinks make it navigable.

6. **Commit and push the vault** when the entry is durable:

    ```sh
    cd ~/dev/priv/ai-knowledge-base
    git add wiki/pes/<file>.md
    git commit -m "🤖: Log <subject> from <hostname>"
    git push
    ```

    The vault commits use the same `🤖:` prefix as code commits (see global `AGENTS.md` "AI-output format").

## Pitfalls

- **No hostname = entry is wrong.** Two machines feed this vault. An untagged playtest verdict is unfalsifiable across machines.
- **No status theater.** Drop "I have updated the AIKB", "I will now log this", "I should have …". Name the file edited and stop. See `AGENTS.md` "Outbound text is the highest priority".
- **No marketing in the wiki.** "Best ever", "revolutionary", "always" are wrong even when the user is excited. Write what was tested, on what hashes, with what verdict.
- **Quote evoweb posts surgically.** Single sentence plus URL, never multi-paragraph copy. See `AGENTS.md` "Evidence and citations".
- **Do not delete a prior verdict to replace it.** Edit in place; the historical entry (with its date and hostname) is the audit trail. See `AGENTS.md` "Edit in place".

## Out of scope

- Running the status check that produces the verdict: see pes-gameplay-status.
- Scraping evoweb for the facts being logged: see pes-evoweb-research.
- Anything in the database/ side of the vault. This skill is wiki-only.
