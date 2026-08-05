---
name: pes-live-install
description: 'Install or restore a PES or Football Life patch stack, write an EDIT save, or add a custom player to a live install.'
metadata:
  trusted_sources:
    - https://evoweb.uk/threads/soulsofkaos-unleashed-9-pes2013.101010/
---

# PES live install

## Steps

1. Run `hostname`. Record the verified host with each install-state entry.

2. Read the live install's `AGENTS.md` and `README.md`.

3. Run `pes-gameplay-status`. Record current hashes, Sider entries, and save-cache state.

4. Confirm the user-provided install root and staging directory. Do not search other drives unless the provided path fails or the user asks.

5. Read the patch instructions from its README or Evoweb thread. List the required parts in their documented order.

6. Before each write, state the exact target and get explicit user approval.

7. Apply one patch part at a time. Verify it before the next part. Skip a standalone EDIT save when a later approved part already contains it.

8. Write an EDIT save only to:

   ```text
   %USERPROFILE%\Documents\KONAMI\eFootball PES 2021 SEASON UPDATE\2026\save\EDIT00000000
   ```

   Treat copies under MEGA as sources, not live targets.

9. For a custom player, import the complete player record with the configured EDIT-save editor. Include attributes, physical data, traits, name, and kit data.

10. Use `pes-roster-extract` and `pes-faces-install` when the operation also changes player IDs or face folders.

11. Apply required Windows settings. Keep Hardware-accelerated GPU scheduling off.

12. When elevation is required, run an approved script with an elevated PowerShell process from this session.

   ```powershell
   Start-Process (Get-Command pwsh.exe).Source -Verb RunAs -Wait -ArgumentList @('-NoProfile','-ExecutionPolicy','Bypass','-File',<script>)
   ```

13. Run `pes-gameplay-status` again. Confirm the intended hashes and Sider entries.

14. Ask the user to launch the game. Confirm that the save loads without a create-edit-data prompt.

## Review

- Use the Documents save path as the live target.
- Apply patch parts in documented order.
- Use existing staged parts and backup trees as rollback sources.
- Stop when a write has no rollback source or the game requests new edit data.
- Follow the repository cache-safety rule.

## Boundaries

- Use `pes-gameplay-switch` for a gameplay-only switch.
- Use `pes-faces-install` for face folders.
- Use `pes-roster-extract` for roster extraction.
- Use `pes-evoweb-research` to assess a release.
- Use `pes-wiki-log` to record a verdict.
