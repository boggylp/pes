---
name: pes-gameplay-status
description: 'Inspect the active gameplay stack of a live PES or Football Life install before answering a gameplay-status question.'
---

# PES gameplay status

This check is read-only.

## Steps

1. Run the helper: `pwsh -File .\tools\fl-gameplay.ps1 status`.
2. Resolve each hash against the matching article under `~/memory/priv/wiki/pes/`, and report an unknown one by hash alone.
3. Inspect what the helper does not name:
   - Executable hooks and replacements.
   - A bundled Lua module counts as active only when it registers a Sider event.
   - Bundled animation files count as gameplay.
4. State the effective gameplay from load order, and the wired stack too when they differ.
5. Before a machine-specific record, run `hostname` and use its result.
