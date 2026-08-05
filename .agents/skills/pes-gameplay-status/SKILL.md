---
name: pes-gameplay-status
description: 'Inspect the active gameplay stack of a live PES or Football Life install before answering a gameplay-status question.'
metadata:
  trusted_sources:
    - https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/
---

# PES gameplay status

## Steps

1. Run the helper from this repository.

   ```powershell
   pwsh -File .\tools\fl-gameplay.ps1 status
   ```

2. Run `hostname` before creating a machine-specific record. Use the verified hostname in that record.

3. Resolve each SHA-256 hash against the relevant AI Knowledge Base article under `wiki/pes/`.

4. Inventory the complete gameplay stack:

   - `Data/dt13_all.cpk`
   - `Data/dt18_all.cpk`
   - The game executable
   - Each gameplay-related `lua.module` entry
   - Each gameplay-related `cpk.root` entry
   - Executable hooks and replacements
   - SYSTEM-cache presence

5. Determine effective gameplay from load order. State both the wired stack and effective gameplay when they differ.

6. Report unknown components by hash without assigning a release name.

## Review

- Identify files by hash, not filename.
- Treat a bundled Lua module as active only when it registers a real Sider event.
- Treat bundled animation files as gameplay-stack components.
- Keep the check read-only.
- Report cache presence without changing it.

## Boundaries

- Use `pes-gameplay-switch` for a gameplay change.
- Use `pes-wiki-log` to record a verdict.
- Keep Sider configuration unchanged during a status check.
